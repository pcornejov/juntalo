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
