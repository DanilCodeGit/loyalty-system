package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"

	"loyalty/internal/domain/entity"
	"loyalty/internal/domain/tool"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// UserRepository is an interface for user repository.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserRepository
type UserRepository interface {
	GetUserByLogin(ctx context.Context, login string) (*entity.User, error)
	CreateUser(ctx context.Context, user *entity.User) (*entity.User, error)
	GetUserByUUID(ctx context.Context, userUUID uuid.UUID) (*entity.User, error)

	// Transaction support
	BeginTx(ctx context.Context) (context.Context, error)
	CommitTx(ctx context.Context) error
	RollbackTx(ctx context.Context) error
}

//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=BalanceCreator
type BalanceCreator interface {
	CreateBalanceForUser(ctx context.Context, userUUID uuid.UUID) error
}

// UserService is a service for managing users.
type UserService struct {
	repository     UserRepository
	secretKey      string
	logger         *logger.Logger
	balanceService BalanceCreator
}

// NewUserService returns a new user service.
func NewUserService(repository UserRepository, balanceService BalanceCreator, logger *logger.Logger, secretKey string) *UserService {
	return &UserService{
		repository:     repository,
		secretKey:      secretKey,
		logger:         logger,
		balanceService: balanceService,
	}
}

// Register register a new user.
func (s *UserService) Register(ctx context.Context, login, password string) (string, error) {
	// Начинаем транзакцию
	txCtx, err := s.repository.BeginTx(ctx)
	if err != nil {
		return "", err
	}

	// Генерация хеша пароля
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		err := s.repository.RollbackTx(txCtx)
		if err != nil {
			return "", err
		}
		return "", err
	}

	// Создание пользователя
	user := entity.NewUser(login, string(passwordHash), "", uuid.Nil)
	user, err = s.repository.CreateUser(txCtx, user)
	if err != nil {
		err := s.repository.RollbackTx(txCtx)
		if err != nil {
			return "", err
		}
		return "", err
	}

	// Создание баланса для пользователя
	err = s.balanceService.CreateBalanceForUser(txCtx, user.UUID)
	if err != nil {
		err := s.repository.RollbackTx(txCtx)
		if err != nil {
			return "", err
		}
		return "", err
	}

	// Коммит транзакции
	err = s.repository.CommitTx(txCtx)
	if err != nil {
		return "", err
	}

	// Создание JWT
	jwtString, err := tool.CreateJWT(user.UUID, s.secretKey)
	if err != nil {
		return "", err
	}

	return jwtString, nil
}

// Authenticate authorize a user.
func (s *UserService) Authenticate(ctx context.Context, login, password string) (string, error) {
	const op = "domain.services.UserService.Authorize"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	user, err := s.repository.GetUserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, entity.ErrUserNotFound) {
			return "", entity.ErrUserWrongPasswordOrLogin
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		log.Info("Wrong password", log.ErrorField(err))
		return "", entity.ErrUserWrongPasswordOrLogin
	}

	jwtString, err := tool.CreateJWT(user.UUID, s.secretKey)
	if err != nil {
		log.Error("Failed to create JWT", log.ErrorField(err))
		return "", err
	}

	return jwtString, nil
}

// Authorize authenticates a user.
func (s *UserService) Authorize(ctx context.Context, token string) (*entity.User, error) {
	const op = "domain.services.UserService.Authenticate"
	log := s.logger.With(s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	userUUID, err := tool.CheckJWT(token, s.secretKey)
	if err != nil || userUUID == uuid.Nil {
		log.Info("Invalid JWT", log.ErrorField(err))
		return nil, err
	}

	user := entity.NewUser("", "", token, userUUID)
	return user, nil
}
