ALTER TABLE beans ADD COLUMN image_key text CHECK (image_key <> '');
ALTER TABLE beans DROP COLUMN image_link;
