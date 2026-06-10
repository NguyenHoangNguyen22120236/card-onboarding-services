package account

import (
	"errors"
	"testing"
)

func TestServiceGetInterestDetailsSuccess(t *testing.T) {
	service := &Service{}

	got, err := service.GetInterestDetails("CUST123")
	if err != nil {
		t.Fatalf("GetInterestDetails() error = %v, want nil", err)
	}

	if got.CustomerId != "CUST123" {
		t.Errorf("CustomerId = %q, want %q", got.CustomerId, "CUST123")
	}
	if got.ProductCode != "SAVINGS_BASIC" {
		t.Errorf("ProductCode = %q, want %q", got.ProductCode, "SAVINGS_BASIC")
	}
	if got.InterestRate != 4.5 {
		t.Errorf("InterestRate = %v, want %v", got.InterestRate, 4.5)
	}
	if got.InterestType != "VARIABLE" {
		t.Errorf("InterestType = %q, want %q", got.InterestType, "VARIABLE")
	}
	if got.Currency != "AUD" {
		t.Errorf("Currency = %q, want %q", got.Currency, "AUD")
	}
}

func TestServiceGetInterestDetailsMissingCustomerIDReturnsValidationError(t *testing.T) {
	service := &Service{}

	_, err := service.GetInterestDetails("")
	if err == nil {
		t.Fatal("GetInterestDetails() error = nil, want validation error")
	}
	if !errors.Is(err, ErrBadRequest) {
		t.Fatalf("GetInterestDetails() error = %v, want ErrBadRequest", err)
	}

	serviceErr := assertServiceError(t, err)
	if serviceErr.Code != errorCodeValidation {
		t.Errorf("Code = %q, want %q", serviceErr.Code, errorCodeValidation)
	}
	if serviceErr.Message != "request validation failed" {
		t.Errorf("Message = %q, want %q", serviceErr.Message, "request validation failed")
	}
	if len(serviceErr.Details) != 1 {
		t.Fatalf("Details length = %d, want 1", len(serviceErr.Details))
	}
	if serviceErr.Details[0].Field != "customerId" {
		t.Errorf("Details[0].Field = %q, want %q", serviceErr.Details[0].Field, "customerId")
	}
	if serviceErr.Details[0].Issue != "customerId is required" {
		t.Errorf("Details[0].Issue = %q, want %q", serviceErr.Details[0].Issue, "customerId is required")
	}
}

func TestServiceGetInterestDetailsNoInterestReturnsNotFoundError(t *testing.T) {
	service := &Service{}

	_, err := service.GetInterestDetails(noInterestCustomerID)
	if err == nil {
		t.Fatal("GetInterestDetails() error = nil, want not found error")
	}
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetInterestDetails() error = %v, want ErrNotFound", err)
	}

	serviceErr := assertServiceError(t, err)
	if serviceErr.Code != errorCodeNotFound {
		t.Errorf("Code = %q, want %q", serviceErr.Code, errorCodeNotFound)
	}
	if serviceErr.Message != "interest details not found" {
		t.Errorf("Message = %q, want %q", serviceErr.Message, "interest details not found")
	}
	if len(serviceErr.Details) != 0 {
		t.Errorf("Details length = %d, want 0", len(serviceErr.Details))
	}
}

func TestServiceGetInterestDetailsFailInterestReturnsInternalError(t *testing.T) {
	service := &Service{}

	_, err := service.GetInterestDetails(failInterestCustomerID)
	if err == nil {
		t.Fatal("GetInterestDetails() error = nil, want internal error")
	}
	if !errors.Is(err, ErrInternal) {
		t.Fatalf("GetInterestDetails() error = %v, want ErrInternal", err)
	}

	serviceErr := assertServiceError(t, err)
	if serviceErr.Code != errorCodeInternal {
		t.Errorf("Code = %q, want %q", serviceErr.Code, errorCodeInternal)
	}
	if serviceErr.Message != "interest details lookup failed" {
		t.Errorf("Message = %q, want %q", serviceErr.Message, "interest details lookup failed")
	}
	if len(serviceErr.Details) != 0 {
		t.Errorf("Details length = %d, want 0", len(serviceErr.Details))
	}
}

func assertServiceError(t *testing.T, err error) *ServiceError {
	t.Helper()

	var serviceErr *ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("error type = %T, want *ServiceError", err)
	}

	return serviceErr
}
