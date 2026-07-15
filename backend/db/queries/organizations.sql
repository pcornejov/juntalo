-- name: CreateOrganization :one
INSERT INTO organizations (name, kind, slug)
VALUES ($1, 'personal', $2)
RETURNING *;

-- name: GetPersonalOrganizationByUserID :one
SELECT o.* FROM organizations o
JOIN organization_members om ON om.organization_id = o.id
WHERE om.user_id = $1 AND o.kind = 'personal'
LIMIT 1;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;

-- name: GetOrganizationBySlug :one
SELECT * FROM organizations WHERE slug = $1;

-- name: GetOrganizationOwnerByOrgID :one
SELECT u.email, u.full_name, u.email_verified_at FROM users u
JOIN organization_members om ON om.user_id = u.id
WHERE om.organization_id = $1 AND om.role = 'owner'
LIMIT 1;

-- name: UpdateOrganizationCommissionRate :one
-- Herramienta del backoffice: ajustar la comisión de una organización sin
-- tocar código. El default global sigue siendo el 5% de la migración
-- inicial — esto permite excepciones puntuales por organización.
UPDATE organizations SET commission_rate = $2, updated_at = now()
WHERE id = $1
RETURNING *;

-- name: UpdateOrganizationPayoutInfo :one
-- Datos de transferencia que el organizador carga desde su panel — sin
-- esto, la liquidación manual del backoffice no tiene a dónde transferir.
UPDATE organizations SET
  rut = $2,
  payout_bank = $3,
  payout_account_type = $4,
  payout_account_number = $5,
  payout_holder_name = $6,
  updated_at = now()
WHERE id = $1
RETURNING *;
