-- name: CreateContributor :one
INSERT INTO contributors (full_name, email, phone)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetContributorByID :one
SELECT * FROM contributors WHERE id = $1;
