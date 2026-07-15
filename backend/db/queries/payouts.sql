-- name: CreatePayout :one
-- Deja constancia de una transferencia manual ya hecha por el operador —
-- no dispara ninguna transferencia real, solo la registra para que
-- ListPendingPayouts no la vuelva a contar como pendiente.
INSERT INTO payouts (organization_id, amount, note, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPendingPayouts :many
-- Backoffice: cuánto le queda pendiente de liquidar a cada organización.
-- eligible_net son los pagos confirmados. pending_amount = eligible_net -
-- lo ya pagado (payouts).
SELECT
  o.id AS organization_id,
  o.name AS organization_name,
  o.rut,
  o.payout_bank,
  o.payout_account_type,
  o.payout_account_number,
  o.payout_holder_name,
  COALESCE(eligible.amount, 0)::bigint AS eligible_net,
  COALESCE(paid.amount, 0)::bigint AS total_paid,
  (COALESCE(eligible.amount, 0) - COALESCE(paid.amount, 0))::bigint AS pending_amount
FROM organizations o
LEFT JOIN LATERAL (
  SELECT SUM(p.amount_net) AS amount
  FROM campaigns c
  JOIN contributions ct ON ct.campaign_id = c.id
  JOIN payments p ON p.contribution_id = ct.id
  WHERE c.organization_id = o.id
    AND p.status = 'confirmed'
) eligible ON true
LEFT JOIN LATERAL (
  SELECT SUM(amount) AS amount FROM payouts WHERE organization_id = o.id
) paid ON true
WHERE COALESCE(eligible.amount, 0) - COALESCE(paid.amount, 0) > 0
ORDER BY o.name
LIMIT $1 OFFSET $2;

-- name: ListPayoutsByOrg :many
SELECT * FROM payouts WHERE organization_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
