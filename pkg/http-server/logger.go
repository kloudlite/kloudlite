package httpServer

import (
	"net/http"
	"time"

	"github.com/kloudlite/api/pkg/logging"
)

type loggingResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewLoggingResponseWriter(w http.ResponseWriter) *loggingResponseWriter {
	return &loggingResponseWriter{w, http.StatusOK}
}

func (lrw *loggingResponseWriter) WriteHeader(code int) {
	lrw.statusCode = code
	lrw.ResponseWriter.WriteHeader(code)
}

func NewLoggingMiddleware(logger logging.Logger) HttpMiddleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			lrw := NewLoggingResponseWriter(w)

			timestart := time.Now()
			logger.Infof("👉 %s %s?%s", r.Method, r.URL.Path, r.URL.Query().Encode())
			defer func() {
				logger.Infof("👈 %d %s %s?%s took %.2fs", lrw.statusCode, r.Method, r.URL.Path, r.URL.Query().Encode(), time.Since(timestart).Seconds())
			}()

			next(w, r)
		}
	}
}
