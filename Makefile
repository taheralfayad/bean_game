start:
	docker compose up

check-frontend:
	docker compose run --build --rm frontend npm run check

build-frontend:
	docker compose run --build --rm frontend npm run build

migrate:
	docker compose run migrate up

enter-db:
	docker compose exec db psql -U bean -d bean
