-- Datos de desarrollo: usuario demo con una campaña de prueba.
-- password_hash corresponde a "password123" (argon2id) — generado en Hito 1;
-- por ahora este seed solo crea la estructura de organización para probar queries.

INSERT INTO users (id, email, full_name)
VALUES ('00000000-0000-0000-0000-000000000001', 'demo@juntalo.cl', 'Usuario Demo')
ON CONFLICT (email) DO NOTHING;

INSERT INTO organizations (id, name, kind, commission_rate)
VALUES ('00000000-0000-0000-0000-000000000002', 'Organización de Usuario Demo', 'personal', 0.0500)
ON CONFLICT (id) DO NOTHING;

INSERT INTO organization_members (organization_id, user_id, role)
VALUES ('00000000-0000-0000-0000-000000000002', '00000000-0000-0000-0000-000000000001', 'owner')
ON CONFLICT DO NOTHING;
