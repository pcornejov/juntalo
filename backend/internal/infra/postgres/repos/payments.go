package repos

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/pcornejov/juntalo/backend/internal/app"
	"github.com/pcornejov/juntalo/backend/internal/domain/apperr"
	"github.com/pcornejov/juntalo/backend/internal/domain/contribution"
	"github.com/pcornejov/juntalo/backend/internal/domain/money"
	"github.com/pcornejov/juntalo/backend/internal/domain/payment"
	"github.com/pcornejov/juntalo/backend/internal/infra/postgres/sqlc"
)

var errDuplicateContribution = apperr.New("duplicate_contribution", "Este aporte ya fue registrado")

type PaymentRepo struct {
	pool *pgxpool.Pool
	q    *sqlc.Queries
}

func NewPaymentRepo(pool *pgxpool.Pool) *PaymentRepo {
	return &PaymentRepo{pool: pool, q: sqlc.New(pool)}
}

func (r *PaymentRepo) Create(ctx context.Context, in app.CreatePaymentInput) (payment.Payment, error) {
	snapshot, err := json.Marshal(in.PayeeSnapshot)
	if err != nil {
		return payment.Payment{}, fmt.Errorf("marshal payee snapshot: %w", err)
	}

	p, err := r.q.CreatePayment(ctx, sqlc.CreatePaymentParams{
		ContributionID:        in.ContributionID,
		IdempotencyKey:        in.IdempotencyKey,
		Provider:              in.Provider,
		ProviderRef:           pgtype.Text{String: in.ProviderRef, Valid: in.ProviderRef != ""},
		Status:                string(in.Status),
		AmountGross:           int64(in.AmountGross),
		CommissionRateApplied: toNumeric(in.CommissionRateApplied),
		CommissionAmount:      int64(in.CommissionAmount),
		AmountNet:             int64(in.AmountNet),
		PayeeSnapshot:         snapshot,
	})
	if err != nil {
		if isUniqueViolation(err) {
			return payment.Payment{}, errDuplicateContribution
		}
		return payment.Payment{}, fmt.Errorf("create payment: %w", err)
	}
	return mapPayment(p), nil
}

func (r *PaymentRepo) GetByIdempotencyKey(ctx context.Context, key string) (payment.Payment, bool, error) {
	p, err := r.q.GetPaymentByIdempotencyKey(ctx, key)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return payment.Payment{}, false, nil
		}
		return payment.Payment{}, false, fmt.Errorf("get payment by idempotency key: %w", err)
	}
	return mapPayment(p), true, nil
}

func (r *PaymentRepo) GetByContributionID(ctx context.Context, contributionID uuid.UUID) (payment.Payment, bool, error) {
	p, err := r.q.GetPaymentByContributionID(ctx, contributionID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return payment.Payment{}, false, nil
		}
		return payment.Payment{}, false, fmt.Errorf("get payment by contribution: %w", err)
	}
	return mapPayment(p), true, nil
}

// ConfirmByProviderRef is the ★ transactional core of Hito 3: locks the
// payment row, validates the transition in domain, updates payment y su
// contribution vinculada en una sola transacción (Etapa 3 §5, Etapa 4 §5).
// Idempotente: un evento repetido para un pago ya en newStatus es un no-op.
func (r *PaymentRepo) ConfirmByProviderRef(ctx context.Context, provider, providerRef string, newStatus payment.Status) (payment.Payment, bool, error) {
	var result payment.Payment
	var transitioned bool

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		q := sqlc.New(tx)

		row, err := q.GetPaymentByProviderRefForUpdate(ctx, sqlc.GetPaymentByProviderRefForUpdateParams{
			Provider:    provider,
			ProviderRef: pgtype.Text{String: providerRef, Valid: true},
		})
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.New("payment_not_found", "Pago no encontrado")
			}
			return fmt.Errorf("lock payment: %w", err)
		}

		current := payment.Status(row.Status)
		if current == newStatus {
			result = mapPayment(row)
			return nil // idempotente: mismo evento reprocesado, sin efecto
		}
		if !payment.CanTransition(current, newStatus) {
			return apperr.New("invalid_payment_transition", "Transición de pago no permitida")
		}

		now := time.Now()
		params := sqlc.UpdatePaymentStatusParams{ID: row.ID, Status: string(newStatus)}
		if newStatus == payment.StatusConfirmed {
			params.ConfirmedAt = pgtype.Timestamptz{Time: now, Valid: true}
		}
		if newStatus == payment.StatusFailed {
			params.FailedAt = pgtype.Timestamptz{Time: now, Valid: true}
		}
		updated, err := q.UpdatePaymentStatus(ctx, params)
		if err != nil {
			return fmt.Errorf("update payment status: %w", err)
		}

		contributionStatus := mapPaymentStatusToContributionStatus(newStatus)
		if err := q.UpdateContributionStatus(ctx, sqlc.UpdateContributionStatusParams{
			ID:     row.ContributionID,
			Status: string(contributionStatus),
		}); err != nil {
			return fmt.Errorf("update contribution status: %w", err)
		}

		result = mapPayment(updated)
		transitioned = true
		return nil
	})
	return result, transitioned, err
}

