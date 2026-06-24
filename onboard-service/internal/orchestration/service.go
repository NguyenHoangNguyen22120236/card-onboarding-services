package orchestration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	accountapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
	customerapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/client"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/observability"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/pkg/onboard"
	openapi_types "github.com/oapi-codegen/runtime/types"
)

const (
	customerRegistrationSuccessMessage = "Customer registered successfully"
	interestDetailsSuccessMessage      = "Interest details fetched successfully"
	accountOnboardingSuccessMessage    = "Account onboarded successfully"
)

type Service interface {
	OnboardCard(ctx context.Context, req entity.OnboardingRequest) (api.OnboardCardResponse, error)
}

type service struct {
	statusStore         store.RequestStatusStore
	accountDetailsStore store.AccountDetailsStore
	customerClient      client.CustomerClient
	accountClient       client.AccountClient
}

func NewService(
	statusStore store.RequestStatusStore,
	accountDetailsStore store.AccountDetailsStore,
	customerClient client.CustomerClient,
	accountClient client.AccountClient,
) Service {
	return &service{
		statusStore:         statusStore,
		accountDetailsStore: accountDetailsStore,
		customerClient:      customerClient,
		accountClient:       accountClient,
	}
}

func (s *service) OnboardCard(ctx context.Context, req entity.OnboardingRequest) (api.OnboardCardResponse, error) {
	if err := ctx.Err(); err != nil {
		return api.OnboardCardResponse{}, err
	}

	status, newStatus, err := s.loadOrCreateStatus(ctx, req)
	if err != nil {
		return api.OnboardCardResponse{}, err
	}
	if !newStatus && !isCompleted(status) {
		observability.LogCount(observability.MetricResumeCount, metricFields(req, "resume", "used"))
	}

	if isCompleted(status) {
		details, err := s.accountDetailsStore.GetByCustomerID(ctx, req.CustomerID)
		if err != nil {
			return api.OnboardCardResponse{}, err
		}
		return buildResponse(details, status), nil
	}

	if shouldRegisterCustomer(status, newStatus) {
		status, err = s.registerCustomer(ctx, req, status)
		if err != nil {
			return api.OnboardCardResponse{}, err
		}
	}

	if shouldFetchInterestDetails(status) {
		status, err = s.fetchInterestDetails(ctx, req, status)
		if err != nil {
			return api.OnboardCardResponse{}, err
		}
	}

	if shouldOnboardAccount(status) {
		status, err = s.onboardAccount(ctx, req, status)
		if err != nil {
			return api.OnboardCardResponse{}, err
		}
	}

	if !isCompleted(status) {
		return api.OnboardCardResponse{}, fmt.Errorf("onboarding is not complete for customer %s", req.CustomerID)
	}

	details, err := s.accountDetailsStore.GetByCustomerID(ctx, req.CustomerID)
	if err != nil {
		return api.OnboardCardResponse{}, err
	}

	return buildResponse(details, status), nil
}

func (s *service) loadOrCreateStatus(ctx context.Context, req entity.OnboardingRequest) (entity.RequestStatus, bool, error) {
	status, err := s.statusStore.GetByCustomerID(ctx, req.CustomerID)
	if err == nil {
		return status, false, nil
	}
	if !errors.Is(err, store.ErrNotFound) {
		return entity.RequestStatus{}, false, err
	}

	now := time.Now().UTC()
	status = entity.RequestStatus{
		CustomerID:                 req.CustomerID,
		OverallStatus:              entity.StatusInProgress,
		CustomerRegistrationStatus: entity.StatusInProgress,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}

	if err := s.statusStore.Save(ctx, status); err != nil {
		return entity.RequestStatus{}, false, err
	}

	return status, true, nil
}

func (s *service) registerCustomer(ctx context.Context, req entity.OnboardingRequest, status entity.RequestStatus) (entity.RequestStatus, error) {
	resp, err := s.customerClient.RegisterCustomer(ctx, customerapi.RegisterCustomerRequest{
		CorrelationId: req.CorrelationID,
		CustomerId:    req.CustomerID,
		Email:         openapi_types.Email(req.Email),
		HolderName:    req.HolderName,
	})
	if err != nil {
		observability.LogCount(observability.MetricCustomerRegisterFailedCount, metricFields(req, "customer_register", "failed"))
		status.OverallStatus = entity.StatusFailed
		status.CustomerRegistrationStatus = entity.StatusFailed
		status.CustomerRegistrationMessage = err.Error()
		status.UpdatedAt = time.Now().UTC()
		if updateErr := s.statusStore.Update(ctx, status); updateErr != nil {
			return status, updateErr
		}
		return status, err
	}

	details := entity.AccountDetails{
		CustomerID:     req.CustomerID,
		CoreCustomerID: resp.CoreCustomerId,
		CustomerName:   req.HolderName,
		Email:          req.Email,
	}
	if err := s.upsertAccountDetails(ctx, details); err != nil {
		observability.LogCount(observability.MetricCustomerRegisterFailedCount, metricFields(req, "customer_register", "failed"))
		return status, err
	}

	status.CustomerRegistrationStatus = entity.StatusSucceeded
	status.CustomerRegistrationMessage = customerRegistrationSuccessMessage
	status.InterestDetailsStatus = entity.StatusInProgress
	status.UpdatedAt = time.Now().UTC()
	if err := s.statusStore.Update(ctx, status); err != nil {
		observability.LogCount(observability.MetricCustomerRegisterFailedCount, metricFields(req, "customer_register", "failed"))
		return status, err
	}

	observability.LogCount(observability.MetricCustomerRegisterSuccessCount, metricFields(req, "customer_register", "success"))
	return status, nil
}

