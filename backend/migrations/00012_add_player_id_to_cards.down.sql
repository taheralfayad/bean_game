ALTER TABLE cards DROP COLUMN player_id;
ALTER TABLE cards ADD COLUMN BIGINT game_player_id
