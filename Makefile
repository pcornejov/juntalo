.PHONY: dev down check build migrate seed

dev:
	docker compose -f docker/docker-compose.yml up --build

down:
	docker compose -f docker/docker-compose.yml down

check:
	cd backend && $(MAKE) check
	cd frontend && npm run check

build:
	cd backend && go build ./...
	cd frontend && npm run build

migrate:
	cd backend && $(MAKE) migrate-up

seed:
	cd backend && $(MAKE) seed
