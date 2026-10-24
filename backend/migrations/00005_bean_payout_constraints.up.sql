ALTER TABLE beans_payout
	DROP CONSTRAINT beans_payout_bean_id_fkey,
	ADD CONSTRAINT beans_payout_bean_id_fkey
		FOREIGN KEY (bean_id) REFERENCES beans(id) ON DELETE CASCADE,
	ADD CONSTRAINT beans_payout_pkey PRIMARY KEY (bean_id, num_cards),
	ADD CONSTRAINT beans_payout_num_cards_positive CHECK (num_cards > 0),
	ADD CONSTRAINT beans_payout_coins_positive CHECK (coins > 0);
