-- Liquidaciones manuales a organizadores (Etapa post-MVP: "quién junta el
-- dinero y cuándo transfiere"). El operador de la plataforma sigue
-- transfiriendo a mano desde su banco — esta tabla solo deja constancia de
-- que ya se hizo, para calcular cuánto queda pendiente sin volver a sumar
-- lo ya pagado.
CREATE TABLE payouts (
    id              uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id uuid NOT NULL REFERENCES organizations(id),
    amount          bigint NOT NULL CHECK (amount > 0),
    note            text,
    created_by      uuid NOT NULL REFERENCES users(id),
    created_at      timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX idx_payouts_organization_id ON payouts (organization_id);
