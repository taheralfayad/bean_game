ALTER TABLE game_variants
ADD CONSTRAINT game_variants_player_range_no_overlap
EXCLUDE USING gist (
	(int4range(min_players, max_players, '[]')) WITH &&
);
