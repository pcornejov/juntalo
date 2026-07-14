-- Auto-publicar campaña en fecha programada: un draft con publish_at en el
-- pasado lo levanta el scheduler en background (ver internal/app/campaigns
-- y cmd/api) sin intervención manual del organizador.
ALTER TABLE campaigns ADD COLUMN publish_at timestamptz;
CREATE INDEX idx_campaigns_due_for_publish ON campaigns (publish_at)
    WHERE status = 'draft' AND publish_at IS NOT NULL;
