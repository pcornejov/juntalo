-- name: CreateFile :one
INSERT INTO files (organization_id, kind, storage_key, mime_type, size_bytes)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetFileByID :one
SELECT * FROM files WHERE id = $1;
