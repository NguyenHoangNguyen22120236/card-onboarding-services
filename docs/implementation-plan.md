# Services Implementation Plan

## Phase 1: Repository Setup

Create repository:

```text
card-onboarding-services
```

Create service folders:

```text
onboard-service
customer-management-service
account-management-service
```

Add root files:

```text
go.mod
go.sum
Makefile
pipeline.yaml
README.md
VERSION
CHANGELOG.md
```

## Phase 2: OpenAPI Contracts

Create OpenAPI files for each service:

```text
swagger.yaml
swagger-internal.yaml
oapi-codegen.yaml
```

Start with:

```text
GET /health
```

Then add required internal APIs.

## Phase 3: Generate Code

Use `oapi-codegen v2` to generate:

```text
types.gen.go
client.gen.go
server.gen.go
```

Required packages:

```text
onboard-service/pkg/onboard
customer-management-service/pkg/customer
account-management-service/pkg/account
```

## Phase 4: Implement customer-management-service

Implement:

```text
POST /internal/customers/register
GET /internal/customers/{customerId}
GET /health
```

Add failure simulation:

```text
CUST_FAIL_REGISTER → HTTP 500
CUST_BAD_REQUEST → HTTP 400
```

## Phase 5: Implement account-management-service

Implement:

```text
GET /internal/accounts/{customerId}/interest-details
GET /health
```

Add failure simulation:

```text
CUST_FAIL_INTEREST → HTTP 500
CUST_NO_INTEREST → HTTP 404
```

## Phase 6: Implement onboard-service

Implement:

```text
POST /internal/cards/onboard
GET /internal/cards/{customerId}/status
GET /health
```

Add orchestration logic:

```text
Check status by customerId
Call customer-management-service if needed
Call account-management-service if needed
Save status
Save account details
Support resume behavior
Support idempotency
```

## Phase 7: Add DynamoDB Store Layer

Create store interfaces:

```text
RequestStatusStore
AccountDetailsStore
```

Start with in-memory implementation for local testing.

Later add DynamoDB implementation.

## Phase 8: Add Tests

Required tests:

```text
onboard-service handler test
successful onboarding service test
customer registration failure test
interest details failure test
resume from customerRegistrationStatus = SUCCEEDED
resume from interestDetailsStatus = SUCCEEDED
overallStatus = SUCCEEDED idempotent response
request status store test
account details store test
customer-management-service success/error tests
account-management-service success/error tests
```

## Phase 9: Add Dockerfiles

Each service should have a Dockerfile.

## Phase 10: Add README and CHANGELOG

Each service must have:

```text
README.md
CHANGELOG.md
```
