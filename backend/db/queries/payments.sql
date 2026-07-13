-- name: CreatePayment :one
INSERT INTO payments (
  contribution_id, idempotency_key, provider, provider_ref, status,
  amount_gross, commission_rate_applied, commission_amount, amount_net, payee_snapshot
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetPaymentByIdempotencyKey :one
SELECT * FROM payments WHERE idempotency_key = $1;

-- name: GetPaymentByContributionID :one
SELECT * FROM payments WHERE contribution_id = $1;

-- name: GetPaymentByProviderRefForUpdate :one
SELECT * FROM payments WHERE provider = $1 AND provider_ref = $2 FOR UPDATE;

-- name: UpdatePaymentStatus :one
UPDATE payments SET status = $2, confirmed_at = $3, failed_at = $4, updated_at = now()
WHERE id = $1
RETURNING *;
