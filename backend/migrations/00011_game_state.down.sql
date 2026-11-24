ALTER TABLE games DROP CONSTRAINT games_started_has_variant;

ALTER TABLE games
ADD COLUMN locked BOOLEAN NOT NULL DEFAULT false;

UPDATE games SET locked = true WHERE state IN ('starting', 'started');

ALTER TABLE games DROP COLUMN state;
