package client

import (
	"context"
	"fmt"
	"net/http"
	"time"

	accountapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
)

type AccountClient interface {
	GetInterestDetails(ctx context.Context, customerID string, correlationID string) (accountapi.InterestDetailsResponse, error)
}

type AccountClientConfig struct {
	BaseURL string
	Timeout time.Duration
}

type generatedAccountClient struct {
	client  accountapi.ClientWithResponsesInterface
	timeout time.Duration
}

func NewAccountClient(config AccountClientConfig) (AccountClient, error) {
	generatedClient, err := accountapi.NewClientWithResponses(config.BaseURL)
	if err != nil {
		return nil, err
	}

	return &generatedAccountClient{
		client:  generatedClient,
		timeout: config.Timeout,
	}, nil
}

func (c *generatedAccountClient) GetInterestDetails(ctx context.Context, customerID string, correlationID string) (accountapi.InterestDetailsResponse, error) {
	callCtx, cancel := contextWithTimeout(ctx, c.timeout)
	defer cancel()

	header := accountapi.CorrelationIdHeader(correlationID)
	params := &accountapi.GetInterestDetailsParams{
		XCorrelationId: &header,
	}

	response, err := c.client.GetInterestDetailsWithResponse(
		callCtx,
		customerID,
		params,
		accountCorrelationHeader(correlationID),
	)
	if err != nil {
		return accountapi.InterestDetailsResponse{}, mapDownstreamCallError(callCtx, err)
	}

	switch response.StatusCode() {
	case http.StatusBadRequest:
		return accountapi.InterestDetailsResponse{}, ErrDownstreamBadRequest
	case http.StatusNotFound:
		return accountapi.InterestDetailsResponse{}, ErrDownstreamNotFound
	case http.StatusInternalServerError:
		return accountapi.InterestDetailsResponse{}, ErrDownstreamInternal
	}

	if response.StatusCode() >= http.StatusOK && response.StatusCode() < http.StatusMultipleChoices {
		if response.JSON200 == nil {
			return accountapi.InterestDetailsResponse{}, fmt.Errorf("%w: empty interest details response", ErrDownstreamInternal)
		}
		return *response.JSON200, nil
	}

	return accountapi.InterestDetailsResponse{}, fmt.Errorf("%w: unexpected interest details status %d", ErrDownstreamInternal, response.StatusCode())
}

func accountCorrelationHeader(correlationID string) accountapi.RequestEditorFn {
	return func(ctx context.Context, req *http.Request) error {
		if correlationID != "" {
			req.Header.Set("X-Correlation-Id", correlationID)
		}
		return nil
	}
}
