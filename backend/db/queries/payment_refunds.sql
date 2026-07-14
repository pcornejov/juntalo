-- name: CreatePaymentRefund :one
INSERT INTO payment_refunds (payment_id, amount, provider_ref, reason)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetRefundedAmountByPaymentID :one
SELECT COALESCE(SUM(amount), 0)::bigint FROM payment_refunds WHERE payment_id = $1;
