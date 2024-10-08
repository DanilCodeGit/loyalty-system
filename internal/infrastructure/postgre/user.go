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

// UserRepository is an implementation of user repository.
type UserRepository struct {
	db  *pgxpool.Pool
	log *logger.Logger
}

// NewUserRepository returns a new postgre user repository
func NewUserRepository(db *pgxpool.Pool, log *logger.Logger) *UserRepository {
	storage := &UserRepository{db: db, log: log}
	return storage
}

// Migrate migrates the database
func (r *UserRepository) Migrate(ctx context.Context) error {
	const op = "infrastructure.postgre.UserRepository.Migrate"

	log := r.log.With(r.log.StringField("op", op))

	_, err := r.db.Exec(ctx, `CREATE TABLE IF NOT EXISTS users (
        uuid UUID PRIMARY KEY,
        login TEXT UNIQUE NOT NULL,
        password_hash TEXT NOT NULL);`)
	if err != nil {
		log.Error("Failed to create table", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// GetUserByLogin returns a user by login.
func (r *UserRepository) GetUserByLogin(ctx context.Context, login string) (*entity.User, error) {
	const op = "infrastructure.postgre.UserRepository.GetUserByLogin"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_login", login),
	)

	var user entity.User
	err := r.db.QueryRow(ctx, `SELECT uuid, login, password_hash FROM users WHERE login = $1`, login).Scan(&user.UUID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Info("User not found", log.StringField("user_login", login))
			return nil, entity.ErrUserNotFound
		}
		log.Error("Failed to get user by login", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &user, nil

}

// CreateUser creates a new user.
func (r *UserRepository) CreateUser(ctx context.Context, user *entity.User) (*entity.User, error) {
	const op = "infrastructure.postgre.UserRepository.CreateUser"

	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_login", user.Login),
	)
	_, err := r.db.Exec(ctx, `INSERT INTO users (uuid, login, password_hash) VALUES ($1, $2, $3)`, user.UUID, user.Login, user.PasswordHash)
	if err != nil {
		//check user already exists
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				log.Info("User already exists!")
				return nil, entity.ErrUserExists
			}
		}
		log.Error("Failed to create user", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return user, nil
}

// GetUserByUUID returns a user by UUID.
func (r *UserRepository) GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error) {
	const op = "infrastructure.postgre.UserRepository.GetUserByUUID"
	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
		r.log.StringField("user_uuid", userUUID.String()),
	)

	var user entity.User
	err := r.db.QueryRow(ctx, `SELECT uuid, login, password_hash FROM users WHERE uuid = $1`, userUUID).Scan(&user.UUID, &user.Login, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			log.Info("User not found", log.StringField("user_uuid", userUUID.String()))
			return nil, entity.ErrUserNotFound
		}
		log.Error("Failed to get user by UUID", log.ErrorField(err))
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &user, nil
}

// Определяем ключ для хранения транзакции в контексте
type txKeyType struct{}

var txKey = txKeyType{}

// BeginTx начинается транзакция и сохраняет её в контексте.
func (r *UserRepository) BeginTx(ctx context.Context) (context.Context, error) {
	const op = "infrastructure.postgre.UserRepository.BeginTx"

	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	tx, err := r.db.Begin(ctx)
	if err != nil {
		log.Error("Failed to begin transaction", log.ErrorField(err))
		return ctx, fmt.Errorf("%s: %w", op, err)
	}

	// Сохраняем транзакцию в контексте
	return context.WithValue(ctx, txKey, tx), nil
}

// CommitTx коммитит транзакцию, если она существует в контексте.
func (r *UserRepository) CommitTx(ctx context.Context) error {
	const op = "infrastructure.postgre.UserRepository.CommitTx"

	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		log.Error("No transaction found in context")
		return errors.New("transaction not found in context")
	}

	if err := tx.Commit(ctx); err != nil {
		log.Error("Failed to commit transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Transaction committed successfully")
	return nil
}

// RollbackTx откатывает транзакцию в случае ошибки.
func (r *UserRepository) RollbackTx(ctx context.Context) error {
	const op = "infrastructure.postgre.UserRepository.RollbackTx"

	log := r.log.With(
		r.log.StringField("op", op),
		r.log.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	tx, ok := ctx.Value(txKey).(pgx.Tx)
	if !ok {
		log.Error("No transaction found in context")
		return errors.New("transaction not found in context")
	}

	if err := tx.Rollback(ctx); err != nil && err != pgx.ErrTxClosed {
		log.Error("Failed to rollback transaction", log.ErrorField(err))
		return fmt.Errorf("%s: %w", op, err)
	}

	log.Info("Transaction rolled back successfully")
	return nil
}
