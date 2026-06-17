package main

import (
	"context"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/card"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/client"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/config"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/orchestration"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func main() {
	cfg := config.Load()

	var statusStore store.RequestStatusStore = store.NewInMemoryRequestStatusStore()
	var accountDetailsStore store.AccountDetailsStore = store.NewInMemoryAccountDetailsStore()

	if cfg.RequestStatusTableName != "" || cfg.AccountDetailsTableName != "" {
		awsConfig, err := awscfg.LoadDefaultConfig(context.Background())
		if err != nil {
			panic(err)
		}
		dynamoClient := dynamodb.NewFromConfig(awsConfig)

		if cfg.RequestStatusTableName != "" {
			statusStore = store.NewDynamoDBRequestStatusStore(dynamoClient, cfg.RequestStatusTableName)
		}
		if cfg.AccountDetailsTableName != "" {
			accountDetailsStore = store.NewDynamoDBAccountDetailsStore(dynamoClient, cfg.AccountDetailsTableName)
		}
	}

	customerClient, err := client.NewCustomerClient(client.CustomerClientConfig{
		BaseURL: cfg.CustomerManagementBaseURL,
		Timeout: cfg.DownstreamTimeout,
	})
	if err != nil {
		panic(err)
	}

	accountClient, err := client.NewAccountClient(client.AccountClientConfig{
		BaseURL: cfg.AccountManagementBaseURL,
		Timeout: cfg.DownstreamTimeout,
	})
	if err != nil {
		panic(err)
	}

	orchestrationService := orchestration.NewService(
		statusStore,
		accountDetailsStore,
		customerClient,
		accountClient,
	)
	cardService := card.NewService(orchestrationService, statusStore)

	router := card.NewRouter(cardService)
	if err := router.Run(cfg.ServerAddress); err != nil {
		panic(err)
	}
}
