package store

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
)

func TestInMemoryRequestStatusStoreSaveAndGet(t *testing.T) {
	store := NewInMemoryRequestStatusStore()
	ctx := context.Background()
	status := entity.RequestStatus{
		CustomerID:                 "customer-1",
		OverallStatus:              entity.StatusInProgress,
		CustomerRegistrationStatus: entity.StatusSucceeded,
		InterestDetailsStatus:      entity.StatusInProgress,
		AccountOnboardingStatus:    entity.StatusFailed,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	if err := store.Save(ctx, status); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	got, err := store.GetByCustomerID(ctx, status.CustomerID)
	if err != nil {
		t.Fatalf("GetByCustomerID() error = %v", err)
	}

	if got != status {
		t.Fatalf("GetByCustomerID() = %#v, want %#v", got, status)
	}
}

func TestInMemoryRequestStatusStoreGetMissingReturnsErrNotFound(t *testing.T) {
	store := NewInMemoryRequestStatusStore()

	_, err := store.GetByCustomerID(context.Background(), "missing-customer")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetByCustomerID() error = %v, want ErrNotFound", err)
	}
}

func TestInMemoryRequestStatusStoreUpdate(t *testing.T) {
	store := NewInMemoryRequestStatusStore()
	ctx := context.Background()
	status := entity.RequestStatus{
		CustomerID:                 "customer-1",
		OverallStatus:              entity.StatusInProgress,
		CustomerRegistrationStatus: entity.StatusInProgress,
		InterestDetailsStatus:      entity.StatusInProgress,
		AccountOnboardingStatus:    entity.StatusInProgress,
		CreatedAt:                  time.Now(),
		UpdatedAt:                  time.Now(),
	}

	if err := store.Save(ctx, status); err != nil {
		t.Fatalf("Save() error = %v", err)
	}

	status.OverallStatus = entity.StatusSucceeded
	status.CustomerRegistrationStatus = entity.StatusSucceeded
	status.InterestDetailsStatus = entity.StatusSucceeded
	status.AccountOnboardingStatus = entity.StatusSucceeded
	status.UpdatedAt = status.UpdatedAt.Add(time.Minute)

	if err := store.Update(ctx, status); err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	got, err := store.GetByCustomerID(ctx, status.CustomerID)
	if err != nil {
		t.Fatalf("GetByCustomerID() error = %v", err)
	}

	if got != status {
		t.Fatalf("GetByCustomerID() = %#v, want %#v", got, status)
	}
}
