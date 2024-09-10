package register

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/render"
	"github.com/go-playground/validator/v10"

	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/contexter"
	"loyalty/internal/lib/logger"
)

// UserRegistrar is an interface for user registration.
//
//go:generate go run github.com/vektra/mockery/v2@v2.28.2 --name=UserRegistrar
type UserRegistrar interface {
	Register(ctx context.Context, login, password string) (string, error)
}

// Request struct for HTTP Request in JSON
type Request struct {
	Login    string `json:"login" validate:"required" example:"test@test.com"`
	Password string `json:"password" validate:"required" example:"test_Password"`
}

// New returned func for registering a new user.
//
//	@Tags			User
//	@Summary		Регистрация нового пользователя.
//	@Description	Эндпоинт используется для регистрации нового пользователя.
//	@Description	Логин приводится к нижнему регистру на стороне сервера
//	@Description	В заголовке Authorization возвращается JWT токен авторизации
//	@Accept			json
//	@Produce		plain
//	@Router			/user/register [post]
//	@Param			Request	body	register.Request	true	"Register Request"
//	@Success		200		"User registered successfully"
//	@Failure		409		"User already exists"
//	@Failure		400		"Bad request"
//	@Failure		500		"Internal server error"
//	@Header			200		{string}	Authorization	"JWT Token"
func New(log *logger.Logger, service UserRegistrar) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const op = "app.http-server.handler.api.user.register.New"

		ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
		defer cancel()

		reqID := middleware.GetReqID(ctx)
		ctx = context.WithValue(ctx, contexter.RequestID, reqID)
		logWith := log.With(
			log.StringField("op", op),
			log.StringField("request_id", reqID),
		)

		var req Request
		err := render.DecodeJSON(r.Body, &req)
		if err != nil {
			logWith.Info("Failed to decode request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		//convert Login to lower case
		req.Login = strings.ToLower(req.Login)
		logWith.Info("Request decoded", log.AnyField("login", req.Login))

		if err := validator.New().Struct(req); err != nil {
			logWith.Info("Failed to validate request body", log.ErrorField(err))
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		jwtString, err := service.Register(ctx, req.Login, req.Password)
		if err != nil {
			if errors.Is(err, entity.ErrUserExists) {
				w.WriteHeader(http.StatusConflict)
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		w.Header().Set("Authorization", jwtString)
		logWith.Info("User registered")
		w.WriteHeader(http.StatusOK)
	}
}
