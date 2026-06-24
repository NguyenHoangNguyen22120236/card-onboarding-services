package card

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/observability"
	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/pkg/onboard"
)

var _ api.ServerInterface = (*Handler)(nil)

type Handler struct {
	service CardService
}

func NewHandler(service CardService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, api.HealthResponse{
		Status: "ok",
	})
}

func (h *Handler) OnboardCard(c *gin.Context, params api.OnboardCardParams) {
	start := time.Now()
	fields := observability.NewFields()
	if params.XCorrelationId != nil {
		fields.CorrelationID = *params.XCorrelationId
	}
	observability.LogCount(observability.MetricRequestCount, fields)
	defer func() {
		observability.LogDuration(observability.MetricResponseTimeMilliseconds, time.Since(start), fields)
	}()

	var req api.OnboardCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		observability.LogCount(observability.MetricFailedCount, fields)
		writeError(c, http.StatusBadRequest, api.ErrorResponse{
			Code:          errorCodeBadRequest,
			Message:       "invalid request body",
			CorrelationId: onboardCorrelationID(params, req),
		})
		return
	}

	fields = onboardMetricFields(params, req)
	resp, err := h.service.OnboardCard(c.Request.Context(), req)
	if err != nil {
		observability.LogCount(observability.MetricFailedCount, fields)
		writeServiceError(c, err, onboardCorrelationID(params, req))
		return
	}

	observability.LogCount(observability.MetricSuccessCount, fields)
	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetCardOnboardingStatus(c *gin.Context, customerId string, params api.GetCardOnboardingStatusParams) {
	resp, err := h.service.GetCardOnboardingStatus(c.Request.Context(), customerId)
	if err != nil {
		writeServiceError(c, err, statusCorrelationID(params))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func writeServiceError(c *gin.Context, err error, correlationID string) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		statusCode := http.StatusInternalServerError
		switch {
		case errors.Is(err, ErrBadRequest):
			statusCode = http.StatusBadRequest
		case errors.Is(err, ErrNotFound):
			statusCode = http.StatusNotFound
		}

		resp := api.ErrorResponse{
			Code:          serviceErr.Code,
			Message:       serviceErr.Message,
			CorrelationId: correlationID,
		}
		if len(serviceErr.Details) > 0 {
			details := serviceErr.Details
			resp.Details = &details
		}

		writeError(c, statusCode, resp)
		return
	}

	writeError(c, http.StatusInternalServerError, api.ErrorResponse{
		Code:          errorCodeInternal,
		Message:       "internal service error",
		CorrelationId: correlationID,
	})
}

func writeError(c *gin.Context, statusCode int, resp api.ErrorResponse) {
	if resp.Code == "" {
		resp.Code = errorCodeInternal
	}
	if resp.Message == "" {
		resp.Message = "internal service error"
	}

	c.JSON(statusCode, resp)
}

func onboardCorrelationID(params api.OnboardCardParams, req api.OnboardCardRequest) string {
	if params.XCorrelationId != nil {
		return *params.XCorrelationId
	}

	return req.CorrelationId
}

func onboardMetricFields(params api.OnboardCardParams, req api.OnboardCardRequest) observability.Fields {
	fields := observability.NewFields()
	fields.CorrelationID = onboardCorrelationID(params, req)
	fields.JobID = req.JobId
	fields.RecordID = req.RecordId
	fields.CustomerID = req.CustomerId
	fields.SourceFile = req.SourceFile
	fields.RowNumber = int(req.RowNumber)
	return fields
}

func statusCorrelationID(params api.GetCardOnboardingStatusParams) string {
	if params.XCorrelationId != nil {
		return *params.XCorrelationId
	}

	return ""
}
