run-dev:
	nodemon --signal SIGTERM

run-production:
	go run ./cmd/api/.

start-postgres:
	docker run -d --rm --name postgres -e POSTGRES_USER=root -p 5432:5432 -e POSTGRES_PASSWORD=${DATABASE_PASSWORD}  -e POSTGRES_DB=${DATABASE_DBNAME} -e PGDATA=/var/lib/postgresql/data/pgdata -v "caroption_go":/var/lib/postgresql/data postgres:17.5

stop-postgres:
	docker stop postgres

migrate-up:
	migrate -path ./internal/db/migrations -database postgres://root:${DATABASE_PASSWORD}@localhost:5432/caroption_go?sslmode=disable up

migrate-down:
	migrate -path ./internal/db/migrations -database postgres://root:${DATABASE_PASSWORD}@localhost:5432/caroption_go?sslmode=disable down

migrate-fix:
	migrate -path ./internal/db/migrations -database postgres://root:${DATABASE_PASSWORD}@localhost:5432/caroption_go?sslmode=disable force VERSION

test:
	go test **/**
