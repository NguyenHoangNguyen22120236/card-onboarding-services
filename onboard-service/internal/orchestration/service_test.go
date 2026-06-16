package orchestration

import (
	"context"
	"errors"
	"testing"
	"time"

	accountapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/account-management-service/pkg/account"
	customerapi "github.com/NguyenHoangNguyen22120236/card-onboarding-services/customer-management-service/pkg/customer"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/store"
)

func TestOnboardCardSuccess(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	customerClient := &fakeCustomerClient{
		resp: customerapi.RegisterCustomerResponse{
			CustomerId:     "CUST001",
			CoreCustomerId: "CORE-CUST001",
			Status:         "REGISTERED",
			RegisteredAt:   time.Now().UTC(),
		},
	}
	accountClient := &fakeAccountClient{
		resp: accountapi.InterestDetailsResponse{
			CustomerId:   "CUST001",
			ProductCode:  "SAVINGS_BASIC",
			InterestRate: 4.5,
			InterestType: "VARIABLE",
			Currency:     "AUD",
		},
	}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)

	resp, err := service.OnboardCard(context.Background(), testRequest("CUST001"))
	if err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}

	if resp.CustomerId != "CUST001" || resp.CoreCustomerId != "CORE-CUST001" ||
		resp.AccountId != "ACC-CUST001" || resp.CardId != "CARD-CUST001-001" ||
		resp.Status != string(entity.StatusSucceeded) {
		t.Fatalf("OnboardCard() response = %#v, want completed response", resp)
	}
	if customerClient.calls != 1 {
		t.Fatalf("customer calls = %d, want 1", customerClient.calls)
	}
	if accountClient.calls != 1 {
		t.Fatalf("account calls = %d, want 1", accountClient.calls)
	}

	status, err := statusStore.GetByCustomerID(context.Background(), "CUST001")
	if err != nil {
		t.Fatalf("GetByCustomerID() status error = %v", err)
	}
	if status.OverallStatus != entity.StatusSucceeded ||
		status.CustomerRegistrationStatus != entity.StatusSucceeded ||
		status.CustomerRegistrationMessage != customerRegistrationSuccessMessage ||
		status.InterestDetailsStatus != entity.StatusSucceeded ||
		status.InterestDetailsMessage != interestDetailsSuccessMessage ||
		status.AccountOnboardingStatus != entity.StatusSucceeded ||
		status.AccountOnboardingMessage != accountOnboardingSuccessMessage {
		t.Fatalf("status = %#v, want all steps succeeded", status)
	}

	details, err := detailsStore.GetByCustomerID(context.Background(), "CUST001")
	if err != nil {
		t.Fatalf("GetByCustomerID() details error = %v", err)
	}
	if details.CoreCustomerID != "CORE-CUST001" ||
		details.CustomerName != "Alex Customer" ||
		details.Email != "alex@example.com" ||
		details.ProductCode != "SAVINGS_BASIC" ||
		details.InterestRate != 4.5 ||
		details.InterestType != "VARIABLE" ||
		details.Currency != "AUD" ||
		details.AccountID != "ACC-CUST001" ||
		details.CardID != "CARD-CUST001-001" ||
		details.CardType != "DEBIT" ||
		details.CardNumberMasked != "************1111" {
		t.Fatalf("details = %#v, want persisted onboarding details", details)
	}
}

func TestOnboardCardStoresMaskedRequestCardNumber(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	customerClient := &fakeCustomerClient{
		resp: customerapi.RegisterCustomerResponse{
			CustomerId:     "CUST_MASK",
			CoreCustomerId: "CORE-CUST_MASK",
			Status:         "REGISTERED",
			RegisteredAt:   time.Now().UTC(),
		},
	}
	accountClient := &fakeAccountClient{
		resp: accountapi.InterestDetailsResponse{
			CustomerId:   "CUST_MASK",
			ProductCode:  "SAVINGS_BASIC",
			InterestRate: 4.5,
			InterestType: "VARIABLE",
			Currency:     "AUD",
		},
	}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)
	req := testRequest("CUST_MASK")
	req.CardNumber = "5555444433332222"

	if _, err := service.OnboardCard(context.Background(), req); err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}

	details, err := detailsStore.GetByCustomerID(context.Background(), "CUST_MASK")
	if err != nil {
		t.Fatalf("GetByCustomerID() details error = %v", err)
	}
	if details.CardNumberMasked != "************2222" {
		t.Fatalf("CardNumberMasked = %q, want %q", details.CardNumberMasked, "************2222")
	}
}

