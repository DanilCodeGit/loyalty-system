package tests

import (
	"net/url"
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"

	"loyalty/config"
	"loyalty/internal/app/http-server/handler/api/user/register"
)

func TestUserRegistration(t *testing.T) {
	login := gofakeit.Email()
	password := gofakeit.Password(true, true, true, true, false, 12)
	cfg := config.MustLoadConfig()

	//No test for internal server error
	t.Run("Success", func(t *testing.T) {
		u := url.URL{
			Scheme: "http",
			Host:   cfg.Addr,
		}
		e := httpexpect.Default(t, u.String())
		e.POST("/api/user/register").
			WithJSON(register.Request{
				Login:    login,
				Password: password,
			}).
			Expect().
			Status(200).
			Headers().
			ContainsKey("Authorization")
	})
	t.Run("User already exists", func(t *testing.T) {
		u := url.URL{
			Scheme: "http",
			Host:   cfg.Addr,
		}
		e := httpexpect.Default(t, u.String())
		e.POST("/api/user/register").
			WithJSON(register.Request{
				Login:    login,
				Password: password,
			}).
			Expect().
			Status(409)
	})
	t.Run("Bad request", func(t *testing.T) {
		u := url.URL{
			Scheme: "http",
			Host:   cfg.Addr,
		}
		e := httpexpect.Default(t, u.String())
		e.POST("/api/user/register").
			WithJSON(register.Request{
				Login:    "",
				Password: password,
			}).
			Expect().
			Status(400)
	})
}
