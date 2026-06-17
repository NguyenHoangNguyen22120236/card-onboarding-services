package store

import (
	"testing"
	"time"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

func TestDynamoDBRequestStatusMappingRoundTrip(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 11, 12, 123456789, time.UTC)
	status := entity.RequestStatus{
		CustomerID:                  "customer-1",
		OverallStatus:               entity.StatusInProgress,
		CustomerRegistrationStatus:  entity.StatusSucceeded,
		CustomerRegistrationMessage: "customer registered",
		InterestDetailsStatus:       entity.StatusSucceeded,
		InterestDetailsMessage:      "interest attached",
		AccountOnboardingStatus:     entity.StatusFailed,
		AccountOnboardingMessage:    "account failed",
		CreatedAt:                   now,
		UpdatedAt:                   now.Add(time.Minute),
	}

	item, err := marshalRequestStatusItem(status)
	if err != nil {
		t.Fatalf("marshalRequestStatusItem() error = %v", err)
	}

	assertStringAttribute(t, item, "customerId", status.CustomerID)

	got, err := unmarshalRequestStatusItem(item)
	if err != nil {
		t.Fatalf("unmarshalRequestStatusItem() error = %v", err)
	}
	if got != status {
		t.Fatalf("unmarshalRequestStatusItem() = %#v, want %#v", got, status)
	}
}

func TestDynamoDBAccountDetailsMappingRoundTrip(t *testing.T) {
	now := time.Date(2026, 6, 16, 10, 11, 12, 123456789, time.UTC)
	details := entity.AccountDetails{
		CustomerID:       "customer-1",
		CoreCustomerID:   "core-customer-1",
		CustomerName:     "Alex Customer",
		Email:            "alex@example.com",
		ProductCode:      "CARD-GOLD",
		InterestRate:     1.25,
		InterestType:     "FIXED",
		Currency:         "USD",
		AccountID:        "account-1",
		CardID:           "card-1",
		CardType:         "GOLD",
		CardNumberMasked: "**** **** **** 1111",
		CreatedAt:        now,
		UpdatedAt:        now.Add(time.Minute),
	}

	item, err := marshalAccountDetailsItem(details)
	if err != nil {
		t.Fatalf("marshalAccountDetailsItem() error = %v", err)
	}

	assertStringAttribute(t, item, "customerId", details.CustomerID)

	got, err := unmarshalAccountDetailsItem(item)
	if err != nil {
		t.Fatalf("unmarshalAccountDetailsItem() error = %v", err)
	}
	if got != details {
		t.Fatalf("unmarshalAccountDetailsItem() = %#v, want %#v", got, details)
	}
}

func TestDynamoDBCustomerIDKey(t *testing.T) {
	key := customerIDKey("customer-1")
	assertStringAttribute(t, key, "customerId", "customer-1")
}

func assertStringAttribute(t *testing.T, item map[string]types.AttributeValue, name string, want string) {
	t.Helper()

	got, ok := item[name].(*types.AttributeValueMemberS)
	if !ok {
		t.Fatalf("%s attribute = %T, want string attribute", name, item[name])
	}
	if got.Value != want {
		t.Fatalf("%s attribute = %q, want %q", name, got.Value, want)
	}
}
