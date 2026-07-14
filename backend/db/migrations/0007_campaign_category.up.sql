-- Categoría de descubrimiento (Emergencia, Salud, Evento, etc.) — a
-- diferencia de type_key, que rige reglas de negocio y hoy solo tiene
-- 'collection' habilitado, category es puramente descriptiva para filtrar
-- en la sección pública "Explorar campañas" (inspirado en Vaki).
ALTER TABLE campaigns ADD COLUMN category text NOT NULL DEFAULT 'otro'
  CHECK (category IN ('emergencia', 'salud', 'educacion', 'evento', 'comunidad', 'mascotas', 'otro'));

CREATE INDEX idx_campaigns_category ON campaigns (category) WHERE status = 'active';
