package repository

import (
	"cache-app/internal/model"
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type OrderRepository interface {
	SaveOrder(ctx context.Context, order *model.Order) (string, error)
	GetOrderByUID(ctx context.Context, orderUID string) (*model.Order, error)
	GetAllOrders(ctx context.Context) ([]*model.Order, error)
}

type orderRepo struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) OrderRepository {
	return &orderRepo{db: db}
}

func (r *orderRepo) SaveOrder(ctx context.Context, order *model.Order) (string, error) {
	// Если UID не задан — генерируем новый
	if order.OrderUID == "" {
		order.OrderUID = uuid.NewString()
	}

	deliveryJSON, err := json.Marshal(order.Delivery)
	if err != nil {
		return "", fmt.Errorf("marshal delivery: %w", err)
	}

	paymentJSON, err := json.Marshal(order.Payment)
	if err != nil {
		return "", fmt.Errorf("marshal payment: %w", err)
	}

	itemsJSON, err := json.Marshal(order.Items)
	if err != nil {
		return "", fmt.Errorf("marshal items: %w", err)
	}

	query := `
		INSERT INTO orders (
			order_uid, track_number, entry, delivery, payment, items, 
			locale, internal_signature, customer_id, delivery_service, 
			shardkey, sm_id, date_created, oof_shard
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			delivery = EXCLUDED.delivery,
			payment = EXCLUDED.payment,
			items = EXCLUDED.items,
			date_created = EXCLUDED.date_created
	`

	_, err = r.db.Exec(ctx, query,
		order.OrderUID, order.TrackNumber, order.Entry,
		deliveryJSON, paymentJSON, itemsJSON,
		order.Locale, order.InternalSignature, order.CustomerID,
		order.DeliveryService, order.Shardkey, order.SmID,
		order.DateCreated, order.OofShard,
	)

	// Проверка на уникальность
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		if pgErr.Code == "23505" { // unique_violation
			return "", fmt.Errorf("order with uid %s already exists", order.OrderUID)
		}
	}

	if err != nil {
		return "", fmt.Errorf("failed to save order: %w", err)
	}

	return order.OrderUID, nil
}

func (r *orderRepo) GetOrderByUID(ctx context.Context, orderUID string) (*model.Order, error) {
	query := `
		SELECT order_uid, track_number, entry, delivery, payment, items,
		       locale, internal_signature, customer_id, delivery_service,
		       shardkey, sm_id, date_created, oof_shard
		FROM orders WHERE order_uid = $1
	`

	var order model.Order
	var deliveryJSON, paymentJSON, itemsJSON []byte

	err := r.db.QueryRow(ctx, query, orderUID).Scan(
		&order.OrderUID, &order.TrackNumber, &order.Entry, &deliveryJSON, &paymentJSON, &itemsJSON,
		&order.Locale, &order.InternalSignature, &order.CustomerID, &order.DeliveryService,
		&order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
	)

	if err != nil {
		return nil, err
	}

	json.Unmarshal(deliveryJSON, &order.Delivery)
	json.Unmarshal(paymentJSON, &order.Payment)
	json.Unmarshal(itemsJSON, &order.Items)

	return &order, nil
}

func (r *orderRepo) GetAllOrders(ctx context.Context) ([]*model.Order, error) {
	query := `SELECT order_uid FROM orders`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*model.Order
	for rows.Next() {
		var orderUID string
		if err := rows.Scan(&orderUID); err != nil {
			continue
		}
		order, err := r.GetOrderByUID(ctx, orderUID)
		if err != nil {
			continue
		}
		orders = append(orders, order)
	}

	return orders, nil
}
