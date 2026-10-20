migrate:
	docker compose run migrate up

enter-db:
	docker compose exec db psql -U bean -d bean