var (
	errRefundExceedsBalance = apperr.New("refund_exceeds_balance", "El monto a reembolsar supera lo que queda del aporte")
	errPaymentNotRefundable = apperr.New("payment_not_refundable", "Este pago no está en un estado que se pueda reembolsar")
)

// Refund is transactional: locks the payment row, valida que el monto no
// supere el saldo pendiente (bruto - ya reembolsado), inserta el registro en
// payment_refunds, y transiciona el pago (y su contribution) a refunded o
// partially_refunded según si el reembolso cubre el saldo completo.
func (r *PaymentRepo) Refund(ctx context.Context, paymentID uuid.UUID, amount money.CLP, providerRef, reason string) (payment.Payment, error) {
	var result payment.Payment

	err := pgx.BeginFunc(ctx, r.pool, func(tx pgx.Tx) error {
		q := sqlc.New(tx)

		row, err := q.GetPaymentByIDForUpdate(ctx, paymentID)
		if err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return apperr.New("payment_not_found", "Pago no encontrado")
			}
			return fmt.Errorf("lock payment: %w", err)
		}

		current := payment.Status(row.Status)
		if current != payment.StatusConfirmed && current != payment.StatusPartiallyRefunded {
			return errPaymentNotRefundable
		}

		alreadyRefunded, err := q.GetRefundedAmountByPaymentID(ctx, paymentID)
		if err != nil {
			return fmt.Errorf("get refunded amount: %w", err)
		}
		remaining := row.AmountGross - alreadyRefunded
		if int64(amount) > remaining {
			return errRefundExceedsBalance
		}

		if _, err := q.CreatePaymentRefund(ctx, sqlc.CreatePaymentRefundParams{
			PaymentID:   paymentID,
			Amount:      int64(amount),
			ProviderRef: pgtype.Text{String: providerRef, Valid: providerRef != ""},
			Reason:      pgtype.Text{String: reason, Valid: reason != ""},
		}); err != nil {
			return fmt.Errorf("create payment refund: %w", err)
		}

		newStatus := payment.StatusPartiallyRefunded
		if int64(amount) == remaining {
			newStatus = payment.StatusRefunded
		}
		if !payment.CanTransition(current, newStatus) {
			return apperr.New("invalid_payment_transition", "Transición de pago no permitida")
		}

		updated, err := q.UpdatePaymentStatusOnly(ctx, sqlc.UpdatePaymentStatusOnlyParams{ID: paymentID, Status: string(newStatus)})
		if err != nil {
			return fmt.Errorf("update payment status: %w", err)
		}

		contributionStatus := mapPaymentStatusToContributionStatus(newStatus)
		if err := q.UpdateContributionStatus(ctx, sqlc.UpdateContributionStatusParams{
			ID:     row.ContributionID,
			Status: string(contributionStatus),
		}); err != nil {
			return fmt.Errorf("update contribution status: %w", err)
		}

		result = mapPayment(updated)
		return nil
	})
	return result, err
}

func mapPaymentStatusToContributionStatus(s payment.Status) contribution.Status {
	switch s {
	case payment.StatusConfirmed:
		return contribution.StatusConfirmed
	case payment.StatusFailed:
		return contribution.StatusFailed
	case payment.StatusRefunded, payment.StatusPartiallyRefunded:
		return contribution.StatusRefunded
	default:
		return contribution.StatusPending
	}
}
