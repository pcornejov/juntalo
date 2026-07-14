-- Galería de imágenes por campaña (antes: una sola cover_file_id). Se
-- mantiene cover_file_id para no romper datos existentes, pero las
-- campañas nuevas usan campaign_images como fuente de verdad; el "cover"
-- que se muestra en OG tags y como portada es la primera imagen de la
-- galería, con fallback a cover_file_id si la campaña no tiene galería.
ALTER TABLE files DROP CONSTRAINT files_kind_check;
ALTER TABLE files ADD CONSTRAINT files_kind_check CHECK (kind IN ('campaign_cover', 'campaign_gallery'));

CREATE TABLE campaign_images (
    id           uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id  uuid NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    file_id      uuid NOT NULL REFERENCES files(id) ON DELETE CASCADE,
    position     integer NOT NULL,
    created_at   timestamptz NOT NULL DEFAULT now(),
    UNIQUE (campaign_id, position)
);
CREATE INDEX idx_campaign_images_campaign ON campaign_images (campaign_id, position);

-- Backfill: las campañas que ya tenían cover_file_id pasan a tener esa
-- imagen como primera (única) foto de su galería.
INSERT INTO campaign_images (campaign_id, file_id, position)
SELECT id, cover_file_id, 0 FROM campaigns WHERE cover_file_id IS NOT NULL;
