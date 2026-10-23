ALTER TABLE decks ADD COLUMN slug VARCHAR(40) NOT NULL UNIQUE CHECK (slug ~ '^[a-z0-9_]+$');

ALTER TABLE beans_decks DROP CONSTRAINT beans_decks_pkey;
ALTER TABLE beans_decks DROP COLUMN id;
ALTER TABLE beans_decks ADD PRIMARY KEY (deck_id, bean_id);
ALTER TABLE beans_decks ADD CONSTRAINT beans_decks_card_count_positive CHECK (card_count > 0);

ALTER TABLE game_variants ALTER COLUMN deck_id SET NOT NULL;
