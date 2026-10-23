ALTER TABLE game_variants ALTER COLUMN deck_id DROP NOT NULL;

ALTER TABLE beans_decks DROP CONSTRAINT beans_decks_card_count_positive;
ALTER TABLE beans_decks DROP CONSTRAINT beans_decks_pkey;
ALTER TABLE beans_decks ADD COLUMN id SMALLINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY;

ALTER TABLE decks DROP CONSTRAINT decks_slug_check;
ALTER TABLE decks DROP CONSTRAINT decks_slug_key;
ALTER TABLE decks DROP COLUMN slug;
