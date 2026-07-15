DROP VIEW campaign_totals;

CREATE TABLE payment_refunds (
    id            uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id    uuid NOT NULL REFERENCES payments(id) ON DELETE CASCADE,
    amount        bigint NOT NULL CHECK (amount > 0),
    provider_ref  text,
    reason        text,
    created_at    timestamptz NOT NULL DEFAULT now()
);
CREATE INDEX idx_payment_refunds_payment ON payment_refunds (payment_id);

ALTER TABLE contributions DROP CONSTRAINT contributions_status_check;
ALTER TABLE contributions ADD CONSTRAINT contributions_status_check CHECK (status IN ('pending', 'confirmed', 'failed', 'refunded'));

ALTER TABLE payments DROP CONSTRAINT payments_status_check;
ALTER TABLE payments ADD CONSTRAINT payments_status_check CHECK (status IN ('pending', 'confirmed', 'failed', 'refunded', 'partially_refunded'));

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
