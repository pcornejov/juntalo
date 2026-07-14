-- Página pública persistente por organizador (/org/:slug) — inspirado en el
-- link único de por vida de Ceneka: a diferencia de una campaña puntual, un
-- organizador recurrente (ONG, junta de vecinos) puede compartir un solo
-- link fijo que lista todas sus campañas pasadas y activas.
ALTER TABLE organizations ADD COLUMN slug text;

-- Backfill de filas existentes: no depende de la lógica de Slugify en Go
-- (no está disponible en SQL puro), así que el resultado es más tosco que
-- los slugs nuevos (sin quitar tildes) — igual queda único gracias al
-- sufijo del id, que es lo único que importa para una URL válida.
UPDATE organizations
SET slug = regexp_replace(lower(trim(name)), '[^a-z0-9]+', '-', 'g') || '-' || substr(replace(id::text, '-', ''), 1, 6)
WHERE slug IS NULL;

ALTER TABLE organizations ALTER COLUMN slug SET NOT NULL;
ALTER TABLE organizations ADD CONSTRAINT organizations_slug_key UNIQUE (slug);
