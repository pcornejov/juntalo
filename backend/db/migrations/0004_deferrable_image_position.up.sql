-- Reordenar fotos implica reescribir varias posiciones en la misma
-- transacción; con el UNIQUE constraint chequeado por statement (default de
-- Postgres), mover la foto A a la posición de B chocaría contra B antes de
-- que B se actualice. DEFERRABLE INITIALLY DEFERRED lo valida recién al
-- COMMIT, que es cuando el estado final ya es consistente.
ALTER TABLE campaign_images DROP CONSTRAINT campaign_images_campaign_id_position_key;
ALTER TABLE campaign_images
    ADD CONSTRAINT campaign_images_campaign_id_position_key
    UNIQUE (campaign_id, position) DEFERRABLE INITIALLY DEFERRED;
