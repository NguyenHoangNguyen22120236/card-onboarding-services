package orchestration

import (
	"context"
	"testing"
	"time"

	accountapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
	customerapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
)

func TestLocalE2ESimulation_OnboardServiceCompletesAllStepsSucceeded(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	customerClient := &fakeCustomerClient{
		resp: customerapi.RegisterCustomerResponse{
			CustomerId:     "CUST-E2E-001",
			CoreCustomerId: "CORE-CUST-E2E-001",
			Status:         "REGISTERED",
			RegisteredAt:   time.Now().UTC(),
		},
	}
	accountClient := &fakeAccountClient{
		resp: accountapi.InterestDetailsResponse{
			CustomerId:   "CUST-E2E-001",
			ProductCode:  "SAVINGS_BASIC",
			InterestRate: 4.5,
			InterestType: "VARIABLE",
			Currency:     "AUD",
		},
	}
	onboardService := NewService(statusStore, detailsStore, customerClient, accountClient)

	resp, err := onboardService.OnboardCard(context.Background(), entity.OnboardingRequest{
		CorrelationID: "local-job-001",
		JobID:         "local-job-001",
		RecordID:      "REC-001",
		SourceFile:    "e2e_cards.csv",
		RowNumber:     2,
		CustomerID:    "CUST-E2E-001",
		CardType:      "VISA",
		CardNumber:    "4111111111111111",
		ExpiryDate:    "12/29",
		HolderName:    "Alex Customer",
		Email:         "alex@example.com",
	})
	if err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}
	if resp.Status != string(entity.StatusSucceeded) {
		t.Fatalf("OnboardCard() status = %q, want SUCCEEDED", resp.Status)
	}
	if customerClient.calls != 1 {
		t.Fatalf("customer registration calls = %d, want 1", customerClient.calls)
	}
	if accountClient.calls != 1 {
		t.Fatalf("interest details calls = %d, want 1", accountClient.calls)
	}

	status, err := statusStore.GetByCustomerID(context.Background(), "CUST-E2E-001")
	if err != nil {
		t.Fatalf("GetByCustomerID() status error = %v", err)
	}
	if status.OverallStatus != entity.StatusSucceeded ||
		status.CustomerRegistrationStatus != entity.StatusSucceeded ||
		status.InterestDetailsStatus != entity.StatusSucceeded ||
		status.AccountOnboardingStatus != entity.StatusSucceeded {
		t.Fatalf("status = %#v, want all onboarding steps SUCCEEDED", status)
	}

	details, err := detailsStore.GetByCustomerID(context.Background(), "CUST-E2E-001")
	if err != nil {
		t.Fatalf("GetByCustomerID() details error = %v", err)
	}
	if details.CoreCustomerID != "CORE-CUST-E2E-001" ||
		details.ProductCode != "SAVINGS_BASIC" ||
		details.AccountID != "ACC-CUST-E2E-001" ||
		details.CardID != "CARD-CUST-E2E-001-001" ||
		details.CardType != "VISA" {
		t.Fatalf("details = %#v, want completed customer, interest, and account onboarding details", details)
	}
}
