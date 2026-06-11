package account

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
)

var _ api.ServerInterface = (*Handler)(nil)

type Handler struct {
	service AccountService
}

func NewHandler(service AccountService) *Handler {
	if service == nil {
		service = NewService()
	}

	return &Handler{
		service: service,
	}
}

func (h *Handler) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, api.HealthResponse{
		Status: "ok",
	})
}

func (h *Handler) GetInterestDetails(c *gin.Context, customerId string, params api.GetInterestDetailsParams) {
	resp, err := h.service.GetInterestDetails(customerId)
	if err != nil {
		writeServiceError(c, err, interestDetailsCorrelationID(params))
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

func interestDetailsCorrelationID(params api.GetInterestDetailsParams) string {
	if params.XCorrelationId != nil {
		return *params.XCorrelationId
	}

	return ""
}