func (s *service) fetchInterestDetails(ctx context.Context, req entity.OnboardingRequest, status entity.RequestStatus) (entity.RequestStatus, error) {
	resp, err := s.accountClient.GetInterestDetails(ctx, req.CustomerID, req.CorrelationID)
	if err != nil {
		observability.LogCount(observability.MetricInterestDetailsFailedCount, metricFields(req, "interest_details", "failed"))
		status.OverallStatus = entity.StatusFailed
		status.InterestDetailsStatus = entity.StatusFailed
		status.InterestDetailsMessage = err.Error()
		status.UpdatedAt = time.Now().UTC()
		if updateErr := s.statusStore.Update(ctx, status); updateErr != nil {
			return status, updateErr
		}
		return status, err
	}

	if err := s.applyInterestDetails(ctx, req.CustomerID, resp); err != nil {
		observability.LogCount(observability.MetricInterestDetailsFailedCount, metricFields(req, "interest_details", "failed"))
		return status, err
	}

	status.InterestDetailsStatus = entity.StatusSucceeded
	status.InterestDetailsMessage = interestDetailsSuccessMessage
	status.AccountOnboardingStatus = entity.StatusInProgress
	status.UpdatedAt = time.Now().UTC()
	if err := s.statusStore.Update(ctx, status); err != nil {
		observability.LogCount(observability.MetricInterestDetailsFailedCount, metricFields(req, "interest_details", "failed"))
		return status, err
	}

	observability.LogCount(observability.MetricInterestDetailsSuccessCount, metricFields(req, "interest_details", "success"))
	return status, nil
}

func (s *service) onboardAccount(ctx context.Context, req entity.OnboardingRequest, status entity.RequestStatus) (entity.RequestStatus, error) {
	details, err := s.accountDetailsStore.GetByCustomerID(ctx, req.CustomerID)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return status, err
		}
		details = entity.AccountDetails{
			CustomerID: req.CustomerID,
		}
	}

	details.AccountID = "ACC-" + req.CustomerID
	details.CardID = "CARD-" + req.CustomerID + "-001"
	details.CardType = req.CardType
	details.CardNumberMasked = maskCardNumber(req.CardNumber)
	if err := s.upsertAccountDetails(ctx, details); err != nil {
		return status, err
	}

	status.OverallStatus = entity.StatusSucceeded
	status.AccountOnboardingStatus = entity.StatusSucceeded
	status.AccountOnboardingMessage = accountOnboardingSuccessMessage
	status.UpdatedAt = time.Now().UTC()
	if err := s.statusStore.Update(ctx, status); err != nil {
		return status, err
	}

	return status, nil
}

func (s *service) applyInterestDetails(ctx context.Context, customerID string, resp accountapi.InterestDetailsResponse) error {
	details, err := s.accountDetailsStore.GetByCustomerID(ctx, customerID)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return err
		}
		details = entity.AccountDetails{
			CustomerID: customerID,
		}
	}

	details.ProductCode = resp.ProductCode
	details.InterestRate = resp.InterestRate
	details.InterestType = resp.InterestType
	details.Currency = resp.Currency

	return s.upsertAccountDetails(ctx, details)
}

func (s *service) upsertAccountDetails(ctx context.Context, details entity.AccountDetails) error {
	now := time.Now().UTC()
	existing, err := s.accountDetailsStore.GetByCustomerID(ctx, details.CustomerID)
	if err != nil {
		if !errors.Is(err, store.ErrNotFound) {
			return err
		}
		details.CreatedAt = now
		details.UpdatedAt = now
		return s.accountDetailsStore.Save(ctx, details)
	}

	if details.CoreCustomerID == "" {
		details.CoreCustomerID = existing.CoreCustomerID
	}
	if details.CustomerName == "" {
		details.CustomerName = existing.CustomerName
	}
	if details.Email == "" {
		details.Email = existing.Email
	}
	if details.ProductCode == "" {
		details.ProductCode = existing.ProductCode
	}
	if details.InterestRate == 0 {
		details.InterestRate = existing.InterestRate
	}
	if details.InterestType == "" {
		details.InterestType = existing.InterestType
	}
	if details.Currency == "" {
		details.Currency = existing.Currency
	}
	if details.AccountID == "" {
		details.AccountID = existing.AccountID
	}
	if details.CardID == "" {
		details.CardID = existing.CardID
	}
	if details.CardType == "" {
		details.CardType = existing.CardType
	}
	if details.CardNumberMasked == "" {
		details.CardNumberMasked = existing.CardNumberMasked
	}
	details.CreatedAt = existing.CreatedAt
	details.UpdatedAt = now

	return s.accountDetailsStore.Update(ctx, details)
}

func buildResponse(details entity.AccountDetails, status entity.RequestStatus) api.OnboardCardResponse {
	return api.OnboardCardResponse{
		CustomerId:     details.CustomerID,
		CoreCustomerId: details.CoreCustomerID,
		AccountId:      details.AccountID,
		CardId:         details.CardID,
		Status:         string(status.OverallStatus),
	}
}

func maskCardNumber(cardNumber string) string {
	if len(cardNumber) <= 4 {
		return cardNumber
	}

	return strings.Repeat("*", len(cardNumber)-4) + cardNumber[len(cardNumber)-4:]
}

func metricFields(req entity.OnboardingRequest, step string, status string) observability.Fields {
	fields := observability.NewFields()
	fields.CorrelationID = req.CorrelationID
	fields.JobID = req.JobID
	fields.RecordID = req.RecordID
	fields.CustomerID = req.CustomerID
	fields.SourceFile = req.SourceFile
	fields.RowNumber = int(req.RowNumber)
	fields.Step = step
	fields.Status = status
	return fields
}
