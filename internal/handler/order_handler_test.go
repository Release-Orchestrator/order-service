package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Release-Orchestrator/order-service/internal/model"
	"github.com/Release-Orchestrator/order-service/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type mockService struct {
	createFunc func(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error)
	getByIDFunc func(ctx context.Context, id uuid.UUID) (*model.Order, error)
	listFunc    func(ctx context.Context, userID string) ([]*model.Order, error)
	cancelFunc  func(ctx context.Context, id uuid.UUID) error
}

func (m *mockService) Create(ctx context.Context, req *model.CreateOrderRequest) (*model.Order, error) {
	return m.createFunc(ctx, req)
}

func (m *mockService) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	return m.getByIDFunc(ctx, id)
}

func (m *mockService) List(ctx context.Context, userID string) ([]*model.Order, error) {
	return m.listFunc(ctx, userID)
}

func (m *mockService) Cancel(ctx context.Context, id uuid.UUID) error {
	return m.cancelFunc(ctx, id)
}

func setupRouter(h *OrderHandler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "UP"})
	})
	api := r.Group("/api/v1")
	h.RegisterRoutes(api)
	return r
}

func TestHealthEndpoint(t *testing.T) {
	h := NewOrderHandler(&mockService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCreateOrder_Success(t *testing.T) {
	svc := &mockService{
		createFunc: func(_ context.Context, req *model.CreateOrderRequest) (*model.Order, error) {
			return &model.Order{
				ID:          uuid.New(),
				UserID:      uuid.MustParse(req.UserID),
				ProductName: req.Product,
				Amount:      req.Amount,
				Status:      model.OrderStatusCreated,
			}, nil
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id": uuid.New().String(),
		"product": "Laptop",
		"amount":  1500.00,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestCreateOrder_InvalidBody(t *testing.T) {
	h := NewOrderHandler(&mockService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewReader([]byte(`{invalid}`)))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestCreateOrder_UserNotFound(t *testing.T) {
	svc := &mockService{
		createFunc: func(_ context.Context, _ *model.CreateOrderRequest) (*model.Order, error) {
			return nil, service.ErrUserNotFound
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	body, _ := json.Marshal(map[string]interface{}{
		"user_id": uuid.New().String(),
		"product": "Laptop",
		"amount":  100,
	})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/v1/orders", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetOrder_Success(t *testing.T) {
	orderID := uuid.New()
	svc := &mockService{
		getByIDFunc: func(_ context.Context, id uuid.UUID) (*model.Order, error) {
			return &model.Order{ID: id, UserID: uuid.New(), ProductName: "Phone", Status: model.OrderStatusCreated}, nil
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/orders/"+orderID.String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetOrder_NotFound(t *testing.T) {
	svc := &mockService{
		getByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Order, error) {
			return nil, service.ErrOrderNotFound
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/orders/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestGetOrder_InvalidID(t *testing.T) {
	h := NewOrderHandler(&mockService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/orders/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestListOrders_Success(t *testing.T) {
	svc := &mockService{
		listFunc: func(_ context.Context, _ string) ([]*model.Order, error) {
			return []*model.Order{
				{ID: uuid.New(), ProductName: "Laptop", Status: model.OrderStatusCreated},
				{ID: uuid.New(), ProductName: "Phone", Status: model.OrderStatusCreated},
			}, nil
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/orders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestListOrders_Empty(t *testing.T) {
	svc := &mockService{
		listFunc: func(_ context.Context, _ string) ([]*model.Order, error) {
			return []*model.Order{}, nil
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/orders", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCancelOrder_Success(t *testing.T) {
	svc := &mockService{
		cancelFunc: func(_ context.Context, _ uuid.UUID) error {
			return nil
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/orders/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d", w.Code)
	}
}

func TestCancelOrder_NotFound(t *testing.T) {
	svc := &mockService{
		cancelFunc: func(_ context.Context, _ uuid.UUID) error {
			return service.ErrOrderNotFound
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/orders/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}

func TestCancelOrder_InvalidID(t *testing.T) {
	h := NewOrderHandler(&mockService{})
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/api/v1/orders/bad-id", nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleError_InternalError(t *testing.T) {
	svc := &mockService{
		getByIDFunc: func(_ context.Context, _ uuid.UUID) (*model.Order, error) {
			return nil, errors.New("unexpected db failure")
		},
	}
	h := NewOrderHandler(svc)
	r := setupRouter(h)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/orders/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}
