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
	ErrOrderNotFound  = errors.New("order not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrUserNotFound   = errors.New("user not found")
	ErrPaymentFailed  = errors.New("payment failed")
)

type OrderServiceInterface interface {
	Create(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error)
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	List(ctx context.Context, userID string) ([]*model.Order, error)
	Cancel(ctx context.Context, id uuid.UUID) error
}

type OrderService struct {
	repo          repository.OrderRepositoryInterface
	userClient    client.UserServiceClient
	paymentClient client.PaymentServiceClient
}

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
