package withdrawals

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// BalanceWithdrawOperationGetter is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=BalanceWithdrawOperationGetter
type BalanceWithdrawOperationGetter interface {
	GetWithdrawOperations(ctx context.Context, userUUID uuid.UUID) ([]entity.BalanceOperation, error)
}

// UserAuthorizer is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

// New  returned func for showing user's withdrawal operations.
//
//	@Tags			Balance
//	@Summary		Получение списка операций снятия баланса.
//	@Description	Эндпоинт используется для получения списка операций снятия баланса пользователя
//	@Description	В заголовке Authorization необходимо передавать JWT токен.
//	@Produce		json
//	@Accept			plain
//	@Router			/api/user/withdrawals [get]
//	@Param			Authorization	header		string					true	"JWT Token"
//	@Success		200				{object}	[]withdrawals.Response	"User balance successfully returned"
//	@Success		204				"No withdrawal operations found"
//	@Failure		401				"User is not authorized"
//	@Failure		500				"Internal server error"
func New(log *logger.Logger, getter BalanceWithdrawOperationGetter, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.withdrawals.New"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestID, reqID)
		logWith := log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		user, err := authorizer.Authorize(ctx, r.Header.Get("Authorization"))
		if err != nil {
			logWith.Info("Failed to authorize request", log.ErrorField(err))
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		balanceOperations, err := getter.GetWithdrawOperations(ctx, user.UUID)
		if err != nil {
			if errors.Is(err, entity.ErrBalanceOperationsNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result := make([]Response, 0, len(balanceOperations))
		for _, t := range balanceOperations {
			operation := Response{
				OrderNumber: fmt.Sprintf("%d", t.OrderNumber),
				Withdrawal:  t.Withdrawal,
				ProcessedAt: t.ProcessedAt.Format(time.RFC3339),
			}
			result = append(result, operation)
		}

		render.JSON(w, r, result)
		logWith.Info("Withdrawal operations fetched")
		w.WriteHeader(http.StatusOK)
	}
}

// Response is a response for showing user's withdrawal operations
type Response struct {
	Withdrawal  float64 `json:"sum" example:"100"`
	OrderNumber string  `json:"order" example:"12312455"`
	ProcessedAt string  `json:"processed_at" example:"2020-12-10T15:15:45+03:00"`
}