func TestMaskCardNumber(t *testing.T) {
	tests := []struct {
		name       string
		cardNumber string
		want       string
	}{
		{
			name:       "empty",
			cardNumber: "",
			want:       "",
		},
		{
			name:       "shorter than four",
			cardNumber: "123",
			want:       "123",
		},
		{
			name:       "exactly four",
			cardNumber: "1234",
			want:       "1234",
		},
		{
			name:       "longer than four",
			cardNumber: "12345",
			want:       "*2345",
		},
		{
			name:       "standard card number",
			cardNumber: "4111111111111111",
			want:       "************1111",
		},
		{
			name:       "different last four",
			cardNumber: "5555444433332222",
			want:       "************2222",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := maskCardNumber(tt.cardNumber); got != tt.want {
				t.Fatalf("maskCardNumber(%q) = %q, want %q", tt.cardNumber, got, tt.want)
			}
		})
	}
}

func TestOnboardCardCustomerRegistrationFailure(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	customerErr := errors.New("customer registration failed")
	customerClient := &fakeCustomerClient{err: customerErr}
	accountClient := &fakeAccountClient{}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)

	_, err := service.OnboardCard(context.Background(), testRequest("CUST_FAIL_REGISTER"))
	if !errors.Is(err, customerErr) {
		t.Fatalf("OnboardCard() error = %v, want customer error", err)
	}
	if customerClient.calls != 1 {
		t.Fatalf("customer calls = %d, want 1", customerClient.calls)
	}
	if accountClient.calls != 0 {
		t.Fatalf("account calls = %d, want 0", accountClient.calls)
	}

	status, err := statusStore.GetByCustomerID(context.Background(), "CUST_FAIL_REGISTER")
	if err != nil {
		t.Fatalf("GetByCustomerID() status error = %v", err)
	}
	if status.OverallStatus != entity.StatusFailed ||
		status.CustomerRegistrationStatus != entity.StatusFailed ||
		status.CustomerRegistrationMessage != customerErr.Error() {
		t.Fatalf("status = %#v, want customer registration failed", status)
	}
}

func TestOnboardCardInterestDetailsFailure(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	customerClient := &fakeCustomerClient{
		resp: customerapi.RegisterCustomerResponse{
			CustomerId:     "CUST_FAIL_INTEREST",
			CoreCustomerId: "CORE-CUST_FAIL_INTEREST",
			Status:         "REGISTERED",
			RegisteredAt:   time.Now().UTC(),
		},
	}
	interestErr := errors.New("interest details lookup failed")
	accountClient := &fakeAccountClient{err: interestErr}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)

	_, err := service.OnboardCard(context.Background(), testRequest("CUST_FAIL_INTEREST"))
	if !errors.Is(err, interestErr) {
		t.Fatalf("OnboardCard() error = %v, want interest error", err)
	}
	if customerClient.calls != 1 {
		t.Fatalf("customer calls = %d, want 1", customerClient.calls)
	}
	if accountClient.calls != 1 {
		t.Fatalf("account calls = %d, want 1", accountClient.calls)
	}

	status, err := statusStore.GetByCustomerID(context.Background(), "CUST_FAIL_INTEREST")
	if err != nil {
		t.Fatalf("GetByCustomerID() status error = %v", err)
	}
	if status.OverallStatus != entity.StatusFailed ||
		status.CustomerRegistrationStatus != entity.StatusSucceeded ||
		status.InterestDetailsStatus != entity.StatusFailed ||
		status.InterestDetailsMessage != interestErr.Error() ||
		status.AccountOnboardingStatus != "" {
		t.Fatalf("status = %#v, want interest details failed without account onboarding", status)
	}
}

