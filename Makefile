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
DESCRIBE_STACK_EVENTS = $(AWS) cloudformation describe-stack-events --region $(AWS_REGION) --stack-name $(STACK_NAME) --max-items 20 --query "StackEvents[?ResourceStatusReason != null].[Timestamp,LogicalResourceId,ResourceStatus,ResourceStatusReason]" --output table
ifeq ($(OS),Windows_NT)
NULL_DEVICE ?= NUL
RM_RF = if exist "$(1)" rmdir /S /Q "$(1)"
RM_FILE = if exist "$(1)" del /Q "$(1)"
DEPLOY_INFRA_ON_FAILURE = || ($(DESCRIBE_STACK_EVENTS) & exit /B 1)
else
NULL_DEVICE ?= /dev/null
RM_RF = rm -rf "$(1)"
RM_FILE = rm -f "$(1)"
DEPLOY_INFRA_ON_FAILURE = || { status=$$?; $(DESCRIBE_STACK_EVENTS) || true; exit $$status; }
endif

require = $(if $(strip $($(1))),,$(error $(1) is required))

SERVICES := account-management-service customer-management-service onboard-service
SWAGGER_FILES := $(addsuffix /swagger-internal.yaml,$(SERVICES))
GENERATED_DIRS := account-management-service/pkg customer-management-service/pkg onboard-service/pkg
ACCOUNT_MANAGEMENT_IMAGE_URI := $(ECR_REGISTRY)/account-management-service:$(IMAGE_TAG)
CUSTOMER_MANAGEMENT_IMAGE_URI := $(ECR_REGISTRY)/customer-management-service:$(IMAGE_TAG)
ONBOARD_SERVICE_IMAGE_URI := $(ECR_REGISTRY)/onboard-service:$(IMAGE_TAG)

.PHONY: help lint generate swagger-validate generate-check test coverage build docker-build docker-build-push ensure-ecr-repositories ecr-login docker-tag docker-push describe-stack-events deploy-infra deploy-production clean

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
	@echo "  docker-push       - Build and push Docker images to ECR"
	@echo "  deploy-infra      - Deploy the CloudFormation stack"
	@echo "  deploy-production - Build, push, and deploy production"

lint:
	$(if $(strip $(shell gofmt -l .)),$(error gofmt changes are required),)
	$(GO) vet ./...

generate:
	$(MAKE) -C account-management-service generate
	$(MAKE) -C customer-management-service generate
	$(MAKE) -C onboard-service generate

swagger-validate:
	$(OAPI_CODEGEN) -generate types -package validate -o $(NULL_DEVICE) account-management-service/swagger-internal.yaml
	$(OAPI_CODEGEN) -generate types -package validate -o $(NULL_DEVICE) customer-management-service/swagger-internal.yaml
	$(OAPI_CODEGEN) -generate types -package validate -o $(NULL_DEVICE) onboard-service/swagger-internal.yaml

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

docker-build-push:
	$(call require,AWS_ACCOUNT_ID)
	$(DOCKER) buildx build --push --build-arg SERVICE_PATH=account-management-service --build-arg SERVICE_PORT=8082 -t $(ACCOUNT_MANAGEMENT_IMAGE_URI) -f Dockerfile.service .
	$(DOCKER) buildx build --push --build-arg SERVICE_PATH=customer-management-service --build-arg SERVICE_PORT=8081 -t $(CUSTOMER_MANAGEMENT_IMAGE_URI) -f Dockerfile.service .
	$(DOCKER) buildx build --push --build-arg SERVICE_PATH=onboard-service --build-arg SERVICE_PORT=8080 -t $(ONBOARD_SERVICE_IMAGE_URI) -f Dockerfile.service .

ensure-ecr-repositories:
	$(call require,AWS_ACCOUNT_ID)
	$(AWS) ecr describe-repositories --region $(AWS_REGION) --repository-names account-management-service >$(NULL_DEVICE) 2>&1 || $(AWS) ecr create-repository --region $(AWS_REGION) --repository-name account-management-service
	$(AWS) ecr describe-repositories --region $(AWS_REGION) --repository-names customer-management-service >$(NULL_DEVICE) 2>&1 || $(AWS) ecr create-repository --region $(AWS_REGION) --repository-name customer-management-service
	$(AWS) ecr describe-repositories --region $(AWS_REGION) --repository-names onboard-service >$(NULL_DEVICE) 2>&1 || $(AWS) ecr create-repository --region $(AWS_REGION) --repository-name onboard-service

ecr-login:
	$(call require,AWS_ACCOUNT_ID)
	$(AWS) ecr get-login-password --region $(AWS_REGION) | $(DOCKER) login --username AWS --password-stdin $(ECR_REGISTRY)

docker-tag: docker-build
	$(call require,AWS_ACCOUNT_ID)
	$(DOCKER) tag $(IMAGE_PREFIX)/account-management-service:latest $(ACCOUNT_MANAGEMENT_IMAGE_URI)
	$(DOCKER) tag $(IMAGE_PREFIX)/customer-management-service:latest $(CUSTOMER_MANAGEMENT_IMAGE_URI)
	$(DOCKER) tag $(IMAGE_PREFIX)/onboard-service:latest $(ONBOARD_SERVICE_IMAGE_URI)

docker-push: ensure-ecr-repositories ecr-login docker-build-push

describe-stack-events:
	-$(DESCRIBE_STACK_EVENTS)

deploy-infra:
	$(call require,VPC_ID)
	$(call require,PUBLIC_SUBNET_IDS)
	$(AWS) cloudformation deploy --region $(AWS_REGION) --template-file infra/cloudformation.yaml --stack-name $(STACK_NAME) --capabilities CAPABILITY_NAMED_IAM --no-fail-on-empty-changeset --parameter-overrides EnvironmentName=$(ENVIRONMENT_NAME) VpcId=$(VPC_ID) PublicSubnetIds=$(PUBLIC_SUBNET_IDS) OnboardServiceImageUri=$(ONBOARD_SERVICE_IMAGE_URI) CustomerManagementServiceImageUri=$(CUSTOMER_MANAGEMENT_IMAGE_URI) AccountManagementServiceImageUri=$(ACCOUNT_MANAGEMENT_IMAGE_URI) $(DEPLOY_INFRA_ON_FAILURE)

deploy-production:
	$(MAKE) docker-push
	$(MAKE) deploy-infra

clean:
	$(GO) clean
	$(call RM_RF,$(BIN_DIR))
	$(call RM_FILE,coverage.out)
