package orders

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
	"loyalty/internal/lib/luna"
)

// OrderAdder is an interface for adding new order to the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=OrderAdder
type OrderAdder interface {
	Add(ctx context.Context, orderNumber int, userUUID uuid.UUID) error
}

// UserAuthorizer is an interface for authorizing users.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserAuthorizer
type UserAuthorizer interface {
	Authorize(ctx context.Context, token string) (*entity.User, error)
}

// NewAdder  returned func for adding a new order to the user.
//
//	@Tags			Order
//	@Summary		Добавление нового заказа для начисления средств
//	@Description	Эндпоинт используется для добавления нового заказа для начисления средств.
//	@Description	В заголовке Authorization необходимо передавать JWT токен.
//	@Accept			plain
//	@Produce		plain
//	@Router			/api/user/orders [post]
//	@Param			Authorization	header	string	true	"JWT Token"
//	@Param			Order			body	integer	true	"Order Number"	example(123124551)
//	@Success		200				"Order already added from current user"
//	@Success		202				"Order successfully added to process"
//	@Failure		400				"Invalid request"
//	@Failure		401				"User is not authorized"
//	@Failure		409				"Order already added from another user"
//	@Failure		422				"Order number is not valid"
//	@Failure		500				"Internal server error"
func NewAdder(log *logger.Logger, adder OrderAdder, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.orders.NewAdder"

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

		var orderID int
		err = render.DecodeJSON(r.Body, &orderID)
		if err != nil {
			logWith.Info("Failed to decode request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if !luna.Valid(orderID) {
			logWith.Info("Invalid order number", log.AnyField("order_id", orderID))
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		err = adder.Add(ctx, orderID, user.UUID)
		if err != nil {
			if errors.Is(err, entity.ErrOrderAlreadyUploadedByAnotherUser) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			if errors.Is(err, entity.ErrOrderAlreadyUploaded) {
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		logWith.Info(
			"Order successfully added",
			log.AnyField("order_id", orderID),
			log.AnyField("user_uuid", user.UUID),
		)
		w.WriteHeader(http.StatusAccepted)
	}
}

// AllOrdersGetter is an interface for getting an orders from the user.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=AllOrdersGetter
type AllOrdersGetter interface {
	GetAll(ctx context.Context, userUUID uuid.UUID) ([]entity.Order, error)
}

// NewAllGetter  returned func for getting all orders from the user.
//
//	@Tags			Order
//	@Summary		Получение списка загруженных заказов
//	@Description	Эндпоинт для получение списка загруженных номеров заказов и информации по ним
//	@Description	В заголовке Authorization необходимо передавать JWT токен.
//	@Description	Номера заказа в выдаче должны быть отсортированы по времени загрузки от самых старых к самым новым. Формат даты — RFC3339.
//	@Description	Доступные статусы обработки расчётов:
//	@Description	NEW — заказ загружен в систему, но не попал в обработку;
//	@Description	PROCESSING — вознаграждение за заказ рассчитывается;
//	@Description	INVALID — система расчёта вознаграждений отказала в расчёте;
//	@Description	PROCESSED — данные по заказу проверены и информация о расчёте успешно
//	@Accept			plain
//	@Produce		json
//	@Router			/api/user/orders [get]
//	@Param			Authorization	header		string				true	"JWT Token"
//	@Success		200				{object}	[]orders.Response	"Successfully fetched orders"
//	@Success		204				"No content"
//	@Failure		401				"User is not authorized"
//	@Failure		500				"Internal server error"
func NewAllGetter(log *logger.Logger, getter AllOrdersGetter, authorizer UserAuthorizer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.orders.NewAllGetter"

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

		orders, err := getter.GetAll(ctx, user.UUID)
		if err != nil {
			if errors.Is(err, entity.ErrOrderNotFound) {
				w.WriteHeader(http.StatusNoContent)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		result := make([]Response, 0, len(orders))
		for _, t := range orders {
			order := Response{
				Number:     fmt.Sprintf("%d", t.Number),
				Status:     string(t.Status),
				Accrual:    t.Accrual,
				UploadedAt: t.UploadedAt.Format(time.RFC3339),
			}
			result = append(result, order)
		}

		render.JSON(w, r, result)
		logWith.Info("Orders successfully retrieved")
		w.WriteHeader(http.StatusOK)
	}
}

// Response is an order response.
type Response struct {
	Number     string  `json:"number" example:"123124551"`
	Status     string  `json:"status" example:"PROCESSING" enums:"NEW,PROCESSING,INVALID,PROCESSED"`
	Accrual    float64 `json:"accrual,omitempty" example:"500"`
	UploadedAt string  `json:"uploaded_at" example:"2020-12-10T15:15:45+03:00"`
}
