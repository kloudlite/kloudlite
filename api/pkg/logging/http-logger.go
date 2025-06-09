package logging

import (
	"fmt"
	"log/slog"
	"net/http"
	"time"
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

func (lrw *loggingResponseWriter) Flush() {
	flusher := lrw.ResponseWriter.(http.Flusher)
	flusher.Flush()
}

type HttpLogger struct {
	*slog.Logger
	opts HttpLoggerOptions
}

type HttpLoggerOptions struct {
	ShowQuery   bool
	ShowHeaders bool
	SilentPaths []string
}

func NewHttpLogger(logger *slog.Logger, opts HttpLoggerOptions) *HttpLogger {
	return &HttpLogger{logger, opts}
}

func (h *HttpLogger) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := range h.opts.SilentPaths {
			if r.URL.Path == h.opts.SilentPaths[i] {
				next.ServeHTTP(w, r)
				return
			}
		}

		lrw := NewLoggingResponseWriter(w)

		timestart := time.Now()

		route := r.URL.Path
		if r.URL.RawQuery != "" {
			route = fmt.Sprintf("%s?%s", route, r.URL.RawQuery)
		}

		h.Logger.Info(fmt.Sprintf("❯❯ %s %s", r.Method, route))
		defer func() {
			h.Logger.Info(fmt.Sprintf("❮❮ %d %s %s took %.2fs", lrw.statusCode, r.Method, route, time.Since(timestart).Seconds()))
		}()

		next.ServeHTTP(lrw, r)
	})
}
