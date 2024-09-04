package postgre

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// OrderRepository is an implementation of order repository.
type OrderRepository struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

// NewOrderRepository returns a new postgre user repository
func NewOrderRepository(db *pgxpool.Pool, log *logger.Logger) *OrderRepository {
	storage := &OrderRepository{db: db, log: log}
	return storage
}

// Migrate migrates the database
func (r *OrderRepository) Migrate(ctx context.Context) error {
	const op = "infrastructure.postgre.OrderRepository.Migrate"
	log := r.log.With(r.log.StringField("op", op))
	_, err := r.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS orders (
    	user_uuid uuid NOT NULL,
        number BIGINT PRIMARY KEY,
        status TEXT NOT NULL,
        accrual DOUBLE PRECISION NOT NULL DEFAULT 0,
        uploaded_at TIMESTAMP NOT NULL);`)
	if err != nil {
		log.Error("Failed to create table", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = r.db.Exec(ctx, `CREATE INDEX IF NOT EXISTS orders_user_uuid_idx ON orders(user_uuid)`)
	if err != nil {
		log.Error("Failed to create index", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// AddOrderForUser adds a new order to the user.
func (r *OrderRepository) AddOrderForUser(ctx context.Context, order entity.Order) error {
	const op = "infrastructure.postgre.OrderRepository.AddOrderForUser"

	log := r.log.With(r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.AnyField("order_number", order.Number),
		r.log.AnyField("user_uuid", order.UserUUID),
	)

	_, err := r.db.Exec(ctx, `INSERT INTO orders (
                    user_uuid,
                    number,
                    status,
                    uploaded_at,
                    accrual
                    ) VALUES ($1, $2, $3, $4, 0)`, order.UserUUID, order.Number, order.Status, order.UploadedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				var dbUserUUID uuid.UUID
				err = r.db.QueryRow(ctx, `SELECT user_uuid FROM orders WHERE number = $1`, order.Number).Scan(&dbUserUUID)
				if err != nil {
					log.Info("Unknown error", log.ErrorField(err))
					return fmt.Errorf("%s: %w", op, err)
				}
				if dbUserUUID == order.UserUUID {
					log.Info("Order already uploaded from current user")
					return entity.ErrOrderAlreadyUploaded
				} else {
					log.Info("Order already uploaded from another user")
					return entity.ErrOrderAlreadyUploadedByAnotherUser
				}
			}
		}
		log.Error("Failed to create order", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetAllUserOrders returns all orders for a user.
func (r *OrderRepository) GetAllUserOrders(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error) {
	const op = "infrastructure.postgre.OrderRepository.GetAllUserOrders"
	log := r.log.With(r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	rows, _ := r.db.Query(ctx, `SELECT 
    				user_uuid,
                    number,
                    status,
                    uploaded_at,
                    accrual FROM orders WHERE user_uuid = $1`, userUUID)
	defer rows.Close()
	orders, err := pgx.CollectRows(rows, func(row pgx.CollectableRow) (entity.Order, error) {
		var order entity.Order
		err := row.Scan(&order.UserUUID, &order.Number, &order.Status, &order.UploadedAt, &order.Accrual)
		if err != nil {
			log.Error("Failed to scan row", log.ErrorField(err))
			return entity.Order{}, fmt.Errorf("%s: %w", op, err)
		}
		return order, err
	})
	if err != nil {
		log.Error("Failed to get orders", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	if len(orders) == 0 {
		log.Info("No orders found")
		return nil, entity.ErrOrderNotFound
	}
	return orders, nil
}

func (r *OrderRepository) UpdateOrderForUser(ctx context.Context, order entity.Order) error {
	const op = "infrastructure.postgre.OrderRepository.UpdateOrderForUser"
	log := r.log.With(r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.AnyField("order_number", order.Number),
		r.log.AnyField("user_uuid", order.UserUUID),
	)
	_, err := r.db.Exec(ctx, `UPDATE orders SET 
                  status = $1,
                  accrual = $2
              WHERE
                  user_uuid = $3 
                AND
                  number = $4`, order.Status, order.Accrual, order.UserUUID, order.Number)
	if err != nil {
		log.Error("Failed to update order", log.ErrorField(err))
	}
	return nil
}
