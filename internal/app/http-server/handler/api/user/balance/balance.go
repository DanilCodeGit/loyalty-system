package balance

import (
	"context"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/google/uuid"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// UserBalanceGetter is an interface for getting an balance from the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserBalanceGetter
type UserBalanceGetter interface {
	GetBalance(ctx context.Context, userUUID uuid.UUID) (*entity.Balance, error)
}

// UserAuthorizer is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

// New  returned func for getting an balance from the user.
//
//	@Tags			Balance
//	@Summary		Получение баланса пользователя.
//	@Description	Эндпоинт используется для получения текущего балaнаса пользователя.
//	@Description	В заголовке Authorization необходимо передавать JWT токен.
//	@Produce		json
//	@Accept			plain
//	@Router			/user/balance [get]
//	@Param			Authorization	header		string				true	"JWT Token"
//	@Success		200				{object}	balance.Response	"User balance successfully returned"
//	@Failure		401				"User is not authorized"
//	@Failure		500				"Internal server error"
func New(log *logger.Logger, getter UserBalanceGetter, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.balance.New"

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

		balance, err := getter.GetBalance(ctx, user.UUID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		result := Response{
			Current:  balance.Current,
			Withdraw: balance.Withdraw,
		}

		render.JSON(w, r, result)
		logWith.Info("Balance fetched")
		w.WriteHeader(http.StatusOK)
	}
}

// Response is a response for getting an balance from the user.
type Response struct {
	Current  float64 `json:"current" example:"500.5"`
	Withdraw float64 `json:"withdrawn,omitempty" example:"42"`
}
