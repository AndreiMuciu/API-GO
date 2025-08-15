package middleware

import (
	"API-GO/internal/logger"
	"net/http"
	"time"

	"github.com/google/uuid"
)

type statusWriter struct {
    http.ResponseWriter
    status int
    bytes  int
}

func (w *statusWriter) WriteHeader(code int) { w.status = code; w.ResponseWriter.WriteHeader(code) }
func (w *statusWriter) Write(b []byte) (int, error) {
    if w.status == 0 { w.status = http.StatusOK }
    n, err := w.ResponseWriter.Write(b)
    w.bytes += n
    return n, err
}

// RequestLogger logs method, path, status, duration, and request id.
func RequestLogger(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        rid := r.Header.Get("X-Request-Id")
        if rid == "" { rid = uuid.NewString() }
        r = r.WithContext(logger.WithRequestID(r.Context(), rid))
        sw := &statusWriter{ResponseWriter: w}
        next.ServeHTTP(sw, r)
        logger.Infof("http_request", logger.Fields{
            "request_id": rid,
            "method": r.Method,
            "path":   r.URL.Path,
            "status": sw.status,
            "bytes":  sw.bytes,
            "dur_ms": time.Since(start).Milliseconds(),
            "remote": r.RemoteAddr,
        })
    })
}
