DELETE FROM beans_decks
WHERE deck_id = (SELECT id FROM decks WHERE slug = '35playerdeck');

DELETE FROM beans
WHERE name IN (
	'Blue Bean',
	'Chili Bean',
	'Stink Bean',
	'Green Bean',
	'Soy Bean',
	'Black-eyed Bean',
	'Red Bean',
	'Garden Bean',
	'Coffee Bean',
	'Wax Bean',
	'Cocoa Bean'
)
AND NOT EXISTS (SELECT 1 FROM beans_decks WHERE beans_decks.bean_id = beans.id)
AND NOT EXISTS (SELECT 1 FROM cards WHERE cards.bean_id = beans.id)
AND NOT EXISTS (SELECT 1 FROM fields WHERE fields.bean_id = beans.id);

DELETE FROM decks
WHERE slug = '35playerdeck'
AND NOT EXISTS (SELECT 1 FROM game_variants WHERE game_variants.deck_id = decks.id);
