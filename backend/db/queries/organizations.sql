-- name: CreateOrganization :one
INSERT INTO organizations (name, kind)
VALUES ($1, 'personal')
RETURNING *;

-- name: GetPersonalOrganizationByUserID :one
SELECT o.* FROM organizations o
JOIN organization_members om ON om.organization_id = o.id
WHERE om.user_id = $1 AND o.kind = 'personal'
LIMIT 1;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE id = $1;
