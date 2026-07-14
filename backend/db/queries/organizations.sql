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
