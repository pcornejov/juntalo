ALTER TABLE campaign_images DROP CONSTRAINT campaign_images_campaign_id_position_key;
ALTER TABLE campaign_images
    ADD CONSTRAINT campaign_images_campaign_id_position_key
    UNIQUE (campaign_id, position);
