package store

import (
	"context"
	"errors"
	"sync"

	"github.com/NguyenHoangNguyen22120236/card-onboarding-services/onboard-service/internal/entity"
)

var ErrNotFound = errors.New("store: item not found")

type RequestStatusStore interface {
	GetByCustomerID(ctx context.Context, customerID string) (entity.RequestStatus, error)
	Save(ctx context.Context, status entity.RequestStatus) error
	Update(ctx context.Context, status entity.RequestStatus) error
}

type InMemoryRequestStatusStore struct {
	mu       sync.RWMutex
	statuses map[string]entity.RequestStatus
}

func NewInMemoryRequestStatusStore() *InMemoryRequestStatusStore {
	return &InMemoryRequestStatusStore{
		statuses: make(map[string]entity.RequestStatus),
	}
}

func (s *InMemoryRequestStatusStore) GetByCustomerID(ctx context.Context, customerID string) (entity.RequestStatus, error) {
	if err := ctx.Err(); err != nil {
		return entity.RequestStatus{}, err
	}

	s.mu.RLock()
	defer s.mu.RUnlock()

	status, ok := s.statuses[customerID]
	if !ok {
		return entity.RequestStatus{}, ErrNotFound
	}

	return status, nil
}

func (s *InMemoryRequestStatusStore) Save(ctx context.Context, status entity.RequestStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	s.statuses[status.CustomerID] = status
	return nil
}

func (s *InMemoryRequestStatusStore) Update(ctx context.Context, status entity.RequestStatus) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.statuses[status.CustomerID]; !ok {
		return ErrNotFound
	}

	s.statuses[status.CustomerID] = status
	return nil
}
