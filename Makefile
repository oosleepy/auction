up:
	docker compose --env-file .env.docker up --build -d

down:
	docker compose --env-file .env.docker down

logs:
	docker compose --env-file .env.docker logs -f

migrate:
	docker compose --env-file .env.docker run --rm migrate

reset:
	docker compose --env-file .env.docker down -v
	docker compose --env-file .env.docker up --build -d
	docker compose --env-file .env.docker run --rm migrate
