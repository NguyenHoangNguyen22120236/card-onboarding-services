package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
)

func TestInMemoryAccountDetailsStoreSaveAndGet(t *testing.T) {
	store := NewInMemoryAccountDetailsStore()
	ctx := context.Background()
	details := entity.AccountDetails{
		CustomerID:       "customer-1",
		CoreCustomerID:   "core-customer-1",
		CustomerName:     "Alex Customer",
		Email:            "alex@example.com",
		ProductCode:      "CARD-GOLD",
		InterestRate:     1.25,
		InterestType:     "FIXED",
		Currency:         "USD",
		AccountID:        "account-1",
		CardID:           "card-1",
		CardType:         "GOLD",
		CardNumberMasked: "**** **** **** 1111",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := store.Save(ctx, details); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.GetByCustomerID(ctx, details.CustomerID)
	if err != nil {
		t.Fatalf("GetByCustomerID() error = %v", err)
	}

	if got != details {
		t.Fatalf("GetByCustomerID() = %#v, want %#v", got, details)
	}
}

func TestInMemoryAccountDetailsStoreGetMissingReturnsErrNotFound(t *testing.T) {
	store := NewInMemoryAccountDetailsStore()

	_, err := store.GetByCustomerID(context.Background(), "missing-customer")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByCustomerID() error = %v, want ErrNotFound", err)
	}
}

func TestInMemoryAccountDetailsStoreUpdate(t *testing.T) {
	store := NewInMemoryAccountDetailsStore()
	ctx := context.Background()
	details := entity.AccountDetails{
		CustomerID:       "customer-1",
		CoreCustomerID:   "core-customer-1",
		CustomerName:     "Alex Customer",
		Email:            "alex@example.com",
		ProductCode:      "CARD-GOLD",
		InterestRate:     1.25,
		InterestType:     "FIXED",
		Currency:         "USD",
		AccountID:        "account-1",
		CardID:           "card-1",
		CardType:         "GOLD",
		CardNumberMasked: "**** **** **** 1111",
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}

	if err := store.Save(ctx, details); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	details.AccountID = "account-2"
	details.CardID = "card-2"
	details.CardNumberMasked = "**** **** **** 2222"
	details.UpdatedAt = details.UpdatedAt.Add(time.Minute)

	if err := store.Update(ctx, details); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := store.GetByCustomerID(ctx, details.CustomerID)
	if err != nil {
		t.Fatalf("GetByCustomerID() error = %v", err)
	}

	if got != details {
		t.Fatalf("GetByCustomerID() = %#v, want %#v", got, details)
	}
}
