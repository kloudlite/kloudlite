package logging

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/charmbracelet/log"
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
	HttpLoggerOptions
}

type HttpLoggerOptions struct {
	// Logger, can be set if you want to use a different logger than the one provided by the HttpLogger
	Logger *slog.Logger

	// LogLevel, can be set if you want to use a different log level than the one provided by the HttpLogger
	// defaults to log.DebugLevel
	LogLevel    log.Level
	ShowHeaders bool
	SilentPaths []string
}

func NewHttpLogger(opts HttpLoggerOptions) *HttpLogger {
	if opts.Logger == nil {
		logger := log.NewWithOptions(os.Stderr, log.Options{Level: log.DebugLevel})
		opts.Logger = slog.New(logger)
	}
	return &HttpLogger{opts}
}

func (h *HttpLogger) Use(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for i := range h.SilentPaths {
			if r.URL.Path == h.SilentPaths[i] {
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

		// fmt.Fprintf(os.Stderr, "❯❯ %s %s\n", r.Method, route)
		h.Logger.Debug(fmt.Sprintf("❯❯ %s %s", r.Method, route))
		defer func() {
			h.Logger.Info(fmt.Sprintf("❮❮ [%d] %s %s took %.2fs", lrw.statusCode, r.Method, route, time.Since(timestart).Seconds()))
			// fmt.Fprintf(os.Stderr, "❮❮ %d %s %s took %.2fs", lrw.statusCode, r.Method, route, time.Since(timestart).Seconds())
		}()

		next.ServeHTTP(lrw, r)
	})
}
