-- name: CreateUserIdentity :one
INSERT INTO user_identities (user_id, provider, password_hash)
VALUES ($1, 'password', $2)
RETURNING *;

-- name: GetPasswordIdentityByUserID :one
SELECT * FROM user_identities WHERE user_id = $1 AND provider = 'password';

-- name: UpdatePasswordHash :exec
UPDATE user_identities SET password_hash = $2 WHERE user_id = $1 AND provider = 'password';

-- name: CreateGoogleUserIdentity :one
INSERT INTO user_identities (user_id, provider, provider_subject)
VALUES ($1, 'google', $2)
RETURNING *;

-- name: GetUserIdentityByGoogleSubject :one
SELECT * FROM user_identities WHERE provider = 'google' AND provider_subject = $1;
