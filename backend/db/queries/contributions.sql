-- name: CreateContribution :one
INSERT INTO contributions (campaign_id, contributor_id, amount, is_anonymous, message)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetContributionByID :one
SELECT * FROM contributions WHERE id = $1;

-- name: UpdateContributionStatus :exec
UPDATE contributions SET status = $2, updated_at = now() WHERE id = $1;

-- name: ListContributionsByCampaign :many
SELECT * FROM contributions
WHERE campaign_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListParticipantsByCampaign :many
SELECT
  c.id AS contribution_id,
  ct.full_name AS contributor_full_name,
  ct.email AS contributor_email,
  ct.phone AS contributor_phone,
  c.amount,
  COALESCE(r.refunded, 0)::bigint AS refunded_amount,
  c.is_anonymous,
  c.status,
  c.created_at,
  c.message
FROM contributions c
JOIN contributors ct ON ct.id = c.contributor_id
LEFT JOIN payments p ON p.contribution_id = c.id
LEFT JOIN LATERAL (
  SELECT SUM(pr.amount) AS refunded FROM payment_refunds pr WHERE pr.payment_id = p.id
) r ON true
WHERE c.campaign_id = $1
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListParticipantsByCampaignFiltered :many
-- Misma query que ListParticipantsByCampaign, con búsqueda (nombre/email/
-- teléfono, insensible a mayúsculas) y filtro por estado — usada por el
-- buscador del dashboard (Etapa 4). $4/$5 en NULL desactivan cada filtro.
SELECT
  c.id AS contribution_id,
  ct.full_name AS contributor_full_name,
  ct.email AS contributor_email,
  ct.phone AS contributor_phone,
  c.amount,
  COALESCE(r.refunded, 0)::bigint AS refunded_amount,
  c.is_anonymous,
  c.status,
  c.created_at,
  c.message
FROM contributions c
JOIN contributors ct ON ct.id = c.contributor_id
LEFT JOIN payments p ON p.contribution_id = c.id
LEFT JOIN LATERAL (
  SELECT SUM(pr.amount) AS refunded FROM payment_refunds pr WHERE pr.payment_id = p.id
) r ON true
WHERE c.campaign_id = $1
  AND (
    $4::text IS NULL
    OR ct.full_name ILIKE '%' || $4 || '%'
    OR ct.email ILIKE '%' || $4 || '%'
    OR ct.phone ILIKE '%' || $4 || '%'
  )
  AND ($5::text IS NULL OR c.status = $5)
ORDER BY c.created_at DESC
LIMIT $2 OFFSET $3;
