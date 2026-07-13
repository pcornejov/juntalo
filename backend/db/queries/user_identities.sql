-- name: CreateUserIdentity :one
INSERT INTO user_identities (user_id, provider, password_hash)
VALUES ($1, 'password', $2)
RETURNING *;

-- name: GetPasswordIdentityByUserID :one
SELECT * FROM user_identities WHERE user_id = $1 AND provider = 'password';
