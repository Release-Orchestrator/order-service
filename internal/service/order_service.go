// Package service implements business logic for orders.
package service

import (
	"context"
	"errors"
	"time"

	"github.com/Release-Orchestrator/order-service/internal/client"
	"github.com/Release-Orchestrator/order-service/internal/model"
	"github.com/Release-Orchestrator/order-service/internal/repository"
	"github.com/google/uuid"
)

var (
	// ErrOrderNotFound is returned when an order cannot be found.
	ErrOrderNotFound = errors.New("order not found")
	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")
	// ErrUserNotFound is returned when the user does not exist.
	ErrUserNotFound = errors.New("user not found")
	// ErrPaymentFailed is returned when payment processing fails.
	ErrPaymentFailed = errors.New("payment failed")
)

// OrderServiceInterface describes service operations for orders.
type OrderServiceInterface interface {
	Create(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	List(ctx context.Context, userID string) ([]*model.Order, error)
	Cancel(ctx context.Context, id uuid.UUID) error
}

// OrderService implements OrderServiceInterface.
type OrderService struct {
	repo          repository.OrderRepositoryInterface
	userClient    client.UserServiceClient
	paymentClient client.PaymentServiceClient
}

// NewOrderService creates a new OrderService.
func NewOrderService(
	repo repository.OrderRepositoryInterface,
	userClient client.UserServiceClient,
	paymentClient client.PaymentServiceClient,
) *OrderService {
	return &OrderService{
		repo:          repo,
		userClient:    userClient,
		paymentClient: paymentClient,
	}
}

// Create validates user and saves a new order.
func (s *OrderService) Create(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error) {
	if req.Product == "" || req.Amount <= 0 {
		return nil, ErrInvalidInput
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		return nil, ErrInvalidInput
	}

	exists, err := s.userClient.ValidateUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, ErrUserNotFound
	}

	now := time.Now().UTC()
	order := &model.Order{
		ID:          uuid.New(),
		UserID:      userID,
		ProductName: req.Product,
		Amount:      req.Amount,
		Status:      model.OrderStatusCreated,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := s.repo.Create(ctx, order); err != nil {
		return nil, err
	}
	return order, nil
}

// GetByID returns an order by ID.
func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if order == nil {
		return nil, ErrOrderNotFound
	}
	return order, nil
}

// List returns orders filtered by user.
func (s *OrderService) List(ctx context.Context, userIDStr string) ([]*model.Order, error) {
	var userID *uuid.UUID
	if userIDStr != "" {
		parsed, err := uuid.Parse(userIDStr)
		if err != nil {
			return nil, ErrInvalidInput
		}
		userID = &parsed
	}
	return s.repo.List(ctx, userID)
}

// Cancel cancels an order by ID.
func (s *OrderService) Cancel(ctx context.Context, id uuid.UUID) error {
	order, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if order == nil {
		return ErrOrderNotFound
	}

	if order.Status == model.OrderStatusCancelled {
		return nil
	}

	return s.repo.UpdateStatus(ctx, id, model.OrderStatusCancelled, order.PaymentID)
}
