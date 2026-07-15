-- Juntalo es solo para aportar a causas: se elimina el reembolso self-service
-- del organizador y el colchón de payout_hold_days que existía para dejarle
-- margen antes de liquidar. Los pagos/aportes que hubiesen quedado en un
-- estado de reembolso vuelven a 'confirmed' — ese estado ya no existe.

UPDATE payments SET status = 'confirmed' WHERE status IN ('refunded', 'partially_refunded');
UPDATE contributions SET status = 'confirmed' WHERE status = 'refunded';

DROP VIEW campaign_totals;

ALTER TABLE payments DROP CONSTRAINT payments_status_check;
ALTER TABLE payments ADD CONSTRAINT payments_status_check CHECK (status IN ('pending', 'confirmed', 'failed'));

ALTER TABLE contributions DROP CONSTRAINT contributions_status_check;
ALTER TABLE contributions ADD CONSTRAINT contributions_status_check CHECK (status IN ('pending', 'confirmed', 'failed'));

DROP TABLE payment_refunds;

CREATE VIEW campaign_totals AS
SELECT
    c.id AS campaign_id,
    COALESCE(SUM(p.amount_gross) FILTER (WHERE p.status = 'confirmed'), 0) AS raised_gross,
    COALESCE(SUM(p.amount_net) FILTER (WHERE p.status = 'confirmed'), 0) AS raised_net_approx,
    COUNT(DISTINCT ct.id) FILTER (WHERE ct.status = 'confirmed') AS contributor_count
FROM campaigns c
LEFT JOIN contributions ct ON ct.campaign_id = c.id
LEFT JOIN payments p ON p.contribution_id = ct.id
GROUP BY c.id;
