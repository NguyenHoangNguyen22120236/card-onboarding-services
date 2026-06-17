package entity

import "time"

type RequestStatus struct {
	CustomerID                  string    `json:"customerId" dynamodbav:"customerId"`
	OverallStatus               Status    `json:"overallStatus" dynamodbav:"overallStatus"`
	CustomerRegistrationStatus  Status    `json:"customerRegistrationStatus" dynamodbav:"customerRegistrationStatus"`
	CustomerRegistrationMessage string    `json:"customerRegistrationMessage" dynamodbav:"customerRegistrationMessage"`
	InterestDetailsStatus       Status    `json:"interestDetailsStatus" dynamodbav:"interestDetailsStatus"`
	InterestDetailsMessage      string    `json:"interestDetailsMessage" dynamodbav:"interestDetailsMessage"`
	AccountOnboardingStatus     Status    `json:"accountOnboardingStatus" dynamodbav:"accountOnboardingStatus"`
	AccountOnboardingMessage    string    `json:"accountOnboardingMessage" dynamodbav:"accountOnboardingMessage"`
	CreatedAt                   time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt                   time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}
