-- name: CreateAuditLog :exec
INSERT INTO audit_logs (actor_user_id, organization_id, action, entity_type, entity_id, data)
VALUES ($1, $2, $3, $4, $5, $6);
