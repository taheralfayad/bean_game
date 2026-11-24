ALTER TABLE cards DROP COLUMN game_player_id;
ALTER TABLE cards ADD COLUMN player_id BIGINT REFERENCES players(id);
