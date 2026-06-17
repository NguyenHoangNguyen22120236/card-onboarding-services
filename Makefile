GO ?= go
OAPI_CODEGEN ?= oapi-codegen
DOCKER ?= docker
BIN_DIR ?= bin
IMAGE_PREFIX ?= card-onboarding

SERVICES := account-management-service customer-management-service onboard-service
SWAGGER_FILES := $(addsuffix /swagger-internal.yaml,$(SERVICES))
GENERATED_DIRS := account-management-service/pkg customer-management-service/pkg onboard-service/pkg

.PHONY: help lint generate swagger-validate generate-check test coverage build docker-build clean

help:
	@echo "Available targets:"
	@echo "  lint              - Run Go formatting check and vet"
	@echo "  generate          - Regenerate OpenAPI server/client/types code"
	@echo "  swagger-validate  - Validate OpenAPI specs through oapi-codegen parsing"
	@echo "  generate-check    - Verify generated OpenAPI code is committed"
	@echo "  test              - Run unit tests"
	@echo "  coverage          - Run tests with coverage output"
	@echo "  build             - Build all service binaries"
	@echo "  docker-build      - Build Docker images for all services"

lint:
	@test -z "$$(gofmt -l .)"
	$(GO) vet ./...

generate:
	@for service in $(SERVICES); do \
		$(MAKE) -C $$service generate; \
	done

swagger-validate:
	@for spec in $(SWAGGER_FILES); do \
		echo "Validating $$spec"; \
		$(OAPI_CODEGEN) -generate types -package validate "$$spec" >/dev/null; \
	done

generate-check: generate
	git diff --exit-code -- $(GENERATED_DIRS)

test:
	$(GO) test ./...

coverage:
	$(GO) test ./... -coverprofile=coverage.out -covermode=atomic
	$(GO) tool cover -func=coverage.out

build:
	$(GO) build -o $(BIN_DIR)/account-management-service ./account-management-service
	$(GO) build -o $(BIN_DIR)/customer-management-service ./customer-management-service
	$(GO) build -o $(BIN_DIR)/onboard-service ./onboard-service

docker-build:
	$(DOCKER) build --build-arg SERVICE_PATH=account-management-service --build-arg SERVICE_PORT=8082 -t $(IMAGE_PREFIX)/account-management-service:latest -f Dockerfile.service .
	$(DOCKER) build --build-arg SERVICE_PATH=customer-management-service --build-arg SERVICE_PORT=8081 -t $(IMAGE_PREFIX)/customer-management-service:latest -f Dockerfile.service .
	$(DOCKER) build --build-arg SERVICE_PATH=onboard-service --build-arg SERVICE_PORT=8080 -t $(IMAGE_PREFIX)/onboard-service:latest -f Dockerfile.service .

clean:
	$(GO) clean
	rm -rf $(BIN_DIR) coverage.out
