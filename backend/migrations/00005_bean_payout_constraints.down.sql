ALTER TABLE beans_payout
	DROP CONSTRAINT beans_payout_coins_positive,
	DROP CONSTRAINT beans_payout_num_cards_positive,
	DROP CONSTRAINT beans_payout_pkey,
	DROP CONSTRAINT beans_payout_bean_id_fkey,
	ADD CONSTRAINT beans_payout_bean_id_fkey
		FOREIGN KEY (bean_id) REFERENCES beans(id);
