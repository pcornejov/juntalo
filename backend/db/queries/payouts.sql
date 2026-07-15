-- name: CreatePayout :one
-- Deja constancia de una transferencia manual ya hecha por el operador —
-- no dispara ninguna transferencia real, solo la registra para que
-- ListPendingPayouts no la vuelva a contar como pendiente.
INSERT INTO payouts (organization_id, amount, note, created_by)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: ListPendingPayouts :many
-- Backoffice: cuánto le queda pendiente de liquidar a cada organización.
-- eligible_net son los pagos confirmados (menos reembolsos) con al menos
-- payout_hold_days de antigüedad — el mismo colchón documentado en /bases
-- para dejar margen a reembolsos antes de que la plata salga de la
-- plataforma. pending_amount = eligible_net - lo ya pagado (payouts).
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
  SELECT SUM(p.amount_net - COALESCE(r.refunded, 0)) AS amount
  FROM campaigns c
  JOIN contributions ct ON ct.campaign_id = c.id
  JOIN payments p ON p.contribution_id = ct.id
  LEFT JOIN LATERAL (
    SELECT SUM(pr.amount) AS refunded FROM payment_refunds pr WHERE pr.payment_id = p.id
  ) r ON true
  WHERE c.organization_id = o.id
    AND p.status IN ('confirmed', 'partially_refunded')
    AND p.confirmed_at <= now() - make_interval(days => $1::int)
) eligible ON true
LEFT JOIN LATERAL (
  SELECT SUM(amount) AS amount FROM payouts WHERE organization_id = o.id
) paid ON true
WHERE COALESCE(eligible.amount, 0) - COALESCE(paid.amount, 0) > 0
ORDER BY o.name
LIMIT $2 OFFSET $3;

-- name: ListPayoutsByOrg :many
SELECT * FROM payouts WHERE organization_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
