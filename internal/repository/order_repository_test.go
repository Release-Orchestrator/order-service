package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Release-Orchestrator/order-service/internal/model"
	"github.com/google/uuid"
)

type mockDB struct {
	OrderRepositoryInterface
	orders          map[uuid.UUID]*model.Order
	createErr       error
	getByIDErr      error
	listErr         error
	updateStatusErr error
}

func newMockDB() *mockDB {
	return &mockDB{
		orders: make(map[uuid.UUID]*model.Order),
	}
}

func (m *mockDB) Create(_ context.Context, order *model.Order) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.orders[order.ID] = order
	return nil
}

func (m *mockDB) GetByID(_ context.Context, id uuid.UUID) (*model.Order, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	o, ok := m.orders[id]
	if !ok {
		return nil, nil
	}
	return o, nil
}

func (m *mockDB) List(_ context.Context, userID *uuid.UUID) ([]*model.Order, error) {
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

func (m *mockDB) UpdateStatus(_ context.Context, id uuid.UUID, status string, paymentID *uuid.UUID) error {
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

func now() time.Time {
	return time.Now().UTC().Truncate(time.Second)
}

func TestCreateOrder_Success(t *testing.T) {
	db := newMockDB()
	pid := uuid.New()
	o := &model.Order{
		ID:          uuid.New(),
		UserID:      uuid.New(),
		ProductName: "Laptop",
		Amount:      1500.00,
		Status:      model.OrderStatusCreated,
		PaymentID:   &pid,
		CreatedAt:   now(),
		UpdatedAt:   now(),
	}

	err := db.Create(context.Background(), o)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(db.orders) != 1 {
		t.Fatalf("expected 1 order, got %d", len(db.orders))
	}
}

func TestCreateOrder_Error(t *testing.T) {
	db := newMockDB()
	db.createErr = errors.New("db error")

	err := db.Create(context.Background(), &model.Order{ID: uuid.New()})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetByID_Success(t *testing.T) {
	db := newMockDB()
	o := &model.Order{ID: uuid.New(), UserID: uuid.New(), ProductName: "Phone"}
	db.orders[o.ID] = o

	result, err := db.GetByID(context.Background(), o.ID)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result.ID != o.ID {
		t.Fatalf("expected ID %v, got %v", o.ID, result.ID)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	db := newMockDB()
	result, err := db.GetByID(context.Background(), uuid.New())
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != nil {
		t.Fatal("expected nil, got order")
	}
}

func TestList_All(t *testing.T) {
	db := newMockDB()
	db.orders[uuid.New()] = &model.Order{ID: uuid.New(), UserID: uuid.New()}
	db.orders[uuid.New()] = &model.Order{ID: uuid.New(), UserID: uuid.New()}

	result, err := db.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 orders, got %d", len(result))
	}
}

func TestList_Empty(t *testing.T) {
	db := newMockDB()
	result, err := db.List(context.Background(), nil)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 orders, got %d", len(result))
	}
}

func TestList_FilterByUserID(t *testing.T) {
	db := newMockDB()
	targetUser := uuid.New()
	db.orders[uuid.New()] = &model.Order{ID: uuid.New(), UserID: targetUser}
	db.orders[uuid.New()] = &model.Order{ID: uuid.New(), UserID: uuid.New()}

	result, err := db.List(context.Background(), &targetUser)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 order, got %d", len(result))
	}
}

func TestUpdateStatus_Success(t *testing.T) {
	db := newMockDB()
	pid := uuid.New()
	o := &model.Order{ID: uuid.New(), Status: model.OrderStatusCreated}
	db.orders[o.ID] = o

	err := db.UpdateStatus(context.Background(), o.ID, model.OrderStatusPaid, &pid)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if db.orders[o.ID].Status != model.OrderStatusPaid {
		t.Fatalf("expected status %s, got %s", model.OrderStatusPaid, db.orders[o.ID].Status)
	}
	if *db.orders[o.ID].PaymentID != pid {
		t.Fatalf("expected paymentID %v, got %v", pid, *db.orders[o.ID].PaymentID)
	}
}
