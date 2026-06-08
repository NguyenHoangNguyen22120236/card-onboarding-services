# Card Onboarding Services Architecture

## 1. Overview

The `card-onboarding-services` repository contains the backend services for the Card Onboarding File Processing Platform.

This repository owns three services:

1. `onboard-service`
2. `customer-management-service`
3. `account-management-service`

The main responsibility of this repository is to expose internal APIs, generate Go client packages from OpenAPI specifications, and implement the onboarding orchestration flow.

## 2. Services

### onboard-service

The `onboard-service` is the main orchestration service.

Responsibilities:

- Receive onboarding requests from `card-onboarding-worker`
- Check existing onboarding status by `customerId`
- Resume onboarding from the correct step if the customer was processed before
- Call `customer-management-service`
- Call `account-management-service`
- Save onboarding status to DynamoDB
- Save account details to DynamoDB
- Expose OpenAPI specification
- Expose generated Go client package

### customer-management-service

The `customer-management-service` is a mock core banking customer service.

Responsibilities:

- Register customer
- Return core customer information
- Support failure simulation for testing
- Expose OpenAPI specification
- Expose generated Go client package

### account-management-service

The `account-management-service` is a mock account/product service.

Responsibilities:

- Return customer interest details
- Support failure simulation for testing
- Expose OpenAPI specification
- Expose generated Go client package

## 3. Service Communication

Service-to-service communication must use generated Go clients from OpenAPI.

Raw HTTP calls are not allowed in business logic.

Required generated packages:

```text
onboard-service/pkg/onboard
customer-management-service/pkg/customer
account-management-service/pkg/account
```

Communication flow:

```
card-onboarding-worker
    ↓ generated onboard client
onboard-service
    ↓ generated customer client
customer-management-service

onboard-service
    ↓ generated account client
account-management-service
```

## 4. DynamoDB Ownership

The onboard-service owns two DynamoDB tables:

- `onboard-service-request-status`
- `onboard-service-account-details`

The `onboard-service-request-status` table stores workflow progress.

The `onboard-service-account-details` table stores business account and card details.

## 5. Resume and Idempotency

The onboard-service must check existing onboarding status by customerId before processing.

Resume rules:

- **No existing status** → start customer registration

- **customerRegistrationStatus = SUCCEEDED** → skip customer registration → continue from interest details

- **interestDetailsStatus = SUCCEEDED** → skip customer and account services → continue account/card save logic

- **overallStatus = SUCCEEDED** → return existing account details

## 6. Important Rules

- Do not use raw HTTP calls between services
- Use generated Go clients only
- Do not use NOT_STARTED status
- Do not initialize all step statuses at the beginning
- Only update the current onboarding step
- Use customerId for idempotency
- Mask card number before storing or logging
