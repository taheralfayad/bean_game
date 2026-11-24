INSERT INTO beans (name) VALUES
	('Blue Bean'),
	('Chili Bean'),
	('Stink Bean'),
	('Green Bean'),
	('Soy Bean'),
	('Black-eyed Bean'),
	('Red Bean'),
	('Garden Bean'),
	('Coffee Bean'),
	('Wax Bean'),
	('Cocoa Bean')
ON CONFLICT (name) DO NOTHING;

INSERT INTO beans_payout (bean_id, num_cards, coins)
SELECT beans.id, payout.num_cards, payout.coins
FROM beans
JOIN (
	VALUES
		('Blue Bean', 4, 1),
		('Blue Bean', 6, 2),
		('Blue Bean', 8, 3),
		('Blue Bean', 10, 4),
		('Chili Bean', 3, 1),
		('Chili Bean', 6, 2),
		('Chili Bean', 8, 3),
		('Chili Bean', 9, 4),
		('Stink Bean', 3, 1),
		('Stink Bean', 5, 2),
		('Stink Bean', 7, 3),
		('Stink Bean', 8, 4),
		('Green Bean', 3, 1),
		('Green Bean', 5, 2),
		('Green Bean', 6, 3),
		('Green Bean', 7, 4),
		('Soy Bean', 2, 1),
		('Soy Bean', 4, 2),
		('Soy Bean', 6, 3),
		('Soy Bean', 7, 4),
		('Black-eyed Bean', 2, 1),
		('Black-eyed Bean', 4, 2),
		('Black-eyed Bean', 5, 3),
		('Black-eyed Bean', 6, 4),
		('Red Bean', 2, 1),
		('Red Bean', 3, 2),
		('Red Bean', 4, 3),
		('Red Bean', 5, 4),
		('Garden Bean', 2, 2),
		('Garden Bean', 3, 3),
		('Coffee Bean', 4, 1),
		('Coffee Bean', 7, 2),
		('Coffee Bean', 10, 3),
		('Coffee Bean', 12, 4),
		('Wax Bean', 4, 1),
		('Wax Bean', 7, 2),
		('Wax Bean', 9, 3),
		('Wax Bean', 11, 4),
		('Cocoa Bean', 2, 2),
		('Cocoa Bean', 3, 3),
		('Cocoa Bean', 4, 4)
) AS payout (name, num_cards, coins) ON payout.name = beans.name
ON CONFLICT (bean_id, num_cards) DO UPDATE SET coins = EXCLUDED.coins;

INSERT INTO decks (slug) VALUES ('35playerdeck')
ON CONFLICT (slug) DO NOTHING;

DELETE FROM beans_decks
WHERE deck_id = (SELECT id FROM decks WHERE slug = '35playerdeck');

INSERT INTO beans_decks (deck_id, bean_id, card_count)
SELECT decks.id, beans.id, cards.card_count
FROM decks
JOIN (
	VALUES
		('Blue Bean', 20),
		('Chili Bean', 18),
		('Stink Bean', 16),
		('Green Bean', 14),
		('Soy Bean', 12),
		('Black-eyed Bean', 10),
		('Red Bean', 8),
		('Garden Bean', 6),
		('Coffee Bean', 24),
		('Wax Bean', 22),
		('Cocoa Bean', 4)
) AS cards (name, card_count) ON true
JOIN beans ON beans.name = cards.name
WHERE decks.slug = '35playerdeck';
