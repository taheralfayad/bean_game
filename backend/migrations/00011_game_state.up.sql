ALTER TABLE games
ADD COLUMN state VARCHAR(10) NOT NULL DEFAULT 'lobby'
CHECK (state IN ('lobby', 'starting', 'started'));

UPDATE games SET state = 'starting' WHERE locked = true AND variant_id IS NULL;
UPDATE games SET state = 'started' WHERE variant_id IS NOT NULL;

ALTER TABLE games DROP COLUMN locked;

ALTER TABLE games
ADD CONSTRAINT games_started_has_variant
CHECK ((state = 'started') = (variant_id IS NOT NULL));
