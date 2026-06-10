package customer

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
)

func TestHandler_GetHealth(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/health", nil)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp api.HealthResponse
	decodeResponse(t, recorder, &resp)
	if resp.Status != "ok" {
		t.Fatalf("Status = %q, want ok", resp.Status)
	}
}

func TestHandler_RegisterCustomer_Success(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	requestBody := validRegisterCustomerRequest()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/customers/register", jsonBody(t, requestBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp api.RegisterCustomerResponse
	decodeResponse(t, recorder, &resp)
	assertCustomerResponse(t, resp, requestBody.CustomerId)
}

func TestHandler_RegisterCustomer_InvalidJSONReturnsBadRequest(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/customers/register", bytes.NewBufferString("{"))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Correlation-Id", "corr-header")

	router.ServeHTTP(recorder, req)

	assertErrorResponse(t, recorder, http.StatusBadRequest, errorCodeBadRequest, "invalid request body", "corr-header")
}

func TestHandler_RegisterCustomer_MissingFieldsReturnsBadRequest(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	requestBody := validRegisterCustomerRequest()
	requestBody.CustomerId = ""
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/customers/register", jsonBody(t, requestBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	resp := assertErrorResponse(t, recorder, http.StatusBadRequest, errorCodeValidation, "request validation failed", requestBody.CorrelationId)
	if resp.Details == nil || len(*resp.Details) != 1 {
		t.Fatalf("Details = %#v, want one validation detail", resp.Details)
	}
	if (*resp.Details)[0] != (api.ErrorDetail{Field: "customerId", Issue: "customerId is required"}) {
		t.Fatalf("Details[0] = %#v, want customerId validation detail", (*resp.Details)[0])
	}
}

func TestHandler_RegisterCustomer_BadRequestSimulationReturnsBadRequest(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	requestBody := validRegisterCustomerRequest()
	requestBody.CustomerId = badRequestCustomerID
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/customers/register", jsonBody(t, requestBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	assertErrorResponse(t, recorder, http.StatusBadRequest, errorCodeBadRequest, "customer registration rejected", requestBody.CorrelationId)
}

func TestHandler_RegisterCustomer_InternalErrorSimulationReturnsInternalServerError(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	requestBody := validRegisterCustomerRequest()
	requestBody.CustomerId = failRegisterCustomerID
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/internal/customers/register", jsonBody(t, requestBody))
	req.Header.Set("Content-Type", "application/json")

	router.ServeHTTP(recorder, req)

	assertErrorResponse(t, recorder, http.StatusInternalServerError, errorCodeInternal, "customer registration failed", requestBody.CorrelationId)
}

func TestHandler_GetCustomer_Success(t *testing.T) {
	t.Parallel()

	router := newTestRouter()
	recorder := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/internal/customers/CUST-123", nil)

	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var resp api.RegisterCustomerResponse
	decodeResponse(t, recorder, &resp)
	assertCustomerResponse(t, resp, "CUST-123")
}

func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	api.RegisterHandlers(router, NewHandler(NewService()))
	return router
}

func jsonBody(t *testing.T, value any) *bytes.Buffer {
	t.Helper()

	var body bytes.Buffer
	if err := json.NewEncoder(&body).Encode(value); err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	return &body
}

func decodeResponse(t *testing.T, recorder *httptest.ResponseRecorder, dest any) {
	t.Helper()

	if err := json.Unmarshal(recorder.Body.Bytes(), dest); err != nil {
		t.Fatalf("Unmarshal() error = %v, body = %s", err, recorder.Body.String())
	}
}

func assertErrorResponse(t *testing.T, recorder *httptest.ResponseRecorder, statusCode int, code string, message string, correlationID string) api.ErrorResponse {
	t.Helper()

	if recorder.Code != statusCode {
		t.Fatalf("status = %d, want %d, body = %s", recorder.Code, statusCode, recorder.Body.String())
	}

	var resp api.ErrorResponse
	decodeResponse(t, recorder, &resp)
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
