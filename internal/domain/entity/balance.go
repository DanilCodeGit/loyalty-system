package entity

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type Balance struct {
	UserUUID uuid.UUID
	Current  float64
	Withdraw float64
}

type BalanceOperation struct {
	UUID        uuid.UUID
	UserUUID    uuid.UUID
	Accrual     float64
	Withdrawal  float64
	OrderNumber int
	ProcessedAt time.Time
}

var (
	ErrBalanceInsufficientFunds  = errors.New("insufficient funds in the account")
	ErrBalanceOperationsNotFound = errors.New("balance operations not found")
)

func NewBalanceOperation(userUUID uuid.UUID, accrual, withdrawal float64, orderNumber int) BalanceOperation {
	var operation = BalanceOperation{}

	operation.UUID = uuid.New()
	operation.UserUUID = userUUID
	operation.Accrual = accrual
	operation.Withdrawal = withdrawal
	operation.OrderNumber = orderNumber
	operation.ProcessedAt = time.Now()
	return operation
}
