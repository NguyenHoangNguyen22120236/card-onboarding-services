package entity

import "time"

type AccountDetails struct {
	CustomerID       string    `json:"customerId"`
	CoreCustomerID   string    `json:"coreCustomerId"`
	CustomerName     string    `json:"customerName"`
	Email            string    `json:"email"`
	ProductCode      string    `json:"productCode"`
	InterestRate     float64   `json:"interestRate"`
	InterestType     string    `json:"interestType"`
	Currency         string    `json:"currency"`
	AccountID        string    `json:"accountId"`
	CardID           string    `json:"cardId"`
	CardType         string    `json:"cardType"`
	CardNumberMasked string    `json:"cardNumberMasked"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}