func TestOnboardCardResumeFromCustomerRegistrationSucceeded(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	now := time.Now().UTC()
	if err := statusStore.Save(context.Background(), entity.RequestStatus{
		CustomerID:                  "CUST_RESUME_CUSTOMER",
		OverallStatus:               entity.StatusFailed,
		CustomerRegistrationStatus:  entity.StatusSucceeded,
		CustomerRegistrationMessage: customerRegistrationSuccessMessage,
		InterestDetailsStatus:       entity.StatusFailed,
		InterestDetailsMessage:      "previous interest failure",
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}); err != nil {
		t.Fatalf("Save() status error = %v", err)
	}
	if err := detailsStore.Save(context.Background(), entity.AccountDetails{
		CustomerID:     "CUST_RESUME_CUSTOMER",
		CoreCustomerID: "CORE-CUST_RESUME_CUSTOMER",
		CustomerName:   "Alex Customer",
		Email:          "alex@example.com",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("Save() details error = %v", err)
	}
	customerClient := &fakeCustomerClient{}
	accountClient := &fakeAccountClient{
		resp: accountapi.InterestDetailsResponse{
			CustomerId:   "CUST_RESUME_CUSTOMER",
			ProductCode:  "CARD_GOLD",
			InterestRate: 5.5,
			InterestType: "FIXED",
			Currency:     "USD",
		},
	}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)

	resp, err := service.OnboardCard(context.Background(), testRequest("CUST_RESUME_CUSTOMER"))
	if err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}
	if resp.Status != string(entity.StatusSucceeded) {
		t.Fatalf("OnboardCard() status = %q, want SUCCEEDED", resp.Status)
	}
	if customerClient.calls != 0 {
		t.Fatalf("customer calls = %d, want 0", customerClient.calls)
	}
	if accountClient.calls != 1 {
		t.Fatalf("account calls = %d, want 1", accountClient.calls)
	}
}

func TestOnboardCardResumeFromInterestDetailsSucceeded(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	now := time.Now().UTC()
	if err := statusStore.Save(context.Background(), entity.RequestStatus{
		CustomerID:                  "CUST_RESUME_INTEREST",
		OverallStatus:               entity.StatusFailed,
		CustomerRegistrationStatus:  entity.StatusSucceeded,
		CustomerRegistrationMessage: customerRegistrationSuccessMessage,
		InterestDetailsStatus:       entity.StatusSucceeded,
		InterestDetailsMessage:      interestDetailsSuccessMessage,
		AccountOnboardingStatus:     entity.StatusFailed,
		AccountOnboardingMessage:    "previous account failure",
		CreatedAt:                   now,
		UpdatedAt:                   now,
	}); err != nil {
		t.Fatalf("Save() status error = %v", err)
	}
	if err := detailsStore.Save(context.Background(), entity.AccountDetails{
		CustomerID:     "CUST_RESUME_INTEREST",
		CoreCustomerID: "CORE-CUST_RESUME_INTEREST",
		CustomerName:   "Alex Customer",
		Email:          "alex@example.com",
		ProductCode:    "SAVINGS_BASIC",
		InterestRate:   4.5,
		InterestType:   "VARIABLE",
		Currency:       "AUD",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("Save() details error = %v", err)
	}
	customerClient := &fakeCustomerClient{}
	accountClient := &fakeAccountClient{}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)

	resp, err := service.OnboardCard(context.Background(), testRequest("CUST_RESUME_INTEREST"))
	if err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}
	if resp.AccountId != "ACC-CUST_RESUME_INTEREST" ||
		resp.CardId != "CARD-CUST_RESUME_INTEREST-001" ||
		resp.Status != string(entity.StatusSucceeded) {
		t.Fatalf("OnboardCard() response = %#v, want account onboarding completed", resp)
	}
	if customerClient.calls != 0 {
		t.Fatalf("customer calls = %d, want 0", customerClient.calls)
	}
	if accountClient.calls != 0 {
		t.Fatalf("account calls = %d, want 0", accountClient.calls)
	}
}

