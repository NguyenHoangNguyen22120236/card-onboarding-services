package card

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

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
	var req api.OnboardCardRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, api.ErrorResponse{
			Code:          errorCodeBadRequest,
			Message:       "invalid request body",
			CorrelationId: onboardCorrelationID(params, req),
		})
		return
	}

	resp, err := h.service.OnboardCard(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err, onboardCorrelationID(params, req))
		return
	}

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

func statusCorrelationID(params api.GetCardOnboardingStatusParams) string {
	if params.XCorrelationId != nil {
		return *params.XCorrelationId
	}

	return ""
}
