package account

import (
	"errors"
	"fmt"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
)

const (
	failInterestCustomerID = "CUST_FAIL_INTEREST"
	noInterestCustomerID   = "CUST_NO_INTEREST"

	errorCodeInternal   = "INTERNAL_ERROR"
	errorCodeNotFound   = "NOT_FOUND"
	errorCodeValidation = "VALIDATION_ERROR"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrInternal   = errors.New("internal error")
	ErrNotFound   = errors.New("not found")
)

type AccountService interface {
	GetInterestDetails(customerID string) (api.InterestDetailsResponse, error)
}

type Service struct{}

type ServiceError struct {
	Code    string
	Message string
	Details []api.ErrorDetail
	Err     error
}

func NewService() AccountService {
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

func (s *Service) GetInterestDetails(customerID string) (api.InterestDetailsResponse, error) {
	if customerID == "" {
		return api.InterestDetailsResponse{}, &ServiceError{
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

	switch customerID {
	case failInterestCustomerID:
		return api.InterestDetailsResponse{}, &ServiceError{
			Code:    errorCodeInternal,
			Message: "interest details lookup failed",
			Err:     ErrInternal,
		}
	case noInterestCustomerID:
		return api.InterestDetailsResponse{}, &ServiceError{
			Code:    errorCodeNotFound,
			Message: "interest details not found",
			Err:     ErrNotFound,
		}
	default:
		return api.InterestDetailsResponse{
			CustomerId:   customerID,
			ProductCode:  "SAVINGS_BASIC",
			InterestRate: 4.5,
			InterestType: "VARIABLE",
			Currency:     "AUD",
		}, nil
	}
}
