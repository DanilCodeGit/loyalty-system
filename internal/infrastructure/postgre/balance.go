package postgre

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// BalanceRepository is structure for repository for balance service
type BalanceRepository struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

// NewBalanceRepository returns a new BalanceRepository
func NewBalanceRepository(db *pgxpool.Pool, log *logger.Logger) *BalanceRepository {
	storage := &BalanceRepository{db: db, log: log}
	return storage
}

// Migrate migrates the database
func (r *BalanceRepository) Migrate(ctx context.Context) error {
	const op = "infrastructure.postgre.BalanceRepository.Migrate"
	log := r.log.With(r.log.StringField("op", op))
	_, err := r.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS user_balances (
        user_uuid uuid PRIMARY KEY NOT NULL,
        current DOUBLE PRECISION NOT NULL DEFAULT 0,
        withdraw DOUBLE PRECISION NOT NULL DEFAULT 0);`)
	if err != nil {
		log.Error("Failed to create table user_balances", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = r.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS balance_operations (
        uuid uuid PRIMARY KEY NOT NULL,
        user_uuid uuid NOT NULL,
        accrual DOUBLE PRECISION NOT NULL DEFAULT 0,
        withdrawal DOUBLE PRECISION NOT NULL DEFAULT 0,
        order_number BIGINT NOT NULL UNIQUE,
        processed_at TIMESTAMP NOT NULL)`)
	if err != nil {
		log.Error("Failed to create table balance_operations", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// GetBalance returns the balance of the user
func (r *BalanceRepository) GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error) {
	const op = "infrastructure.postgre.BalanceRepository.GetBalance"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	var balance entity.Balance
	err := r.db.QueryRow(ctx, `SELECT current, withdraw FROM user_balances WHERE user_uuid = $1`, userUUID).Scan(&balance.Current, &balance.Withdraw)
	if err != nil {
		log.Error("Failed to get balance", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &balance, nil
}

// GetWithdrawOperations returns a list of withdrawal operations
func (r *BalanceRepository) GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error) {
	const op = "infrastructure.postgre.BalanceRepository.GetWithdrawOperations"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	var operations []entity.BalanceOperation
	result, err := r.db.Query(ctx, `SELECT 
											uuid, 
											user_uuid, 
											accrual, 
											withdrawal, 
											order_number, 
											processed_at 
										FROM 
										    balance_operations 
										WHERE 
										    user_uuid = $1 
										  AND 
										    withdrawal > 0`, userUUID)
	if err != nil {
		log.Error("Failed to get withdraw operations", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer result.Close()
	for result.Next() {
		var operation entity.BalanceOperation
		err = result.Scan(&operation.UUID, &operation.UserUUID, &operation.Accrual, &operation.Withdrawal, &operation.OrderNumber, &operation.ProcessedAt)
		if err != nil {
			log.Error("Failed to scan row", log.ErrorField(err))
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		operations = append(operations, operation)
	}
	if len(operations) == 0 {
		log.Info("No withdraw operations found")
		return nil, entity.ErrBalanceOperationsNotFound
	}
	return operations, nil
}

// CreateBalance creates a new balance for the user
func (r *BalanceRepository) CreateBalance(ctx context.Context, userUUID uuid.UUID) error {
	const op = "infrastructure.postgre.BalanceRepository.CreateBalance"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)
	_, err := r.db.Exec(ctx, `INSERT INTO user_balances (user_uuid, current, withdraw) VALUES ($1, 0, 0)`, userUUID)
	if err != nil {
		log.Error("Failed to create balance", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

// Withdraw executes a balance operation
func (r *BalanceRepository) Withdraw(ctx context.Context, operation entity.BalanceOperation) error {
	const op = "infrastructure.postgre.BalanceRepository.Withdraw"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", operation.UserUUID.String()),
		r.log.AnyField("order_number", operation.OrderNumber),
		r.log.AnyField("accrual", operation.Accrual),
		r.log.AnyField("withdrawal", operation.Withdrawal),
	)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("Failed to begin transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	var current float64
	var withdrawn float64
	balance := tx.QueryRow(ctx, `SELECT current, withdraw FROM user_balances WHERE user_uuid = $1 FOR UPDATE`, operation.UserUUID)
	err = balance.Scan(&current, &withdrawn)
	if err != nil {
		log.Error("Failed to get current balance", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	if current < operation.Withdrawal {
		log.Error("Insufficient funds in the account")
		r.rollback(tx, ctx)
		return entity.ErrBalanceInsufficientFunds
	}

	newBalance := current - operation.Withdrawal
	_, err = tx.Exec(ctx, `UPDATE user_balances SET current = $1, withdraw = withdraw + $2 WHERE user_uuid = $3`, newBalance, operation.Withdrawal, operation.UserUUID)
	if err != nil {
		log.Error("Failed to update balance", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO balance_operations (
                                user_uuid, 
                                accrual, 
                                withdrawal, 
                                order_number, 
                                processed_at, 
                                uuid
                                ) VALUES ($1, $2, $3, $4, $5, $6)`,
		operation.UserUUID,
		operation.Accrual,
		operation.Withdrawal,
		operation.OrderNumber,
		operation.ProcessedAt,
		operation.UUID)
	if err != nil {
		log.Error("Failed to create balance operation", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *BalanceRepository) Accrue(ctx context.Context, operation entity.BalanceOperation) error {
	const op = "infrastructure.postgre.BalanceRepository.Accrue"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", operation.UserUUID.String()),
		r.log.AnyField("order_number", operation.OrderNumber),
		r.log.AnyField("accrual", operation.Accrual),
		r.log.AnyField("withdrawal", operation.Withdrawal),
	)

	tx, err := r.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		log.Error("Failed to begin transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	var current float64
	balance := tx.QueryRow(ctx, `SELECT current FROM user_balances WHERE user_uuid = $1 FOR UPDATE`, operation.UserUUID)
	err = balance.Scan(&current)
	if err != nil {
		log.Error("Failed to get current balance", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = tx.Exec(ctx, `UPDATE user_balances SET current = current + $1 WHERE user_uuid = $2`, operation.Accrual, operation.UserUUID)
	if err != nil {
		log.Error("Failed to update balance", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}
	_, err = tx.Exec(ctx, `INSERT INTO balance_operations (
                                user_uuid, 
                                accrual, 
                                withdrawal, 
                                order_number, 
                                processed_at, 
                                uuid
                                ) VALUES ($1, $2, $3, $4, $5, $6)`,
		operation.UserUUID,
		operation.Accrual,
		operation.Withdrawal,
		operation.OrderNumber,
		operation.ProcessedAt,
		operation.UUID)
	if err != nil {
		log.Error("Failed to create balance operation", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}
	err = tx.Commit(ctx)
	if err != nil {
		log.Error("Failed to commit transaction", log.ErrorField(err))
		r.rollback(tx, ctx)
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *BalanceRepository) rollback(tx pgx.Tx, ctx context.Context) {
	const op = "infrastructure.postgre.BalanceRepository.rollback"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
	)
	err := tx.Rollback(ctx)
	if err != nil {
		log.Error("Failed to rollback transaction", log.ErrorField(err))
	}
}
