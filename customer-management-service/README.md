# Customer Management Service

## Responsibility

`customer-management-service` registers customers for onboarding and returns registered customer records by customer ID.

## Architecture / Flow

```text
onboard-service
  -> POST /internal/customers/register
  -> customer validation and in-memory registration

onboard-service or operator
  -> GET /internal/customers/{customerId}
  -> registered customer response
```

## API List

| Method | Path | Purpose |
| --- | --- | --- |
| `GET` | `/health` | Health check. |
| `POST` | `/internal/customers/register` | Register a customer. |
| `GET` | `/internal/customers/{customerId}` | Get customer by customer ID. |

## Swagger Location

`swagger-internal.yaml`

## Generated Client Package Location

`pkg/customer`

## Config Variables

No runtime config variables are currently read by this service binary. It listens on `:8081`.

## Local Run Command

From `card-onboarding-services`:

```sh
go run ./customer-management-service
```

## Docker Build Command

From `card-onboarding-services`:

```sh
docker build --build-arg SERVICE_PATH=customer-management-service --build-arg SERVICE_PORT=8081 -t card-onboarding/customer-management-service:latest -f Dockerfile.service .
```

## Unit Test Command

From `card-onboarding-services`:

```sh
go test ./customer-management-service/...
```

## Smoke Test Command

No separate smoke-test command is currently defined for this service.

## Deployment Command

From `card-onboarding-services`, deploy with the full services stack:

```sh
make deploy-production
```
