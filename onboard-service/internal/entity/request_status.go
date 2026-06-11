package entity

import "time"

type RequestStatus struct {
	CustomerID                  string    `json:"customerId"`
	OverallStatus               Status    `json:"overallStatus"`
	CustomerRegistrationStatus  Status    `json:"customerRegistrationStatus"`
	CustomerRegistrationMessage string    `json:"customerRegistrationMessage"`
	InterestDetailsStatus       Status    `json:"interestDetailsStatus"`
	InterestDetailsMessage      string    `json:"interestDetailsMessage"`
	AccountOnboardingStatus     Status    `json:"accountOnboardingStatus"`
	AccountOnboardingMessage    string    `json:"accountOnboardingMessage"`
	CreatedAt                   time.Time `json:"createdAt"`
	UpdatedAt                   time.Time `json:"updatedAt"`
}
