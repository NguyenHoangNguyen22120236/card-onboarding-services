package store

import (
	"context"
	"sync"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
)

type AccountDetailsStore interface {
	GetByCustomerID(ctx context.Context, customerID string) (entity.AccountDetails, error)
	Save(ctx context.Context, details entity.AccountDetails) error
	Update(ctx context.Context, details entity.AccountDetails) error
}

type InMemoryAccountDetailsStore struct {
	mu      sync.RWMutex
	details map[string]entity.AccountDetails
}

func NewInMemoryAccountDetailsStore() *InMemoryAccountDetailsStore {
	return &InMemoryAccountDetailsStore{
		details: make(map[string]entity.AccountDetails),
	}
}

func (s *InMemoryAccountDetailsStore) GetByCustomerID(ctx context.Context, customerID string) (entity.AccountDetails, error) {
	if err := ctx.Err(); err != nil {
		return entity.AccountDetails{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	details, ok := s.details[customerID]
	if !ok {
		return entity.AccountDetails{}, ErrNotFound
	}

	return details, nil
}

func (s *InMemoryAccountDetailsStore) Save(ctx context.Context, details entity.AccountDetails) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.details[details.CustomerID] = details
	return nil
}

func (s *InMemoryAccountDetailsStore) Update(ctx context.Context, details entity.AccountDetails) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.details[details.CustomerID]; !ok {
		return ErrNotFound
	}

	s.details[details.CustomerID] = details
	return nil
}
