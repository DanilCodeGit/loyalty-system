package logger

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5/middleware"

	"loyalty/internal/lib/logger"
)

// New returns a new http.Handler that logs requests.
func New(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const op = "internal.app.http-server.middleware.logger.New"
		logWith := log.With(
			log.StringField("op", op),
		)

		log.Info("logger middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			entry := logWith.With(
				log.StringField("method", r.Method),
				log.StringField("path", r.URL.Path),
				log.StringField("remote_addr", r.RemoteAddr),
				log.StringField("user_agent", r.UserAgent()),
				log.StringField("request_id", middleware.GetReqID(r.Context())),
			)
			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			t1 := time.Now()
			defer func() {
				entry.Info("request completed",
					log.AnyField("status", ww.Status()),
					log.AnyField("bytes", ww.BytesWritten()),
					log.StringField("duration", time.Since(t1).String()),
				)
			}()

			next.ServeHTTP(ww, r)
		}

		return http.HandlerFunc(fn)
	}
}
