package card

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/client"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/pkg/onboard"
)

func TestHandlerGetHealth(t *testing.T) {
	router := newCardTestRouter(&fakeCardService{})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp api.HealthResponse
	decodeHTTPResponse(t, recorder, &resp)
	if resp.Status != "ok" {
		t.Fatalf("Status = %q, want ok", resp.Status)
	}
}

func TestHandlerOnboardCardSuccess(t *testing.T) {
	service := fakeCardService{
		onboardResp: api.OnboardCardResponse{
			CustomerId:     "CUST001",
			CoreCustomerId: "CORE-CUST001",
			AccountId:      "ACC-CUST001",
			CardId:         "CARD-CUST001-001",
			Status:         "SUCCEEDED",
		},
	}
	router := newCardTestRouter(&service)
	recorder := httptest.NewRecorder()
	body := validOnboardCardRequest()
	req := httptest.NewRequest(http.MethodPost, "/internal/cards/onboard", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if service.onboardCalls != 1 {
		t.Fatalf("onboardCalls = %d, want 1", service.onboardCalls)
	}
	if service.lastOnboardRequest.CustomerId != body.CustomerId {
		t.Fatalf("CustomerId = %q, want %q", service.lastOnboardRequest.CustomerId, body.CustomerId)
	}

	var resp api.OnboardCardResponse
	decodeHTTPResponse(t, recorder, &resp)
	if resp != service.onboardResp {
		t.Fatalf("response = %#v, want %#v", resp, service.onboardResp)
	}
}

func TestHandlerOnboardCardInvalidJSONReturnsBadRequest(t *testing.T) {
	router := newCardTestRouter(&fakeCardService{})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/cards/onboard", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-Id", "corr-header")

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusBadRequest, errorCodeBadRequest, "invalid request body", "corr-header")
}

func TestHandlerOnboardCardMissingFieldsReturnsBadRequest(t *testing.T) {
	router := newCardTestRouter(NewService(fakeOrchestration{}, store.NewInMemoryRequestStatusStore()))
	body := validOnboardCardRequest()
	body.CustomerId = ""
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/cards/onboard", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	resp := assertHTTPErrorResponse(t, recorder, http.StatusBadRequest, errorCodeValidation, "request validation failed", body.CorrelationId)
	if resp.Details == nil || len(*resp.Details) != 1 {
		t.Fatalf("Details = %#v, want one validation detail", resp.Details)
	}
	if (*resp.Details)[0] != (api.ErrorDetail{Field: "customerId", Issue: "customerId is required"}) {
		t.Fatalf("Details[0] = %#v, want customerId validation detail", (*resp.Details)[0])
	}
}

func TestHandlerOnboardCardInternalErrorReturnsInternalServerError(t *testing.T) {
	router := newCardTestRouter(NewService(fakeOrchestration{err: client.ErrDownstreamInternal}, store.NewInMemoryRequestStatusStore()))
	body := validOnboardCardRequest()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/cards/onboard", jsonBody(t, body))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusInternalServerError, errorCodeInternal, "card onboarding failed", body.CorrelationId)
}

func TestServiceOnboardCardCallsOrchestration(t *testing.T) {
	orchestration := &recordingOrchestration{
		resp: api.OnboardCardResponse{
			CustomerId:     "CUST001",
			CoreCustomerId: "CORE-CUST001",
			AccountId:      "ACC-CUST001",
			CardId:         "CARD-CUST001-001",
			Status:         "SUCCEEDED",
		},
	}
	service := NewService(orchestration, store.NewInMemoryRequestStatusStore())
	req := validOnboardCardRequest()

	resp, err := service.OnboardCard(context.Background(), req)
	if err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}

	if orchestration.calls != 1 {
		t.Fatalf("orchestration calls = %d, want 1", orchestration.calls)
	}
	if orchestration.lastRequest.CustomerID != req.CustomerId ||
		orchestration.lastRequest.CorrelationID != req.CorrelationId ||
		orchestration.lastRequest.Email != string(req.Email) ||
		orchestration.lastRequest.CardNumber != req.CardNumber {
		t.Fatalf("orchestration request = %#v, want fields converted from API request", orchestration.lastRequest)
	}
	if resp != orchestration.resp {
		t.Fatalf("response = %#v, want %#v", resp, orchestration.resp)
	}
}

