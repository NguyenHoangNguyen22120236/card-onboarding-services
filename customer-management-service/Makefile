.PHONY: generate run test

generate:
	oapi-codegen -config oapi-codegen.yaml swagger-internal.yaml
	oapi-codegen -config oapi-codegen-server.yaml swagger-internal.yaml
	oapi-codegen -config oapi-codegen-client.yaml swagger-internal.yaml

run:
	go run main.go

test:
	go test ./...
