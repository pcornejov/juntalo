-- video_url: link de YouTube o Vimeo (no se aloja video propio — evita el
-- costo de storage/range-requests de servir video). NULL = sin video.
ALTER TABLE campaigns ADD COLUMN video_url text;
