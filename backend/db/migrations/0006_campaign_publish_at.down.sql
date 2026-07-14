DROP INDEX IF EXISTS idx_campaigns_due_for_publish;
ALTER TABLE campaigns DROP COLUMN publish_at;
