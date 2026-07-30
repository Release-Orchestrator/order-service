package service

import (
	"context"
	"errors"
	"testing"

	"github.com/Release-Orchestrator/order-service/internal/client"
	"github.com/Release-Orchestrator/order-service/internal/model"
	"github.com/google/uuid"
)

type mockRepo struct {
	orders          map[uuid.UUID]*model.Order
	createErr       error
	getByIDErr      error
	listErr         error
	updateStatusErr error
}

func newMockRepo() *mockRepo {
	return &mockRepo{
		orders: make(map[uuid.UUID]*model.Order),
	}
}

func (m *mockRepo) Create(_ context.Context, order *model.Order) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.orders[order.ID] = order
	return nil
}

func (m *mockRepo) GetByID(_ context.Context, id uuid.UUID) (*model.Order, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	o, ok := m.orders[id]
	if !ok {
		return nil, nil
	}
	return o, nil
}

func (m *mockRepo) List(_ context.Context, userID *uuid.UUID) ([]*model.Order, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	var result []*model.Order
	for _, o := range m.orders {
		if userID != nil && o.UserID != *userID {
			continue
		}
		result = append(result, o)
	}
	return result, nil
}

func (m *mockRepo) UpdateStatus(_ context.Context, id uuid.UUID, status string, paymentID *uuid.UUID) error {
	if m.updateStatusErr != nil {
		return m.updateStatusErr
	}
	o, ok := m.orders[id]
	if !ok {
		return nil
	}
	o.Status = status
	if paymentID != nil {
		o.PaymentID = paymentID
	}
	return nil
}

type mockUserClient struct {
	validateFunc func(ctx context.Context, userID uuid.UUID) (bool, error)
}

func (m *mockUserClient) ValidateUser(ctx context.Context, userID uuid.UUID) (bool, error) {
	return m.validateFunc(ctx, userID)
}

type mockPaymentClient struct {
	processFunc func(ctx context.Context, orderID uuid.UUID, amount float64) (*client.PaymentResult, error)
}

func (m *mockPaymentClient) ProcessPayment(ctx context.Context, orderID uuid.UUID, amount float64) (*client.PaymentResult, error) {
	return m.processFunc(ctx, orderID, amount)
}

func TestCreateOrder_Success(t *testing.T) {
	repo := newMockRepo()
	userClient := &mockUserClient{
		validateFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	paymentClient := &mockPaymentClient{}

	svc := NewOrderService(repo, userClient, paymentClient)

	order, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  uuid.New().String(),
		Product: "Laptop",
		Amount:  1500.00,
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ProductName != "Laptop" {
		t.Fatalf("expected Laptop, got %s", order.ProductName)
	}
	if order.Status != model.OrderStatusCreated {
		t.Fatalf("expected %s, got %s", model.OrderStatusCreated, order.Status)
	}
	if order.ID == uuid.Nil {
		t.Fatal("expected non-nil UUID")
	}
}

func TestCreateOrder_InvalidProduct(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	_, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  uuid.New().String(),
		Product: "",
		Amount:  100,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateOrder_InvalidAmount(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	_, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  uuid.New().String(),
		Product: "Laptop",
		Amount:  0,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateOrder_InvalidUserID(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	_, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  "not-a-uuid",
		Product: "Laptop",
		Amount:  100,
	})
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCreateOrder_UserNotFound(t *testing.T) {
	repo := newMockRepo()
	userClient := &mockUserClient{
		validateFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return false, nil
		},
	}
	svc := NewOrderService(repo, userClient, &mockPaymentClient{})

	_, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  uuid.New().String(),
		Product: "Laptop",
		Amount:  100,
	})
	if !errors.Is(err, ErrUserNotFound) {
		t.Fatalf("expected ErrUserNotFound, got %v", err)
	}
}

func TestCreateOrder_UserClientError(t *testing.T) {
	repo := newMockRepo()
	userClient := &mockUserClient{
		validateFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return false, errors.New("user service unavailable")
		},
	}
	svc := NewOrderService(repo, userClient, &mockPaymentClient{})

	_, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  uuid.New().String(),
		Product: "Laptop",
		Amount:  100,
	})
	if err == nil || err.Error() != "user service unavailable" {
		t.Fatalf("expected 'user service unavailable', got %v", err)
	}
}

func TestCreateOrder_RepoError(t *testing.T) {
	repo := newMockRepo()
	repo.createErr = errors.New("db error")
	userClient := &mockUserClient{
		validateFunc: func(_ context.Context, _ uuid.UUID) (bool, error) {
			return true, nil
		},
	}
	svc := NewOrderService(repo, userClient, &mockPaymentClient{})

	_, err := svc.Create(context.Background(), &model.CreateOrderRequest{
		UserID:  uuid.New().String(),
		Product: "Laptop",
		Amount:  100,
	})
	if err == nil || err.Error() != "db error" {
		t.Fatalf("expected 'db error', got %v", err)
	}
}

func TestGetByID_Success(t *testing.T) {
	repo := newMockRepo()
	oid := uuid.New()
	repo.orders[oid] = &model.Order{ID: oid, UserID: uuid.New(), ProductName: "Phone"}

	svc := NewOrderService(repo, &mockUserClient{}, &mockPaymentClient{})

	order, err := svc.GetByID(context.Background(), oid)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if order.ID != oid {
		t.Fatalf("expected ID %v, got %v", oid, order.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	_, err := svc.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestList_Success(t *testing.T) {
	repo := newMockRepo()
	repo.orders[uuid.New()] = &model.Order{ID: uuid.New(), UserID: uuid.New()}
	repo.orders[uuid.New()] = &model.Order{ID: uuid.New(), UserID: uuid.New()}

	svc := NewOrderService(repo, &mockUserClient{}, &mockPaymentClient{})

	orders, err := svc.List(context.Background(), "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(orders))
	}
}

func TestList_Empty(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	orders, err := svc.List(context.Background(), "")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(orders) != 0 {
		t.Fatalf("expected 0 orders, got %d", len(orders))
	}
}

func TestList_InvalidUserID(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	_, err := svc.List(context.Background(), "bad-uuid")
	if !errors.Is(err, ErrInvalidInput) {
		t.Fatalf("expected ErrInvalidInput, got %v", err)
	}
}

func TestCancel_Success(t *testing.T) {
	repo := newMockRepo()
	oid := uuid.New()
	repo.orders[oid] = &model.Order{ID: oid, UserID: uuid.New(), Status: model.OrderStatusCreated}

	svc := NewOrderService(repo, &mockUserClient{}, &mockPaymentClient{})

	err := svc.Cancel(context.Background(), oid)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if repo.orders[oid].Status != model.OrderStatusCancelled {
		t.Fatalf("expected %s, got %s", model.OrderStatusCancelled, repo.orders[oid].Status)
	}
}

func TestCancel_NotFound(t *testing.T) {
	svc := NewOrderService(newMockRepo(), &mockUserClient{}, &mockPaymentClient{})

	err := svc.Cancel(context.Background(), uuid.New())
	if !errors.Is(err, ErrOrderNotFound) {
		t.Fatalf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestCancel_AlreadyCancelled(t *testing.T) {
	repo := newMockRepo()
	oid := uuid.New()
	repo.orders[oid] = &model.Order{ID: oid, UserID: uuid.New(), Status: model.OrderStatusCancelled}

	svc := NewOrderService(repo, &mockUserClient{}, &mockPaymentClient{})

	err := svc.Cancel(context.Background(), oid)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
}
