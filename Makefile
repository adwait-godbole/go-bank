DB_URL=postgresql://root:secret@localhost:5432/go_bank?sslmode=disable

network:
	docker network create bank-network

postgres:
	docker run --name postgres12 --network bank-network -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root go_bank

dropdb:
	docker exec -it postgres12 dropdb go_bank

new_migration:
	migrate create -ext sql -dir db/migration -seq $(NAME)

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up $(N)

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down $(N)

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

sqlc:
	sqlc generate

test:
	go clean -testcache
	go test -v -cover -short ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/adwait-godbole/go-bank/db/sqlc Store
	mockgen -package mockworker -destination worker/mock/distributor.go github.com/adwait-godbole/go-bank/worker TaskDistributor

proto:
	rm -f pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --proto_path=proto --go_out=pb --go_opt=paths=source_relative \
	--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
	--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=go_bank \
	proto/*.proto
	statik -f -src=./doc/swagger -dest=./doc

evans:
	evans --host localhost --port 9090 -r repl

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

test-redis:
	docker exec -it redis redis-cli ping

.PHONY: network postgres createdb dropdb new_migration migrateup migratedown db_docs db_schema sqlc test server mock proto evans