func TestOnboardCardOverallSucceededIdempotentResponse(t *testing.T) {
	statusStore := store.NewInMemoryRequestStatusStore()
	detailsStore := store.NewInMemoryAccountDetailsStore()
	now := time.Now().UTC()
	if err := statusStore.Save(context.Background(), entity.RequestStatus{
		CustomerID:                 "CUST_DONE",
		OverallStatus:              entity.StatusSucceeded,
		CustomerRegistrationStatus: entity.StatusSucceeded,
		InterestDetailsStatus:      entity.StatusSucceeded,
		AccountOnboardingStatus:    entity.StatusSucceeded,
		CreatedAt:                  now,
		UpdatedAt:                  now,
	}); err != nil {
		t.Fatalf("Save() status error = %v", err)
	}
	if err := detailsStore.Save(context.Background(), entity.AccountDetails{
		CustomerID:     "CUST_DONE",
		CoreCustomerID: "CORE-CUST_DONE",
		AccountID:      "ACC-CUST_DONE",
		CardID:         "CARD-CUST_DONE-001",
		CreatedAt:      now,
		UpdatedAt:      now,
	}); err != nil {
		t.Fatalf("Save() details error = %v", err)
	}
	customerClient := &fakeCustomerClient{}
	accountClient := &fakeAccountClient{}
	service := NewService(statusStore, detailsStore, customerClient, accountClient)

	resp, err := service.OnboardCard(context.Background(), testRequest("CUST_DONE"))
	if err != nil {
		t.Fatalf("OnboardCard() error = %v, want nil", err)
	}
	if resp.CustomerId != "CUST_DONE" ||
		resp.CoreCustomerId != "CORE-CUST_DONE" ||
		resp.AccountId != "ACC-CUST_DONE" ||
		resp.CardId != "CARD-CUST_DONE-001" ||
		resp.Status != string(entity.StatusSucceeded) {
		t.Fatalf("OnboardCard() response = %#v, want existing successful response", resp)
	}
	if customerClient.calls != 0 {
		t.Fatalf("customer calls = %d, want 0", customerClient.calls)
	}
	if accountClient.calls != 0 {
		t.Fatalf("account calls = %d, want 0", accountClient.calls)
	}
}

func testRequest(customerID string) entity.OnboardingRequest {
	return entity.OnboardingRequest{
		CorrelationID: "corr-123",
		JobID:         "job-123",
		RecordID:      "rec-123",
		SourceFile:    "cards.csv",
		RowNumber:     2,
		CustomerID:    customerID,
		CardType:      "DEBIT",
		CardNumber:    "4111111111111111",
		ExpiryDate:    "12/28",
		HolderName:    "Alex Customer",
		Email:         "alex@example.com",
	}
}

type fakeCustomerClient struct {
	resp  customerapi.RegisterCustomerResponse
	err   error
	calls int
}

func (c *fakeCustomerClient) RegisterCustomer(context.Context, customerapi.RegisterCustomerRequest) (customerapi.RegisterCustomerResponse, error) {
	c.calls++
	if c.err != nil {
		return customerapi.RegisterCustomerResponse{}, c.err
	}
	return c.resp, nil
}

type fakeAccountClient struct {
	resp  accountapi.InterestDetailsResponse
	err   error
	calls int
}

func (c *fakeAccountClient) GetInterestDetails(context.Context, string, string) (accountapi.InterestDetailsResponse, error) {
	c.calls++
	if c.err != nil {
		return accountapi.InterestDetailsResponse{}, c.err
	}
	return c.resp, nil
}
