package account

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
)

func TestHandlerGetHealth(t *testing.T) {
	router := newAccountTestRouter(NewService())
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

func TestHandlerGetInterestDetailsSuccess(t *testing.T) {
	router := newAccountTestRouter(NewService())
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/accounts/CUST123/interest-details", nil)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp api.InterestDetailsResponse
	decodeHTTPResponse(t, recorder, &resp)
	if resp.CustomerId != "CUST123" {
		t.Fatalf("CustomerId = %q, want %q", resp.CustomerId, "CUST123")
	}
	if resp.ProductCode != "SAVINGS_BASIC" {
		t.Fatalf("ProductCode = %q, want %q", resp.ProductCode, "SAVINGS_BASIC")
	}
	if resp.InterestRate != 4.5 {
		t.Fatalf("InterestRate = %v, want %v", resp.InterestRate, 4.5)
	}
	if resp.InterestType != "VARIABLE" {
		t.Fatalf("InterestType = %q, want %q", resp.InterestType, "VARIABLE")
	}
	if resp.Currency != "AUD" {
		t.Fatalf("Currency = %q, want %q", resp.Currency, "AUD")
	}
}

func TestHandlerGetInterestDetailsNotFoundReturnsNotFound(t *testing.T) {
	router := newAccountTestRouter(NewService())
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/accounts/"+noInterestCustomerID+"/interest-details", nil)

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusNotFound, errorCodeNotFound, "interest details not found", "")
}

func TestHandlerGetInterestDetailsInternalErrorReturnsInternalServerError(t *testing.T) {
	router := newAccountTestRouter(NewService())
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/accounts/"+failInterestCustomerID+"/interest-details", nil)
	req.Header.Set("X-Correlation-Id", "corr-interest")

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusInternalServerError, errorCodeInternal, "interest details lookup failed", "corr-interest")
}

func TestHandlerGetInterestDetailsValidationErrorReturnsBadRequest(t *testing.T) {
	serviceErr := &ServiceError{
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
	router := newAccountTestRouter(fakeAccountService{err: serviceErr})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/accounts/CUST123/interest-details", nil)
	req.Header.Set("X-Correlation-Id", "corr-validation")

	router.ServeHTTP(recorder, req)

	resp := assertHTTPErrorResponse(t, recorder, http.StatusBadRequest, errorCodeValidation, "request validation failed", "corr-validation")
	if resp.Details == nil || len(*resp.Details) != 1 {
		t.Fatalf("Details = %#v, want one validation detail", resp.Details)
	}
	if (*resp.Details)[0] != (api.ErrorDetail{Field: "customerId", Issue: "customerId is required"}) {
		t.Fatalf("Details[0] = %#v, want customerId validation detail", (*resp.Details)[0])
	}
}

func TestHandlerGetInterestDetailsUnknownErrorReturnsInternalServerError(t *testing.T) {
	router := newAccountTestRouter(fakeAccountService{err: errors.New("database offline")})
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/accounts/CUST123/interest-details", nil)

	router.ServeHTTP(recorder, req)

	assertHTTPErrorResponse(t, recorder, http.StatusInternalServerError, errorCodeInternal, "internal service error", "")
}

func newAccountTestRouter(service AccountService) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api.RegisterHandlers(router, NewHandler(service))
	return router
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

type fakeAccountService struct {
	resp api.InterestDetailsResponse
	err  error
}

func (s fakeAccountService) GetInterestDetails(string) (api.InterestDetailsResponse, error) {
	return s.resp, s.err
}
