ALTER TABLE beans_decks DROP CONSTRAINT beans_decks_deck_id_fkey;
ALTER TABLE beans_decks ADD CONSTRAINT beans_decks_deck_id_fkey
	FOREIGN KEY (deck_id) REFERENCES decks(id);
