-- name: CreateCampaign :one
INSERT INTO campaigns (organization_id, type_key, title, slug, description, goal_amount, starts_at, ends_at, publish_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
RETURNING *;

-- name: GetCampaignByID :one
SELECT * FROM campaigns WHERE id = $1 AND deleted_at IS NULL;

-- name: GetCampaignByIDForOrg :one
SELECT * FROM campaigns WHERE id = $1 AND organization_id = $2 AND deleted_at IS NULL;

-- name: GetCampaignBySlug :one
SELECT * FROM campaigns WHERE slug = $1 AND deleted_at IS NULL;

-- name: SlugExists :one
SELECT EXISTS(SELECT 1 FROM campaigns WHERE slug = $1);

-- name: ListCampaignsByOrg :many
SELECT * FROM campaigns
WHERE organization_id = $1 AND deleted_at IS NULL
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateCampaign :one
UPDATE campaigns SET
  title = $2,
  description = $3,
  goal_amount = $4,
  starts_at = $5,
  ends_at = $6,
  cover_file_id = $7,
  publish_at = $8,
  updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: UpdateCampaignStatus :one
UPDATE campaigns SET status = $2, updated_at = now()
WHERE id = $1 AND deleted_at IS NULL
RETURNING *;

-- name: SoftDeleteCampaign :exec
UPDATE campaigns SET deleted_at = now() WHERE id = $1;

-- name: GetCampaignTotals :one
SELECT * FROM campaign_totals WHERE campaign_id = $1;

-- name: PublishDueCampaigns :many
-- El scheduler en background (ver cmd/api) llama esto cada minuto: publica
-- atómicamente todo draft cuya publish_at ya venció, sin condición de
-- carrera entre el scheduler y una publicación manual del organizador
-- (el UPDATE solo afecta filas que siguen en 'draft').
UPDATE campaigns SET status = 'active', updated_at = now()
WHERE status = 'draft' AND publish_at IS NOT NULL AND publish_at <= now() AND deleted_at IS NULL
RETURNING *;
