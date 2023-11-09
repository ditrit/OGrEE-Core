all:
	docker compose -p mytenant --profile web --profile doc up -d

api:
	docker compose -p mytenant up -d

dev_api:
	docker compose -p mytenant -f docker-compose.yml -f docker-compose.dev.yml up -d --build



