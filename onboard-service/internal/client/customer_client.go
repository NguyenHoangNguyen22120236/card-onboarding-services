package client

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	customerapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
)

var (
	ErrDownstreamBadRequest = errors.New("downstream bad request")
	ErrDownstreamNotFound   = errors.New("downstream not found")
	ErrDownstreamInternal   = errors.New("downstream internal error")
	ErrDownstreamTimeout    = errors.New("downstream timeout")
)

type CustomerClient interface {
	RegisterCustomer(ctx context.Context, request customerapi.RegisterCustomerRequest) (customerapi.RegisterCustomerResponse, error)
}

type CustomerClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

type generatedCustomerClient struct {
	client  customerapi.ClientWithResponsesInterface
	timeout time.Duration
}

func NewCustomerClient(config CustomerClientConfig) (CustomerClient, error) {
	generatedClient, err := customerapi.NewClientWithResponses(config.BaseURL)
	if err != nil {
		return nil, err
	}

	return &generatedCustomerClient{
		client:  generatedClient,
		timeout: config.Timeout,
	}, nil
}

func (c *generatedCustomerClient) RegisterCustomer(ctx context.Context, request customerapi.RegisterCustomerRequest) (customerapi.RegisterCustomerResponse, error) {
	callCtx, cancel := contextWithTimeout(ctx, c.timeout)
	defer cancel()

	correlationID := customerapi.CorrelationIdHeader(request.CorrelationId)
	params := &customerapi.RegisterCustomerParams{
		XCorrelationId: &correlationID,
	}

	response, err := c.client.RegisterCustomerWithResponse(
		callCtx,
		params,
		customerapi.RegisterCustomerJSONRequestBody(request),
		customerCorrelationHeader(request.CorrelationId),
	)
	if err != nil {
		return customerapi.RegisterCustomerResponse{}, mapDownstreamCallError(callCtx, err)
	}

	switch response.StatusCode() {
	case http.StatusBadRequest:
		return customerapi.RegisterCustomerResponse{}, ErrDownstreamBadRequest
	case http.StatusInternalServerError:
		return customerapi.RegisterCustomerResponse{}, ErrDownstreamInternal
	}

	if response.StatusCode() >= http.StatusOK && response.StatusCode() < http.StatusMultipleChoices {
		if response.JSON200 == nil {
			return customerapi.RegisterCustomerResponse{}, fmt.Errorf("%w: empty customer registration response", ErrDownstreamInternal)
		}
		return *response.JSON200, nil
	}

	return customerapi.RegisterCustomerResponse{}, fmt.Errorf("%w: unexpected customer registration status %d", ErrDownstreamInternal, response.StatusCode())
}

func customerCorrelationHeader(correlationID string) customerapi.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if correlationID != "" {
			req.Header.Set("X-Correlation-Id", correlationID)
		}
		return nil
	}
}

func contextWithTimeout(ctx context.Context, timeout time.Duration) (context.Context, context.CancelFunc) {
	if timeout <= 0 {
		return ctx, func() {}
	}
	return context.WithTimeout(ctx, timeout)
}

func mapDownstreamCallError(ctx context.Context, err error) error {
	if errors.Is(ctx.Err(), context.DeadlineExceeded) || errors.Is(err, context.DeadlineExceeded) {
		return ErrDownstreamTimeout
	}
	return fmt.Errorf("%w: %v", ErrDownstreamInternal, err)
}
