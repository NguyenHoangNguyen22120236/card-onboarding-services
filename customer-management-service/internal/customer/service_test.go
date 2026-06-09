package customer

import (
	"context"
	"errors"
	"testing"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
)

func TestService_RegisterCustomer_Success(t *testing.T) {
	t.Parallel()

	service := &Service{}
	req := validRegisterCustomerRequest()

	resp, err := service.RegisterCustomer(context.Background(), req)
	if err != nil {
		t.Fatalf("RegisterCustomer() error = %v, want nil", err)
	}

	assertCustomerResponse(t, resp, req.CustomerId)
}

func TestService_RegisterCustomer_MissingCustomerIDReturnsValidationError(t *testing.T) {
	t.Parallel()

	service := &Service{}
	req := validRegisterCustomerRequest()
	req.CustomerId = ""

	_, err := service.RegisterCustomer(context.Background(), req)

	assertServiceError(t, err, errorCodeValidation, "request validation failed", ErrBadRequest)
	assertErrorDetails(t, err, []api.ErrorDetail{
		{
			Field: "customerId",
			Issue: "customerId is required",
		},
	})
}

func TestService_RegisterCustomer_MissingHolderNameReturnsValidationError(t *testing.T) {
	t.Parallel()

	service := &Service{}
	req := validRegisterCustomerRequest()
	req.HolderName = ""

	_, err := service.RegisterCustomer(context.Background(), req)

	assertServiceError(t, err, errorCodeValidation, "request validation failed", ErrBadRequest)
	assertErrorDetails(t, err, []api.ErrorDetail{
		{
			Field: "holderName",
			Issue: "holderName is required",
		},
	})
}

func TestService_RegisterCustomer_MissingEmailReturnsValidationError(t *testing.T) {
	t.Parallel()

	service := &Service{}
	req := validRegisterCustomerRequest()
	req.Email = ""

	_, err := service.RegisterCustomer(context.Background(), req)

	assertServiceError(t, err, errorCodeValidation, "request validation failed", ErrBadRequest)
	assertErrorDetails(t, err, []api.ErrorDetail{
		{
			Field: "email",
			Issue: "email is required",
		},
	})
}

func TestService_RegisterCustomer_FailRegisterReturnsInternalError(t *testing.T) {
	t.Parallel()

	service := &Service{}
	req := validRegisterCustomerRequest()
	req.CustomerId = failRegisterCustomerID

	_, err := service.RegisterCustomer(context.Background(), req)

	assertServiceError(t, err, errorCodeInternal, "customer registration failed", ErrInternal)
}

func TestService_RegisterCustomer_BadRequestReturnsBadRequestError(t *testing.T) {
	t.Parallel()

	service := &Service{}
	req := validRegisterCustomerRequest()
	req.CustomerId = badRequestCustomerID

	_, err := service.RegisterCustomer(context.Background(), req)

	assertServiceError(t, err, errorCodeBadRequest, "customer registration rejected", ErrBadRequest)
}

func TestService_GetCustomer_Success(t *testing.T) {
	t.Parallel()

	service := &Service{}
	customerID := "CUST-123"

	resp, err := service.GetCustomer(context.Background(), customerID)
	if err != nil {
		t.Fatalf("GetCustomer() error = %v, want nil", err)
	}

	assertCustomerResponse(t, resp, customerID)
}

func validRegisterCustomerRequest() api.RegisterCustomerRequest {
	return api.RegisterCustomerRequest{
		CorrelationId: "corr-123",
		CustomerId:    "CUST-123",
		Email:         "holder@example.com",
		HolderName:    "Example Holder",
	}
}

func assertCustomerResponse(t *testing.T, resp api.RegisterCustomerResponse, customerID string) {
	t.Helper()

	if resp.CustomerId != customerID {
		t.Fatalf("CustomerId = %q, want %q", resp.CustomerId, customerID)
	}

	wantCoreCustomerID := "CORE-" + customerID
	if resp.CoreCustomerId != wantCoreCustomerID {
		t.Fatalf("CoreCustomerId = %q, want %q", resp.CoreCustomerId, wantCoreCustomerID)
	}

	if resp.Status != "REGISTERED" {
		t.Fatalf("Status = %q, want REGISTERED", resp.Status)
	}

	if resp.RegisteredAt.IsZero() {
		t.Fatal("RegisteredAt is zero, want non-zero time")
	}
}

func assertServiceError(t *testing.T, err error, code string, message string, wrapped error) {
	t.Helper()

	var serviceErr *ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("error = %v, want *ServiceError", err)
	}

	if serviceErr.Code != code {
		t.Fatalf("ServiceError.Code = %q, want %q", serviceErr.Code, code)
	}

	if serviceErr.Message != message {
		t.Fatalf("ServiceError.Message = %q, want %q", serviceErr.Message, message)
	}

	if !errors.Is(err, wrapped) {
		t.Fatalf("error = %v, want to wrap %v", err, wrapped)
	}
}

func assertErrorDetails(t *testing.T, err error, want []api.ErrorDetail) {
	t.Helper()

	var serviceErr *ServiceError
	if !errors.As(err, &serviceErr) {
		t.Fatalf("error = %v, want *ServiceError", err)
	}

	if len(serviceErr.Details) != len(want) {
		t.Fatalf("len(ServiceError.Details) = %d, want %d", len(serviceErr.Details), len(want))
	}

	for i := range want {
		if serviceErr.Details[i] != want[i] {
			t.Fatalf("ServiceError.Details[%d] = %#v, want %#v", i, serviceErr.Details[i], want[i])
		}
	}
}
