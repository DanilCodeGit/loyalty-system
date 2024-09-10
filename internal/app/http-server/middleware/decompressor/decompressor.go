package decompressor

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"

	"loyalty/internal/lib/logger"
)

// New returns a middleware function that decompresses the request body if the
// request contains the "Content-Encoding: gzip" header.
func New(log *logger.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		const op = "internal.app.http-server.middleware.decompressor.New"
		logWith := log.With(
			log.StringField("op", op),
		)

		logWith.Info("Decompressor middleware enabled")

		fn := func(w http.ResponseWriter, r *http.Request) {
			contentEncoding := r.Header.Get("Content-Encoding")
			sendsGzip := strings.Contains(contentEncoding, "gzip")

			if sendsGzip {
				cr, err := newCompressReader(r.Body)
				if err != nil {
					logWith.Error("Failed init decompressor", log.ErrorField(err))
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				r.Body = cr
				defer func(cr *compressReader) {

					err := cr.Close()
					if err != nil {
						logWith.Error("Failed closing compress reader", log.ErrorField(err))
					}
				}(cr)
			}

			next.ServeHTTP(w, r)
		}
		return http.HandlerFunc(fn)
	}
}

type compressReader struct {
	r  io.ReadCloser
	zr *gzip.Reader
}

func newCompressReader(r io.ReadCloser) (*compressReader, error) {
	zr, err := gzip.NewReader(r)
	if err != nil {
		return nil, err
	}

	return &compressReader{
		r:  r,
		zr: zr,
	}, nil
}

func (c *compressReader) Read(p []byte) (n int, err error) {
	return c.zr.Read(p)
}

func (c *compressReader) Close() error {
	if err := c.r.Close(); err != nil {
		return err
	}
	return c.zr.Close()
}
