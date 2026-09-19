.PHONY: up postgres test-api api web

up:
	docker compose up --build -d

postgres:
	docker compose up -d postgres

test-api:
	cd api && DATABASE_URL=$${DATABASE_URL:-postgres://nacao:nacao@localhost:5432/nacao?sslmode=disable} go test ./...

api:
	cd api && DATABASE_URL=$${DATABASE_URL:-postgres://nacao:nacao@localhost:5432/nacao?sslmode=disable} \
		SEED_OPERADOR_EMAIL=$${SEED_OPERADOR_EMAIL:-operador@nacao.local} \
		SEED_OPERADOR_SENHA=$${SEED_OPERADOR_SENHA:-nacao-dev} \
		go run ./cmd/server

web:
	cd web && npm run dev
