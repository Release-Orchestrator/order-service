package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/Release-Orchestrator/order-service/internal/model"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepositoryInterface interface {
	Create(ctx context.Context, order *model.Order) error
	GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error)
	List(ctx context.Context, userID *uuid.UUID) ([]*model.Order, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status string, paymentID *uuid.UUID) error
}

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, order *model.Order) error {
	query := `
		INSERT INTO orders (id, user_id, product_name, amount, status, payment_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query, order.ID, order.UserID, order.ProductName,
		order.Amount, order.Status, order.PaymentID, order.CreatedAt, order.UpdatedAt)
	return err
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Order, error) {
	query := `SELECT id, user_id, product_name, amount, status, payment_id, created_at, updated_at FROM orders WHERE id = $1`
	row := r.db.QueryRow(ctx, query, id)

	var o model.Order
	err := row.Scan(&o.ID, &o.UserID, &o.ProductName, &o.Amount, &o.Status, &o.PaymentID, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &o, nil
}

func (r *OrderRepository) List(ctx context.Context, userID *uuid.UUID) ([]*model.Order, error) {
	query := `SELECT id, user_id, product_name, amount, status, payment_id, created_at, updated_at FROM orders`
	var args []interface{}
	if userID != nil {
		query += " WHERE user_id = $1"
		args = append(args, *userID)
	}
	query += " ORDER BY created_at DESC"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var o model.Order
		if err := rows.Scan(&o.ID, &o.UserID, &o.ProductName, &o.Amount, &o.Status, &o.PaymentID, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		orders = append(orders, &o)
	}
	return orders, rows.Err()
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status string, paymentID *uuid.UUID) error {
	query := `UPDATE orders SET status = $1, payment_id = COALESCE($2, payment_id), updated_at = NOW() WHERE id = $3`
	_, err := r.db.Exec(ctx, query, status, paymentID, id)
	return err
}
