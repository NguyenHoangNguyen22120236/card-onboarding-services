package entity

import "time"

type AccountDetails struct {
	CustomerID       string    `json:"customerId" dynamodbav:"customerId"`
	CoreCustomerID   string    `json:"coreCustomerId" dynamodbav:"coreCustomerId"`
	CustomerName     string    `json:"customerName" dynamodbav:"customerName"`
	Email            string    `json:"email" dynamodbav:"email"`
	ProductCode      string    `json:"productCode" dynamodbav:"productCode"`
	InterestRate     float64   `json:"interestRate" dynamodbav:"interestRate"`
	InterestType     string    `json:"interestType" dynamodbav:"interestType"`
	Currency         string    `json:"currency" dynamodbav:"currency"`
	AccountID        string    `json:"accountId" dynamodbav:"accountId"`
	CardID           string    `json:"cardId" dynamodbav:"cardId"`
	CardType         string    `json:"cardType" dynamodbav:"cardType"`
	CardNumberMasked string    `json:"cardNumberMasked" dynamodbav:"cardNumberMasked"`
	CreatedAt        time.Time `json:"createdAt" dynamodbav:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt" dynamodbav:"updatedAt"`
}