func TestHandlerGetCardOnboardingStatusSuccess(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	status := entity.RequestStatus{
		CustomerID:                 "CUST001",
		OverallStatus:              entity.StatusSucceeded,
		CustomerRegistrationStatus: entity.StatusSucceeded,
		InterestDetailsStatus:      entity.StatusSucceeded,
		AccountOnboardingStatus:    entity.StatusSucceeded,
		CreatedAt:                  time.Now().UTC(),
		UpdatedAt:                  time.Now().UTC(),
	}
	if err := statusStore.Save(context.Background(), status); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	router := newCardTestRouter(NewService(fakeOrchestration{}, statusStore))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/cards/CUST001/status", nil)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}

	var resp api.CardOnboardingStatusResponse
	decodeHTTPResponse(t, recorder, &resp)
	if resp.CustomerId != status.CustomerID ||
		resp.OverallStatus != string(status.OverallStatus) ||
		resp.CustomerRegistrationStatus != string(status.CustomerRegistrationStatus) ||
		resp.InterestDetailsStatus != string(status.InterestDetailsStatus) ||
		resp.AccountOnboardingStatus != string(status.AccountOnboardingStatus) {
		t.Fatalf("response = %#v, want status fields from store", resp)
	}
}

func TestHandlerGetCardOnboardingStatusMissingReturnsNotFound(t *testing.T) {
	router := newCardTestRouter(NewService(fakeOrchestration{}, store.NewInMemoryRequestStatusStore()))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/cards/CUST_MISSING/status", nil)
	req.Header.Set("X-Correlation-Id", "corr-status")

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusNotFound, errorCodeNotFound, "card onboarding status not found", "corr-status")
}

func TestHandlerGetCardOnboardingStatusInternalErrorReturnsInternalServerError(t *testing.T) {
	router := newCardTestRouter(NewService(fakeOrchestration{}, failingStatusStore{err: errors.New("database offline")}))
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/cards/CUST001/status", nil)

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusInternalServerError, errorCodeInternal, "internal service error", "")
}

func newCardTestRouter(service CardService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api.RegisterHandlers(router, NewHandler(service))
	return router
}

func validOnboardCardRequest() api.OnboardCardRequest {
	return api.OnboardCardRequest{
		CorrelationId: "corr-123",
		JobId:         "job-123",
		RecordId:      "rec-123",
		SourceFile:    "cards.csv",
		RowNumber:     2,
		CustomerId:    "CUST001",
		CardType:      "DEBIT",
		CardNumber:    "4111111111111111",
		ExpiryDate:    "12/28",
		HolderName:    "Nguyen Van A",
		Email:         "a@example.com",
	}
}

func jsonBody(t *testing.T, value any) *bytes.Buffer {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	return &body
}

func decodeHTTPResponse(t *testing.T, recorder *httptest.ResponseRecorder, dest any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), dest); err != nil {
		t.Fatalf("Unmarshal() error = %v, body = %s", err, recorder.Body.String())
	}
}

func assertHTTPErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, statusCode int, code string, message string, correlationID string) api.ErrorResponse {
	t.Helper()

	if recorder.Code != statusCode {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, statusCode, recorder.Body.String())
	}

	var resp api.ErrorResponse
	decodeHTTPResponse(t, recorder, &resp)
	if resp.Code != code {
		t.Fatalf("Code = %q, want %q", resp.Code, code)
	}
	if resp.Message != message {
		t.Fatalf("Message = %q, want %q", resp.Message, message)
	}
	if resp.CorrelationId != correlationID {
		t.Fatalf("CorrelationId = %q, want %q", resp.CorrelationId, correlationID)
	}

	return resp
}

type fakeCardService struct {
	onboardResp        api.OnboardCardResponse
	onboardErr         error
	statusResp         api.CardOnboardingStatusResponse
	statusErr          error
	onboardCalls       int
	lastOnboardRequest api.OnboardCardRequest
}

func (s *fakeCardService) OnboardCard(_ context.Context, req api.OnboardCardRequest) (api.OnboardCardResponse, error) {
	s.onboardCalls++
	s.lastOnboardRequest = req
	return s.onboardResp, s.onboardErr
}

func (s *fakeCardService) GetCardOnboardingStatus(_ context.Context, _ string) (api.CardOnboardingStatusResponse, error) {
	return s.statusResp, s.statusErr
}

type fakeOrchestration struct {
	resp api.OnboardCardResponse
	err  error
	req  entity.OnboardingRequest
}

func (s fakeOrchestration) OnboardCard(_ context.Context, req entity.OnboardingRequest) (api.OnboardCardResponse, error) {
	s.req = req
	return s.resp, s.err
}

type recordingOrchestration struct {
	resp        api.OnboardCardResponse
	err         error
	calls       int
	lastRequest entity.OnboardingRequest
}

func (s *recordingOrchestration) OnboardCard(_ context.Context, req entity.OnboardingRequest) (api.OnboardCardResponse, error) {
	s.calls++
	s.lastRequest = req
	return s.resp, s.err
}

type failingStatusStore struct {
	err error
}

func (s failingStatusStore) GetByCustomerID(context.Context, string) (entity.RequestStatus, error) {
	return entity.RequestStatus{}, s.err
}

func (s failingStatusStore) Save(context.Context, entity.RequestStatus) error {
	return nil
}

func (s failingStatusStore) Update(context.Context, entity.RequestStatus) error {
	return nil
}
