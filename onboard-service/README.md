# Onboard Service

## Responsibility

`onboard-service` accepts card onboarding requests, orchestrates customer registration and account interest lookup, resumes partially failed onboarding attempts, exposes saved onboarding status, and optionally persists state in DynamoDB.

## Architecture / Flow

```text
POST /internal/cards/onboard
  -> validate request
  -> load or create request status
  -> customer-management-service /internal/customers/register
  -> account-management-service /internal/accounts/{customerId}/interest-details
  -> save account details
  -> mark onboarding status SUCCEEDED or FAILED

GET /internal/cards/{customerId}/status
  -> read request status store
```

The service uses in-memory stores by default. When DynamoDB table names are configured, it uses AWS SDK DynamoDB stores for the configured tables.

## API List

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `POST` | `/internal/cards/onboard` | Start or resume card onboarding. |
| `GET` | `/internal/cards/{customerId}/status` | Get onboarding status by customer ID. |

## Swagger Location

`swagger-internal.yaml`

## Generated Client Package Location

`pkg/onboard`

## Config Variables

| Variable | Default | Description |
| --- | --- | --- |
| `SERVER_ADDRESS` | `:8080` | HTTP listen address. |
| `PORT` | empty | Used as `:<port>` when `SERVER_ADDRESS` is not set. |
| `CUSTOMER_MANAGEMENT_BASE_URL` | `http://localhost:8081` | Customer service base URL. |
| `ACCOUNT_MANAGEMENT_BASE_URL` | `http://localhost:8082` | Account service base URL. |
| `DOWNSTREAM_TIMEOUT` | `5s` | Go duration for downstream HTTP calls. |
| `REQUEST_STATUS_TABLE_NAME` | empty | DynamoDB request status table; empty uses memory. |
| `ACCOUNT_DETAILS_TABLE_NAME` | empty | DynamoDB account details table; empty uses memory. |

## DynamoDB Table Usage

- `REQUEST_STATUS_TABLE_NAME`: table keyed by `customerId`; stores overall and per-step onboarding status.
- `ACCOUNT_DETAILS_TABLE_NAME`: table keyed by `customerId`; stores account details returned from the account service.

## Local Run Command

From `card-onboarding-services`:

```sh
go run ./onboard-service
```

## Docker Build Command

From `card-onboarding-services`:

```sh
docker build --build-arg SERVICE_PATH=onboard-service --build-arg SERVICE_PORT=8080 -t card-onboarding/onboard-service:latest -f Dockerfile.service .
```

## Unit Test Command

From `card-onboarding-services`:

```sh
go test ./onboard-service/...
```

## Smoke Test Command

From `card-onboarding-services`:

```sh
go test ./onboard-service/internal/orchestration -run TestLocalE2ESimulation
```

## Deployment Command

From `card-onboarding-services`, deploy with the full services stack:

```sh
make deploy-production
```
