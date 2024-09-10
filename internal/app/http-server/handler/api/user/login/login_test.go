package login_test

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"loyalty/internal/app/http-server/handler/api/user/login"
	"loyalty/internal/app/http-server/handler/api/user/login/mocks"
	"loyalty/internal/domain/entity"
	"loyalty/internal/lib/logger"
)

func TestNew(t *testing.T) {
	jwt := "JWT_test"

	tests := []struct {
		name          string
		login         string
		password      string
		mockError     error
		incorrectJSON bool
		jwt           string
		statusCode    int
	}{
		{
			name:          "Login: Success",
			login:         "test_user",
			password:      "TestPassword",
			mockError:     nil,
			incorrectJSON: false,
			jwt:           jwt,
			statusCode:    http.StatusOK,
		},
		{
			name:          "Login: wrong password",
			login:         "test_user",
			password:      "TestPassword",
			mockError:     entity.ErrUserWrongPasswordOrLogin,
			incorrectJSON: false,
			statusCode:    http.StatusUnauthorized,
		},
		{
			name:          "Login: Incorrect JSON",
			incorrectJSON: true,
			statusCode:    http.StatusBadRequest,
		},
		{
			name:          "Login: Repository error",
			login:         "test_user",
			password:      "TestPassword",
			mockError:     errors.New("repository error"),
			incorrectJSON: false,
			statusCode:    http.StatusInternalServerError,
		},
	}
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			userAuthenticatorMock := mocks.NewUserAuthenticator(t)
			if !tc.incorrectJSON {
				userAuthenticatorMock.On("Authenticate", mock.Anything, tc.login, tc.password).
					Return(tc.jwt, tc.mockError).
					Once()
			}

			log := logger.NewLogger()

			handler := login.New(log, userAuthenticatorMock)

			var input string

			if tc.incorrectJSON != true {
				input = fmt.Sprintf(`{"login": "%s", "password": "%s"}`, tc.login, tc.password)
			} else {
				input = "incorrect JSON"
			}

			req, err := http.NewRequest(http.MethodPost, "/api/user/register", bytes.NewReader([]byte(input)))

			require.NoError(t, err)

			rr := httptest.NewRecorder()

			handler.ServeHTTP(rr, req)

			require.Equal(t, rr.Code, tc.statusCode)

			jwtResponse := rr.Header().Get("Authorization")

			if !tc.incorrectJSON {
				require.Equal(t, tc.jwt, jwtResponse)
			}

		})
	}
}
