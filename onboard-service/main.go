package main

import (
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/card"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/client"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/config"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/orchestration"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
)

func main() {
	cfg := config.Load()

	statusStore := store.NewInMemoryRequestStatusStore()
	accountDetailsStore := store.NewInMemoryAccountDetailsStore()

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
