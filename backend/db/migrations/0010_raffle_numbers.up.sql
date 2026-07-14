-- Rifa con venta de números (tipo "raffle", ya habilitado en el Registry):
-- precio fijo por número, rango total definido al crear la campaña, número
-- asignado automáticamente en orden de compra (sin selector visual — evita
-- el problema de dos personas viendo la misma grilla al mismo tiempo).
ALTER TABLE campaigns ADD COLUMN raffle_unit_price bigint;
ALTER TABLE campaigns ADD COLUMN raffle_total_numbers integer;
-- El organizador lo registra a mano tras el sorteo externo (ej. Kino/Loto de
-- una fecha específica) — Juntalo no sortea nada dentro de la plataforma,
-- solo deja constancia pública del resultado.
ALTER TABLE campaigns ADD COLUMN raffle_winning_number integer;

ALTER TABLE contributions ADD COLUMN raffle_number integer;

-- Único por campaña: nunca se vende el mismo número dos veces. Parcial
-- (solo cuando no es NULL) porque el resto de los tipos de campaña no usa
-- esta columna.
CREATE UNIQUE INDEX contributions_campaign_raffle_number_key
    ON contributions (campaign_id, raffle_number)
    WHERE raffle_number IS NOT NULL;
