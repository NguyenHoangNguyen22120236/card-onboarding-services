package customer

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	api "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
)

var _ api.ServerInterface = (*Handler)(nil)

type Handler struct {
	service CustomerService
}

func NewHandler(service CustomerService) *Handler {
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

func (h *Handler) RegisterCustomer(c *gin.Context, params api.RegisterCustomerParams) {
	var req api.RegisterCustomerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		writeError(c, http.StatusBadRequest, api.ErrorResponse{
			Code:          errorCodeBadRequest,
			Message:       "invalid request body",
			CorrelationId: registerCorrelationID(params, req),
		})
		return
	}

	resp, err := h.service.RegisterCustomer(c.Request.Context(), req)
	if err != nil {
		writeServiceError(c, err, registerCorrelationID(params, req))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetCustomer(c *gin.Context, customerId string, params api.GetCustomerParams) {
	resp, err := h.service.GetCustomer(c.Request.Context(), customerId)
	if err != nil {
		writeServiceError(c, err, getCustomerCorrelationID(params))
		return
	}

	c.JSON(http.StatusOK, resp)
}

func writeServiceError(c *gin.Context, err error, correlationID string) {
	var serviceErr *ServiceError
	if errors.As(err, &serviceErr) {
		statusCode := http.StatusInternalServerError
		if errors.Is(err, ErrBadRequest) {
			statusCode = http.StatusBadRequest
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

func registerCorrelationID(params api.RegisterCustomerParams, req api.RegisterCustomerRequest) string {
	if params.XCorrelationId != nil {
		return *params.XCorrelationId
	}

	return req.CorrelationId
}

func getCustomerCorrelationID(params api.GetCustomerParams) string {
	if params.XCorrelationId != nil {
		return *params.XCorrelationId
	}

	return ""
}
