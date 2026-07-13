-- Juntalo — schema inicial (Etapa 3: modelo entidad-relación)

CREATE EXTENSION IF NOT EXISTS citext;
CREATE EXTENSION IF NOT EXISTS pgcrypto; -- gen_random_uuid()

-- ═══════════════════════════════════════════════════════════════
-- Identidad y tenancy
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE users (
    id                 uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    email              citext NOT NULL UNIQUE,
    email_verified_at  timestamptz,
    full_name          text NOT NULL,
    status             text NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'blocked')),
    created_at         timestamptz NOT NULL DEFAULT now(),
    updated_at         timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE user_identities (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id           uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider          text NOT NULL CHECK (provider IN ('password', 'google')),
    password_hash     text,
    provider_subject  text,
    created_at        timestamptz NOT NULL DEFAULT now(),
    UNIQUE (user_id, provider),
    UNIQUE (provider, provider_subject)
);

CREATE TABLE refresh_tokens (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id     uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash  text NOT NULL UNIQUE,
    expires_at  timestamptz NOT NULL,
    revoked_at  timestamptz,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_refresh_tokens_user ON refresh_tokens (user_id);

CREATE TABLE organizations (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    name                      text NOT NULL,
    kind                      text NOT NULL DEFAULT 'personal' CHECK (kind IN ('personal', 'team')),
    commission_rate           numeric(5,4) NOT NULL DEFAULT 0.0500,
    rut                       text,
    payout_bank               text,
    payout_account_type       text,
    payout_account_number     text,
    payout_holder_name        text,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now()
);

CREATE TABLE organization_members (
    organization_id  uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id          uuid NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role             text NOT NULL DEFAULT 'owner' CHECK (role IN ('owner', 'admin', 'viewer')),
    created_at       timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (organization_id, user_id)
);

-- ═══════════════════════════════════════════════════════════════
-- Archivos
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE files (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    kind             text NOT NULL CHECK (kind IN ('campaign_cover')),
    storage_key      text NOT NULL,
    mime_type        text NOT NULL,
    size_bytes       bigint NOT NULL,
    created_at       timestamptz NOT NULL DEFAULT now()
);

-- ═══════════════════════════════════════════════════════════════
-- Campañas
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE campaigns (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id  uuid NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    type_key         text NOT NULL CHECK (type_key IN ('collection', 'sale', 'event', 'course', 'presale', 'raffle')),
    title            text NOT NULL CHECK (char_length(title) <= 120),
    slug             text NOT NULL UNIQUE,
    description      text NOT NULL DEFAULT '',
    cover_file_id    uuid REFERENCES files(id),
    goal_amount      bigint,
    currency         char(3) NOT NULL DEFAULT 'CLP' CHECK (currency = 'CLP'),
    status           text NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'active', 'paused', 'finished', 'suspended')),
    starts_at        timestamptz,
    ends_at          timestamptz,
    settings         jsonb NOT NULL DEFAULT '{}',
    deleted_at       timestamptz,
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_campaigns_org_status ON campaigns (organization_id, status);
CREATE INDEX idx_campaigns_active ON campaigns (status) WHERE status = 'active';

-- ═══════════════════════════════════════════════════════════════
-- Contribuyentes y aportes
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE contributors (
    id          uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    full_name   text NOT NULL,
    email       citext,
    phone       text,
    created_at  timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_contributors_email ON contributors (email);

CREATE TABLE contributions (
    id               uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    campaign_id      uuid NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    contributor_id   uuid NOT NULL REFERENCES contributors(id),
    amount           bigint NOT NULL CHECK (amount > 0),
    currency         char(3) NOT NULL DEFAULT 'CLP',
    is_anonymous     boolean NOT NULL DEFAULT false,
    message          text,
    status           text NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed', 'failed', 'refunded')),
    created_at       timestamptz NOT NULL DEFAULT now(),
    updated_at       timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_contributions_campaign_status ON contributions (campaign_id, status, created_at DESC);

-- ═══════════════════════════════════════════════════════════════
-- Pagos
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE payments (
    id                        uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    contribution_id           uuid NOT NULL UNIQUE REFERENCES contributions(id) ON DELETE CASCADE,
    idempotency_key           text NOT NULL UNIQUE,
    provider                  text NOT NULL CHECK (provider IN ('mock', 'webpay', 'mercadopago', 'stripe', 'khipu')),
    provider_ref              text,
    status                    text NOT NULL DEFAULT 'pending'
                                 CHECK (status IN ('pending', 'confirmed', 'failed', 'refunded', 'partially_refunded')),
    amount_gross              bigint NOT NULL,
    commission_rate_applied   numeric(5,4) NOT NULL,
    commission_amount         bigint NOT NULL,
    amount_net                bigint NOT NULL,
    currency                  char(3) NOT NULL DEFAULT 'CLP',
    payee_snapshot            jsonb NOT NULL DEFAULT '{}',
    confirmed_at              timestamptz,
    failed_at                 timestamptz,
    created_at                timestamptz NOT NULL DEFAULT now(),
    updated_at                timestamptz NOT NULL DEFAULT now()
);
CREATE UNIQUE INDEX idx_payments_provider_ref ON payments (provider, provider_ref) WHERE provider_ref IS NOT NULL;
CREATE INDEX idx_payments_status_created ON payments (status, created_at);

CREATE TABLE payment_refunds (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id    uuid NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    amount        bigint NOT NULL CHECK (amount > 0),
    provider_ref  text,
    reason        text,
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_payment_refunds_payment ON payment_refunds (payment_id);

-- ═══════════════════════════════════════════════════════════════
-- Auditoría (append-only)
-- ═══════════════════════════════════════════════════════════════

CREATE TABLE audit_logs (
    id                uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    actor_user_id     uuid REFERENCES users(id),
    organization_id   uuid,
    action            text NOT NULL,
    entity_type       text NOT NULL,
    entity_id         uuid NOT NULL,
    data              jsonb NOT NULL DEFAULT '{}',
    created_at        timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_audit_logs_org_created ON audit_logs (organization_id, created_at DESC);

-- ═══════════════════════════════════════════════════════════════
-- Vista: totales de campaña (Etapa 3 §5 — agregación, no contador)
-- ═══════════════════════════════════════════════════════════════

CREATE VIEW campaign_totals AS
SELECT
    c.id AS campaign_id,
    COALESCE(SUM(p.amount_gross) FILTER (WHERE p.status IN ('confirmed', 'partially_refunded')), 0)
        - COALESCE(SUM(r.total_refunded), 0) AS raised_gross,
    COALESCE(SUM(p.amount_net) FILTER (WHERE p.status IN ('confirmed', 'partially_refunded')), 0)
        - COALESCE(SUM(r.total_refunded), 0) AS raised_net_approx,
    COUNT(DISTINCT ct.id) FILTER (WHERE ct.status = 'confirmed') AS contributor_count
FROM campaigns c
LEFT JOIN contributions ct ON ct.campaign_id = c.id
LEFT JOIN payments p ON p.contribution_id = ct.id
LEFT JOIN LATERAL (
    SELECT SUM(pr.amount) AS total_refunded
    FROM payment_refunds pr
    WHERE pr.payment_id = p.id
) r ON true
GROUP BY c.id;
