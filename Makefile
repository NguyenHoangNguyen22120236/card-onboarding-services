GO ?= go
OAPI_CODEGEN ?= oapi-codegen
DOCKER ?= docker
AWS ?= aws
BIN_DIR ?= bin
IMAGE_PREFIX ?= card-onboarding
AWS_REGION ?= ap-southeast-1
AWS_ACCOUNT_ID ?=
ECR_REGISTRY ?= $(AWS_ACCOUNT_ID).dkr.ecr.$(AWS_REGION).amazonaws.com
IMAGE_TAG ?= $(shell git rev-parse --short HEAD)
ENVIRONMENT_NAME ?= prod
STACK_NAME ?= card-onboarding-services-$(ENVIRONMENT_NAME)
VPC_ID ?=
PUBLIC_SUBNET_IDS ?=

SERVICES := account-management-service customer-management-service onboard-service
SWAGGER_FILES := $(addsuffix /swagger-internal.yaml,$(SERVICES))
GENERATED_DIRS := account-management-service/pkg customer-management-service/pkg onboard-service/pkg
ACCOUNT_MANAGEMENT_IMAGE_URI := $(ECR_REGISTRY)/account-management-service:$(IMAGE_TAG)
CUSTOMER_MANAGEMENT_IMAGE_URI := $(ECR_REGISTRY)/customer-management-service:$(IMAGE_TAG)
ONBOARD_SERVICE_IMAGE_URI := $(ECR_REGISTRY)/onboard-service:$(IMAGE_TAG)

.PHONY: help lint generate swagger-validate generate-check test coverage build docker-build ensure-ecr-repositories ecr-login docker-tag docker-push deploy-infra deploy-production clean

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
	@echo "  docker-push       - Tag and push Docker images to ECR"
	@echo "  deploy-infra      - Deploy the CloudFormation stack"
	@echo "  deploy-production - Build, push, and deploy production"

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

ensure-ecr-repositories:
	@test -n "$(AWS_ACCOUNT_ID)" || (echo "AWS_ACCOUNT_ID is required" && exit 1)
	$(AWS) ecr describe-repositories --region $(AWS_REGION) --repository-names account-management-service >/dev/null 2>&1 || $(AWS) ecr create-repository --region $(AWS_REGION) --repository-name account-management-service
	$(AWS) ecr describe-repositories --region $(AWS_REGION) --repository-names customer-management-service >/dev/null 2>&1 || $(AWS) ecr create-repository --region $(AWS_REGION) --repository-name customer-management-service
	$(AWS) ecr describe-repositories --region $(AWS_REGION) --repository-names onboard-service >/dev/null 2>&1 || $(AWS) ecr create-repository --region $(AWS_REGION) --repository-name onboard-service

ecr-login:
	@test -n "$(AWS_ACCOUNT_ID)" || (echo "AWS_ACCOUNT_ID is required" && exit 1)
	$(AWS) ecr get-login-password --region $(AWS_REGION) | $(DOCKER) login --username AWS --password-stdin $(ECR_REGISTRY)

docker-tag:
	@test -n "$(AWS_ACCOUNT_ID)" || (echo "AWS_ACCOUNT_ID is required" && exit 1)
	$(DOCKER) tag $(IMAGE_PREFIX)/account-management-service:latest $(ACCOUNT_MANAGEMENT_IMAGE_URI)
	$(DOCKER) tag $(IMAGE_PREFIX)/customer-management-service:latest $(CUSTOMER_MANAGEMENT_IMAGE_URI)
	$(DOCKER) tag $(IMAGE_PREFIX)/onboard-service:latest $(ONBOARD_SERVICE_IMAGE_URI)

docker-push: ensure-ecr-repositories ecr-login docker-tag
	$(DOCKER) push $(ACCOUNT_MANAGEMENT_IMAGE_URI)
	$(DOCKER) push $(CUSTOMER_MANAGEMENT_IMAGE_URI)
	$(DOCKER) push $(ONBOARD_SERVICE_IMAGE_URI)

deploy-infra:
	@test -n "$(VPC_ID)" || (echo "VPC_ID is required" && exit 1)
	@test -n "$(PUBLIC_SUBNET_IDS)" || (echo "PUBLIC_SUBNET_IDS is required" && exit 1)
	$(AWS) cloudformation deploy \
		--region $(AWS_REGION) \
		--template-file infra/cloudformation.yaml \
		--stack-name $(STACK_NAME) \
		--capabilities CAPABILITY_NAMED_IAM \
		--parameter-overrides \
			EnvironmentName=$(ENVIRONMENT_NAME) \
			VpcId=$(VPC_ID) \
			PublicSubnetIds=$(PUBLIC_SUBNET_IDS) \
			OnboardServiceImageUri=$(ONBOARD_SERVICE_IMAGE_URI) \
			CustomerManagementServiceImageUri=$(CUSTOMER_MANAGEMENT_IMAGE_URI) \
			AccountManagementServiceImageUri=$(ACCOUNT_MANAGEMENT_IMAGE_URI)

deploy-production:
	$(MAKE) docker-build
	$(MAKE) docker-push
	$(MAKE) deploy-infra

clean:
	$(GO) clean
	rm -rf $(BIN_DIR) coverage.out
