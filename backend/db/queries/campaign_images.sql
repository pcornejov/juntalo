-- name: NextCampaignImagePosition :one
SELECT COALESCE(MAX(position) + 1, 0)::int FROM campaign_images WHERE campaign_id = $1;

-- name: CreateCampaignImage :one
INSERT INTO campaign_images (campaign_id, file_id, position)
VALUES ($1, $2, $3)
RETURNING *;

-- name: ListCampaignImages :many
SELECT ci.id, ci.position, f.storage_key
FROM campaign_images ci
JOIN files f ON f.id = ci.file_id
WHERE ci.campaign_id = $1
ORDER BY ci.position ASC;

-- name: DeleteCampaignImage :execrows
DELETE FROM campaign_images WHERE id = $1 AND campaign_id = $2;
