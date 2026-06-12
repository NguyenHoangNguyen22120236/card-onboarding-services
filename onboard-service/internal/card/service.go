package card

import (
	"context"
	"errors"
	"fmt"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/client"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/orchestration"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/pkg/onboard"
)

const (
	errorCodeBadRequest = "BAD_REQUEST"
	errorCodeInternal   = "INTERNAL_ERROR"
	errorCodeNotFound   = "NOT_FOUND"
	errorCodeValidation = "VALIDATION_ERROR"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrInternal   = errors.New("internal error")
	ErrNotFound   = errors.New("not found")
)

type CardService interface {
	OnboardCard(ctx context.Context, req api.OnboardCardRequest) (api.OnboardCardResponse, error)
	GetCardOnboardingStatus(ctx context.Context, customerID string) (api.CardOnboardingStatusResponse, error)
}

type Service struct {
	orchestration orchestration.Service
	statusStore   store.RequestStatusStore
}

type ServiceError struct {
	Code    string
	Message string
	Details []api.ErrorDetail
	Err     error
}

func NewService(orchestration orchestration.Service, statusStore store.RequestStatusStore) CardService {
	return &Service{
		orchestration: orchestration,
		statusStore:   statusStore,
	}
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

func (s *Service) OnboardCard(ctx context.Context, req api.OnboardCardRequest) (api.OnboardCardResponse, error) {
	if err := ctx.Err(); err != nil {
		return api.OnboardCardResponse{}, mapServiceError(err)
	}

	if err := validateOnboardCardRequest(req); err != nil {
		return api.OnboardCardResponse{}, err
	}

	resp, err := s.orchestration.OnboardCard(ctx, entity.OnboardingRequest{
		CorrelationID: req.CorrelationId,
		JobID:         req.JobId,
		RecordID:      req.RecordId,
		SourceFile:    req.SourceFile,
		RowNumber:     req.RowNumber,
		CustomerID:    req.CustomerId,
		CardType:      req.CardType,
		CardNumber:    req.CardNumber,
		ExpiryDate:    req.ExpiryDate,
		HolderName:    req.HolderName,
		Email:         string(req.Email),
	})
	if err != nil {
		return api.OnboardCardResponse{}, mapServiceError(err)
	}

	return resp, nil
}

func (s *Service) GetCardOnboardingStatus(ctx context.Context, customerID string) (api.CardOnboardingStatusResponse, error) {
	if err := ctx.Err(); err != nil {
		return api.CardOnboardingStatusResponse{}, mapServiceError(err)
	}

	if customerID == "" {
		return api.CardOnboardingStatusResponse{}, &ServiceError{
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

	status, err := s.statusStore.GetByCustomerID(ctx, customerID)
	if err != nil {
		return api.CardOnboardingStatusResponse{}, mapServiceError(err)
	}

	return api.CardOnboardingStatusResponse{
		CustomerId:                 status.CustomerID,
		OverallStatus:              string(status.OverallStatus),
		CustomerRegistrationStatus: string(status.CustomerRegistrationStatus),
		InterestDetailsStatus:      string(status.InterestDetailsStatus),
		AccountOnboardingStatus:    string(status.AccountOnboardingStatus),
	}, nil
}

func validateOnboardCardRequest(req api.OnboardCardRequest) error {
	var details []api.ErrorDetail

	addRequired := func(field string) {
		details = append(details, api.ErrorDetail{
			Field: field,
			Issue: field + " is required",
		})
	}

	if req.CorrelationId == "" {
		addRequired("correlationId")
	}
	if req.JobId == "" {
		addRequired("jobId")
	}
	if req.RecordId == "" {
		addRequired("recordId")
	}
	if req.SourceFile == "" {
		addRequired("sourceFile")
	}
	if req.RowNumber == 0 {
		addRequired("rowNumber")
	}
	if req.CustomerId == "" {
		addRequired("customerId")
	}
	if req.CardType == "" {
		addRequired("cardType")
	}
	if req.CardNumber == "" {
		addRequired("cardNumber")
	}
	if req.ExpiryDate == "" {
		addRequired("expiryDate")
	}
	if req.HolderName == "" {
		addRequired("holderName")
	}
	if req.Email == "" {
		addRequired("email")
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

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, client.ErrDownstreamBadRequest):
		return &ServiceError{
			Code:    errorCodeBadRequest,
			Message: "card onboarding rejected",
			Err:     ErrBadRequest,
		}
	case errors.Is(err, store.ErrNotFound), errors.Is(err, client.ErrDownstreamNotFound):
		return &ServiceError{
			Code:    errorCodeNotFound,
			Message: "card onboarding status not found",
			Err:     ErrNotFound,
		}
	case errors.Is(err, client.ErrDownstreamInternal), errors.Is(err, client.ErrDownstreamTimeout), errors.Is(err, context.Canceled), errors.Is(err, context.DeadlineExceeded):
		return &ServiceError{
			Code:    errorCodeInternal,
			Message: "card onboarding failed",
			Err:     ErrInternal,
		}
	default:
		return &ServiceError{
			Code:    errorCodeInternal,
			Message: "internal service error",
			Err:     ErrInternal,
		}
	}
}
