# Card Onboarding Services

Go HTTP services for the Card Onboarding File Processing Platform.

## Responsibility

This repository owns the synchronous service layer:

- `onboard-service`: orchestrates card onboarding, calls downstream customer and account services, exposes onboarding status, and optionally persists request state in DynamoDB.
- `customer-management-service`: registers customers and returns registered customer data.
- `account-management-service`: returns account interest details for registered customers.

## Architecture / Flow

```text
card-onboarding-worker
  -> onboard-service :8080
     -> customer-management-service :8081
     -> account-management-service :8082
     -> DynamoDB request status table, when configured
     -> DynamoDB account details table, when configured
```

Local runs use in-memory stores in `onboard-service` unless DynamoDB table environment variables are set. Production deployment runs all three services on ECS/Fargate behind a shared public Application Load Balancer.

## API List

### `onboard-service`

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `POST` | `/internal/cards/onboard` | Start or resume card onboarding for one customer/card request. |
| `GET` | `/internal/cards/{customerId}/status` | Return the saved onboarding status for a customer. |

### `customer-management-service`

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `POST` | `/internal/customers/register` | Register a customer. |
| `GET` | `/internal/customers/{customerId}` | Return registered customer data. |

### `account-management-service`

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `GET` | `/internal/accounts/{customerId}/interest-details` | Return account interest details for a customer. |

## Swagger Locations

- `onboard-service/swagger-internal.yaml`
- `customer-management-service/swagger-internal.yaml`
- `account-management-service/swagger-internal.yaml`

## Generated Client Packages

- `onboard-service/pkg/onboard`
- `customer-management-service/pkg/customer`
- `account-management-service/pkg/account`

Regenerate all OpenAPI types, servers, and clients:

```sh
make generate
```

## Config Variables

### Repository / Deployment

| Variable | Default | Description |
| --- | --- | --- |
| `AWS_REGION` | `ap-southeast-1` | AWS region for ECR and CloudFormation. |
| `AWS_ACCOUNT_ID` | empty | AWS account used to build ECR registry URLs. Required for push/deploy. |
| `IMAGE_PREFIX` | `card-onboarding` | Local Docker image prefix. |
| `IMAGE_TAG` | current git short SHA | ECR image tag. |
| `ENVIRONMENT_NAME` | `prod` | Environment suffix for deployed resources. |
| `STACK_NAME` | `card-onboarding-services-$(ENVIRONMENT_NAME)` | CloudFormation stack name. |
| `VPC_ID` | empty | VPC for ECS services. Required for deploy. |
| `PUBLIC_SUBNET_IDS` | empty | Comma-separated public subnet IDs. Required for deploy. |

### `onboard-service`

| Variable | Default | Description |
| --- | --- | --- |
| `SERVER_ADDRESS` | `:8080` | Listen address. Takes precedence over `PORT`. |
| `PORT` | empty | Port converted to `:<port>` when `SERVER_ADDRESS` is not set. |
| `CUSTOMER_MANAGEMENT_BASE_URL` | `http://localhost:8081` | Downstream customer service URL. |
| `ACCOUNT_MANAGEMENT_BASE_URL` | `http://localhost:8082` | Downstream account service URL. |
| `DOWNSTREAM_TIMEOUT` | `5s` | Go duration for downstream HTTP calls. Invalid values fall back to `5s`. |
| `REQUEST_STATUS_TABLE_NAME` | empty | DynamoDB table for onboarding request status. Empty uses in-memory storage. |
| `ACCOUNT_DETAILS_TABLE_NAME` | empty | DynamoDB table for account details. Empty uses in-memory storage. |

### `customer-management-service`

No runtime config variables are currently read by the service binary. It listens on `:8081`.

### `account-management-service`

No runtime config variables are currently read by the service binary. It listens on `:8082`.

## DynamoDB Table Usage

`onboard-service` uses DynamoDB only when table names are configured:

- `onboard-service-request-status-${EnvironmentName}` via `REQUEST_STATUS_TABLE_NAME`: keyed by `customerId`; stores overall status plus customer registration, interest details, and account onboarding step statuses.
- `onboard-service-account-details-${EnvironmentName}` via `ACCOUNT_DETAILS_TABLE_NAME`: keyed by `customerId`; stores account details returned during onboarding so resume and status flows can reuse persisted state.

If only one table variable is set, only that store uses DynamoDB and the other store remains in memory.

## Local Run Commands

Run each command in a separate terminal from the repository root:

```sh
go run ./account-management-service
go run ./customer-management-service
go run ./onboard-service
```

## Docker Build Command

Build all service images:

```sh
make docker-build
```

Build one service image:

```sh
docker build --build-arg SERVICE_PATH=onboard-service --build-arg SERVICE_PORT=8080 -t card-onboarding/onboard-service:latest -f Dockerfile.service .
```

Use `SERVICE_PATH=customer-management-service SERVICE_PORT=8081` or `SERVICE_PATH=account-management-service SERVICE_PORT=8082` for the other services.

## Unit Test Command

```sh
make test
```

## Smoke Test Command

There is no repository-level smoke-test target. The current local end-to-end simulation is:

```sh
go test ./onboard-service/internal/orchestration -run TestLocalE2ESimulation
```

## Deployment Command

Production build, push, and CloudFormation deploy:

```sh
make deploy-production
```

Deploy only the infrastructure with already-built image URIs:

```sh
make deploy-infra AWS_ACCOUNT_ID=123456789012 VPC_ID=vpc-xxxxxxxx PUBLIC_SUBNET_IDS=subnet-aaaaaaaa,subnet-bbbbbbbb
```

The deploy targets validate that `VPC_ID` and `PUBLIC_SUBNET_IDS` exist in `AWS_REGION` before starting CloudFormation. If an initial stack creation failed and the stack is in `ROLLBACK_COMPLETE`, delete the failed stack record before retrying:

```sh
aws cloudformation delete-stack --region ap-southeast-1 --stack-name card-onboarding-services-prod
aws cloudformation wait stack-delete-complete --region ap-southeast-1 --stack-name card-onboarding-services-prod
```

See `infra/README.md` for CloudFormation parameters and required AWS permissions.
