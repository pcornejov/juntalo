-- name: ListAdminUsers :many
-- Backoffice del operador de la plataforma: cada usuario, su organización
-- personal y cuántas campañas tiene. No filtra por organización — a
-- diferencia de todo el resto del código, este query es intencionalmente
-- cross-tenant porque solo lo puede llamar un admin de plataforma (ver
-- middleware.RequireAdminUser).
SELECT
  u.id, u.email, u.full_name, u.email_verified_at, u.created_at,
  o.id AS organization_id, o.name AS organization_name,
  COUNT(c.id) AS campaign_count
FROM users u
JOIN organization_members om ON om.user_id = u.id AND om.role = 'owner'
JOIN organizations o ON o.id = om.organization_id
LEFT JOIN campaigns c ON c.organization_id = o.id AND c.deleted_at IS NULL
GROUP BY u.id, o.id
ORDER BY u.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAdminCampaigns :many
-- Todas las campañas de la plataforma, no solo las de una organización.
SELECT
  c.id, c.title, c.slug, c.type_key, c.status, c.category, c.created_at,
  o.name AS organization_name, u.email AS organizer_email,
  COALESCE(ct.raised_gross, 0)::bigint AS raised_gross,
  COALESCE(ct.contributor_count, 0)::bigint AS contributor_count
FROM campaigns c
JOIN organizations o ON o.id = c.organization_id
JOIN organization_members om ON om.organization_id = o.id AND om.role = 'owner'
JOIN users u ON u.id = om.user_id
LEFT JOIN campaign_totals ct ON ct.campaign_id = c.id
WHERE c.deleted_at IS NULL
ORDER BY c.created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListAdminPayments :many
-- Cada pago de la plataforma, para auditar transacciones sin entrar
-- campaña por campaña.
SELECT
  p.id, p.status, p.provider, p.amount_gross, p.amount_net, p.commission_amount,
  p.created_at, p.confirmed_at,
  c.title AS campaign_title, c.slug AS campaign_slug,
  ctr.full_name AS contributor_name
FROM payments p
JOIN contributions co ON co.id = p.contribution_id
JOIN campaigns c ON c.id = co.campaign_id
JOIN contributors ctr ON ctr.id = co.contributor_id
ORDER BY p.created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetAdminMetrics :one
-- Métricas globales para el dashboard del backoffice — raised_gross/
-- raised_net_approx se agregan sobre campaign_totals (que ya descuenta
-- reembolsos), no sobre payments directo, para no duplicar esa lógica.
SELECT
  (SELECT COUNT(*) FROM users)::bigint AS total_users,
  (SELECT COUNT(*) FROM campaigns WHERE deleted_at IS NULL)::bigint AS total_campaigns,
  (SELECT COUNT(*) FROM campaigns WHERE status = 'active' AND deleted_at IS NULL)::bigint AS active_campaigns,
  (SELECT COUNT(*) FROM campaigns WHERE status = 'draft' AND deleted_at IS NULL)::bigint AS draft_campaigns,
  (SELECT COUNT(*) FROM campaigns WHERE status = 'finished' AND deleted_at IS NULL)::bigint AS finished_campaigns,
  (SELECT COUNT(*) FROM contributions WHERE status = 'confirmed')::bigint AS total_contributions,
  (SELECT COALESCE(SUM(raised_gross), 0)::bigint FROM campaign_totals) AS raised_gross,
  (SELECT COALESCE(SUM(raised_net_approx), 0)::bigint FROM campaign_totals) AS raised_net_approx,
  (SELECT COALESCE(SUM(commission_amount), 0)::bigint FROM payments WHERE status IN ('confirmed', 'partially_refunded')) AS total_commission;
