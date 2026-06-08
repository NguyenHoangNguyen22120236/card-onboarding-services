# Data Model

## 1. Overview

The `onboard-service` owns the DynamoDB data model.

It uses two tables:

```text
onboard-service-request-status
onboard-service-account-details
```

## 2. Table: onboard-service-request-status

### Purpose

Stores onboarding workflow status only.

### Primary Key

- **PK**: customerId

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| customerId | string | Unique customer identifier |
| overallStatus | string | Overall onboarding status |
| customerRegistrationStatus | string | Status of customer registration step |
| customerRegistrationMessage | string | Message for customer registration step |
| interestDetailsStatus | string | Status of interest details retrieval |
| interestDetailsMessage | string | Message for interest details step |
| accountOnboardingStatus | string | Status of account onboarding step |
| accountOnboardingMessage | string | Message for account onboarding step |
| createdAt | string | ISO 8601 timestamp of creation |
| updatedAt | string | ISO 8601 timestamp of last update |

### Status Values

Allowed values:

- `IN_PROGRESS`
- `SUCCEEDED`
- `FAILED`

**Do not use:**

- `NOT_STARTED`

If a step has not been reached yet, its status should be empty or not created.

### Example Failed Item

```json
{
  "customerId": "CUST001",
  "overallStatus": "FAILED",
  "customerRegistrationStatus": "SUCCEEDED",
  "customerRegistrationMessage": "Customer registered successfully",
  "interestDetailsStatus": "FAILED",
  "interestDetailsMessage": "Account Management Service timeout",
  "createdAt": "2026-06-06T10:00:00Z",
  "updatedAt": "2026-06-06T10:01:00Z"
}
```

## 3. Table: onboard-service-account-details

### Purpose

Stores business account, customer, product, and card details.

### Primary Key

- **PK**: customerId

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| customerId | string | Unique customer identifier |
| coreCustomerId | string | Core banking system customer ID |
| customerName | string | Full name of the customer |
| email | string | Email address of the customer |
| productCode | string | Product code from account service |
| interestRate | number | Interest rate percentage |
| interestType | string | Type of interest (FIXED or VARIABLE) |
| currency | string | Currency code (e.g., AUD, USD) |
| accountId | string | Unique account identifier |
| cardId | string | Unique card identifier |
| cardType | string | Card type (e.g., VISA, MASTERCARD) |
| cardNumberMasked | string | Masked card number (e.g., ************1111) |
| createdAt | string | ISO 8601 timestamp of creation |
| updatedAt | string | ISO 8601 timestamp of last update |

### Example Item

```json
{
  "customerId": "CUST001",
  "coreCustomerId": "CORE-CUST001",
  "customerName": "Nguyen Van A",
  "email": "a@example.com",
  "productCode": "SAVINGS_BASIC",
  "interestRate": 4.5,
  "interestType": "VARIABLE",
  "currency": "AUD",
  "accountId": "ACC-CUST001",
  "cardId": "CARD-CUST001-001",
  "cardType": "VISA",
  "cardNumberMasked": "************1111",
  "createdAt": "2026-06-06T10:00:00Z",
  "updatedAt": "2026-06-06T10:02:00Z"
}
```

## 4. Status Update Rules

### Start Onboarding

Update request status:

- `overallStatus = IN_PROGRESS`
- `customerRegistrationStatus = IN_PROGRESS`
- `updatedAt = current timestamp`

Do not create future step statuses yet.

### Customer Registration Success

Update request status:

- `customerRegistrationStatus = SUCCEEDED`
- `customerRegistrationMessage = Customer registered successfully`
- `updatedAt = current timestamp`

Update account details:

- `customerId`
- `coreCustomerId`
- `customerName`
- `email`
- `updatedAt`

Then mark next step:

- `interestDetailsStatus = IN_PROGRESS`
- `overallStatus = IN_PROGRESS`
- `updatedAt = current timestamp`

### Customer Registration Failure

- `overallStatus = FAILED`
- `customerRegistrationStatus = FAILED`
- `customerRegistrationMessage = error message`
- `updatedAt = current timestamp`

### Interest Details Success

Update request status:

- `interestDetailsStatus = SUCCEEDED`
- `interestDetailsMessage = Interest details fetched successfully`
- `updatedAt = current timestamp`

Update account details:

- `productCode`
- `interestRate`
- `interestType`
- `currency`
- `updatedAt`

Then mark next step:

- `accountOnboardingStatus = IN_PROGRESS`
- `overallStatus = IN_PROGRESS`
- `updatedAt = current timestamp`

### Interest Details Failure

- `overallStatus = FAILED`
- `interestDetailsStatus = FAILED`
- `interestDetailsMessage = error message`
- `updatedAt = current timestamp`

### Account Onboarding Success

Update request status:

- `overallStatus = SUCCEEDED`
- `accountOnboardingStatus = SUCCEEDED`
- `accountOnboardingMessage = Account onboarded successfully`
- `updatedAt = current timestamp`

Update account details:

- `accountId`
- `cardId`
- `cardType`
- `cardNumberMasked`
- `updatedAt`

### Account Onboarding Failure

- `overallStatus = FAILED`
- `accountOnboardingStatus = FAILED`
- `accountOnboardingMessage = error message`
- `updatedAt = current timestamp`
