ALTER TABLE beans ADD COLUMN image_link varchar(255);
ALTER TABLE beans DROP COLUMN image_file_id;
DROP TABLE files;
