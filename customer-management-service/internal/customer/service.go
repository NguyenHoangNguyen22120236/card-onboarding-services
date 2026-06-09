package customer

import (
	"context"
	"errors"
	"fmt"
	"time"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
)

const (
	failRegisterCustomerID = "CUST_FAIL_REGISTER"
	badRequestCustomerID   = "CUST_BAD_REQUEST"

	errorCodeBadRequest = "BAD_REQUEST"
	errorCodeInternal   = "INTERNAL_ERROR"
	errorCodeValidation = "VALIDATION_ERROR"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrInternal   = errors.New("internal error")
)

type CustomerService interface {
	RegisterCustomer(ctx context.Context, req api.RegisterCustomerRequest) (api.RegisterCustomerResponse, error)
	GetCustomer(ctx context.Context, customerID string) (api.RegisterCustomerResponse, error)
}

type Service struct{}

type ServiceError struct {
	Code    string
	Message string
	Details []api.ErrorDetail
	Err     error
}

func NewService() CustomerService {
	return &Service{}
}

func (e *ServiceError) Error() string {
	if e == nil {
		return ""
	}

	if e.Err == nil {
		return e.Message
	}

	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func (e *ServiceError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Err
}

func (s *Service) RegisterCustomer(ctx context.Context, req api.RegisterCustomerRequest) (api.RegisterCustomerResponse, error) {
	if err := ctx.Err(); err != nil {
		return api.RegisterCustomerResponse{}, err
	}

	if err := validateRegisterCustomerRequest(req); err != nil {
		return api.RegisterCustomerResponse{}, err
	}

	switch req.CustomerId {
	case failRegisterCustomerID:
		return api.RegisterCustomerResponse{}, &ServiceError{
			Code:    errorCodeInternal,
			Message: "customer registration failed",
			Err:     ErrInternal,
		}
	case badRequestCustomerID:
		return api.RegisterCustomerResponse{}, &ServiceError{
			Code:    errorCodeBadRequest,
			Message: "customer registration rejected",
			Err:     ErrBadRequest,
		}
	default:
		return buildCustomerResponse(req.CustomerId), nil
	}
}

func (s *Service) GetCustomer(ctx context.Context, customerID string) (api.RegisterCustomerResponse, error) {
	if err := ctx.Err(); err != nil {
		return api.RegisterCustomerResponse{}, err
	}

	if customerID == "" {
		return api.RegisterCustomerResponse{}, &ServiceError{
			Code:    errorCodeValidation,
			Message: "request validation failed",
			Details: []api.ErrorDetail{
				{
					Field: "customerId",
					Issue: "customerId is required",
				},
			},
			Err: ErrBadRequest,
		}
	}

	return buildCustomerResponse(customerID), nil
}

func validateRegisterCustomerRequest(req api.RegisterCustomerRequest) error {
	var details []api.ErrorDetail

	if req.CustomerId == "" {
		details = append(details, api.ErrorDetail{
			Field: "customerId",
			Issue: "customerId is required",
		})
	}

	if req.HolderName == "" {
		details = append(details, api.ErrorDetail{
			Field: "holderName",
			Issue: "holderName is required",
		})
	}

	if req.Email == "" {
		details = append(details, api.ErrorDetail{
			Field: "email",
			Issue: "email is required",
		})
	}

	if len(details) == 0 {
		return nil
	}

	return &ServiceError{
		Code:    errorCodeValidation,
		Message: "request validation failed",
		Details: details,
		Err:     ErrBadRequest,
	}
}

func buildCustomerResponse(customerID string) api.RegisterCustomerResponse {
	return api.RegisterCustomerResponse{
		CustomerId:     customerID,
		CoreCustomerId: "CORE-" + customerID,
		Status:         "REGISTERED",
		RegisteredAt:   time.Now().UTC(),
	}
}
