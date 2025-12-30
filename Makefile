.PHONY: up migrate ingest api down

up:
	docker compose up -d db

migrate:
	docker compose run --rm migrate

ingest:
	docker compose run --rm ingest

api:
	docker compose up api

all:
	docker compose up --build

down:
	docker compose down -v