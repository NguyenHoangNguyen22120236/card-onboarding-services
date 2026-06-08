# API Contracts

## 1. Overview

All services must define OpenAPI specifications and generate Go server/client code using `oapi-codegen v2`.

Each service must provide:

```text
swagger.yaml
swagger-internal.yaml
oapi-codegen.yaml
```

Expected generated files:

```text
pkg/{service}/types.gen.go
pkg/{service}/client.gen.go
pkg/{service}/server.gen.go
```

## 2. onboard-service APIs

### POST /internal/cards/onboard

Used by card-onboarding-worker to submit a valid onboarding record.

Request example:

```json
{
  "correlationId": "corr-123",
  "jobId": "JOB-20260606-001",
  "recordId": "REC-001",
  "sourceFile": "cards_20260606.csv",
  "rowNumber": 2,
  "customerId": "CUST001",
  "cardType": "VISA",
  "cardNumber": "4111111111111111",
  "expiryDate": "12/28",
  "holderName": "Nguyen Van A",
  "email": "a@example.com"
}
```

Response example:

```json
{
  "customerId": "CUST001",
  "coreCustomerId": "CORE-CUST001",
  "accountId": "ACC-CUST001",
  "cardId": "CARD-CUST001-001",
  "status": "ONBOARDED"
}
```

### GET /internal/cards/{customerId}/status

Returns onboarding status for a customer.

Response example:

```json
{
  "customerId": "CUST001",
  "overallStatus": "SUCCEEDED",
  "customerRegistrationStatus": "SUCCEEDED",
  "interestDetailsStatus": "SUCCEEDED",
  "accountOnboardingStatus": "SUCCEEDED"
}
```

### GET /health

Health check endpoint.

## 3. customer-management-service APIs

### POST /internal/customers/register

Registers a customer in the mock core banking system.

Request example:

```json
{
  "correlationId": "corr-123",
  "customerId": "CUST001",
  "holderName": "Nguyen Van A",
  "email": "a@example.com"
}
```

Response example:

```json
{
  "customerId": "CUST001",
  "coreCustomerId": "CORE-CUST001",
  "status": "REGISTERED",
  "registeredAt": "2026-06-06T10:00:00Z"
}
```

Failure simulation:

- `customerId = CUST_FAIL_REGISTER` → HTTP 500
- `customerId = CUST_BAD_REQUEST` → HTTP 400

### GET /internal/customers/{customerId}

Returns customer information.

### GET /health

Health check endpoint.

## 4. account-management-service APIs

### GET /internal/accounts/{customerId}/interest-details

Returns account/product interest details.

Response example:

```json
{
  "customerId": "CUST001",
  "productCode": "SAVINGS_BASIC",
  "interestRate": 4.5,
  "interestType": "VARIABLE",
  "currency": "AUD"
}
```

Failure simulation:

- `customerId = CUST_FAIL_INTEREST` → HTTP 500
- `customerId = CUST_NO_INTEREST` → HTTP 404

### GET /health

Health check endpoint.

## 5. Common Error Response

All services should use a common error response format.

```json
{
  "code": "VALIDATION_ERROR",
  "message": "request validation failed",
  "correlationId": "corr-123",
  "details": [
    {
      "field": "customerId",
      "issue": "customerId is required"
    }
  ]
}
```

## 6. Makefile Commands

Each service should support:

```bash
make generate
make swagger-validate
make generate-check
```

`make generate-check` must fail if generated code is outdated.
