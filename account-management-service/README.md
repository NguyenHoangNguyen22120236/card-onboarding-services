# Account Management Service

## Responsibility

`account-management-service` returns account interest details used by `onboard-service` during card onboarding.

## Architecture / Flow

```text
onboard-service
  -> GET /internal/accounts/{customerId}/interest-details
  -> account interest details response
```

## API List

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `GET` | `/internal/accounts/{customerId}/interest-details` | Get account interest details by customer ID. |

## Swagger Location

`swagger-internal.yaml`

## Generated Client Package Location

`pkg/account`

## Config Variables

No runtime config variables are currently read by this service binary. It listens on `:8082`.

## Local Run Command

From `card-onboarding-services`:

```sh
go run ./account-management-service
```

## Docker Build Command

From `card-onboarding-services`:

```sh
docker build --build-arg SERVICE_PATH=account-management-service --build-arg SERVICE_PORT=8082 -t card-onboarding/account-management-service:latest -f Dockerfile.service .
```

## Unit Test Command

From `card-onboarding-services`:

```sh
go test ./account-management-service/...
```

## Smoke Test Command

No separate smoke-test command is currently defined for this service.

## Deployment Command

From `card-onboarding-services`, deploy with the full services stack:

```sh
make deploy-production
```
