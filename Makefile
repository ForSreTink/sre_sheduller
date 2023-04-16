mod-download:
	@GO111MODULE=on go mod download

run:
	@GO111MODULE=on go run cmd/scheduler/main.go

generate-api:
	@oapi-codegen --config=./openapi-gen-conf.yaml openapi/api.yaml > internal/api/app/api.gen.go

run-compose:
	@docker-compose down
	docker-compose build
	docker-compose up