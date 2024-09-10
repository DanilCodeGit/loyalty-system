package service

import (
	"context"

	"github.com/google/uuid"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// BalanceRepository is an interface for balance repository.
type BalanceRepository interface {
	GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error)
	GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error)
	Withdraw(ctx context.Context, operation entity.BalanceOperation) error
	Accrue(ctx context.Context, operation entity.BalanceOperation) error
	CreateBalance(ctx context.Context, userUUID uuid.UUID) error
}

// BalanceService is a service for managing balances.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=BalanceService
type BalanceService struct {
	repository BalanceRepository
	logger     *logger.Logger
}

// NewBalanceService returns a new balance service.
func NewBalanceService(logger *logger.Logger, repository BalanceRepository) *BalanceService {
	return &BalanceService{
		repository: repository,
		logger:     logger,
	}
}

// GetBalance returns the balance of a user.
func (s *BalanceService) GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error) {
	const op = "domain.services.BalanceService.GetBalance"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	balance, err := s.repository.GetBalance(ctx, userUUID)
	if err != nil {
		log.Error("Failed to get balance", log.ErrorField(err))
		return nil, err
	}

	log.Info("Balance retrieved")
	return balance, nil
}

// Execute executes a balance operation.
func (s *BalanceService) Execute(ctx context.Context, operation entity.BalanceOperation) error {
	const op = "domain.services.BalanceService.Execute"
	log := s.logger.With(
		s.logger.StringField("op", op),
		s.logger.StringField("request_id", contexter.GetRequestID(ctx)),
	)

	if operation.Withdrawal > 0 {
		log.Info("Executing withdrawal")
		return s.repository.Withdraw(ctx, operation)
	}
	if operation.Accrual > 0 {
		log.Info("Executing accrual")
		return s.repository.Accrue(ctx, operation)
	}
	return nil
}

// GetWithdrawOperations returns all withdrawal operations for a user.
func (s *BalanceService) GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error) {
	operations, err := s.repository.GetWithdrawOperations(ctx, userUUID)
	if err != nil {
		return nil, err
	}
	return operations, nil
}

func (s *BalanceService) CreateBalanceForUser(ctx context.Context, userUUID uuid.UUID) error {
	err := s.repository.CreateBalance(ctx, userUUID)
	if err != nil {
		return err
	}
	return nil
}
