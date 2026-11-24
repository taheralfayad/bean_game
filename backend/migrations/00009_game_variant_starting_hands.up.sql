CREATE TABLE game_variant_starting_hands (
    variant_id SMALLINT NOT NULL REFERENCES game_variants(id) ON DELETE CASCADE,
    seat SMALLINT NOT NULL CHECK (seat >= 1),
    card_count SMALLINT NOT NULL CHECK (card_count > 0),
    PRIMARY KEY (variant_id, seat)
);
