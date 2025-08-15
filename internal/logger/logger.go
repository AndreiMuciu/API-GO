package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync/atomic"
	"time"
)

type Level int32

const (
    Debug Level = 10
    Info  Level = 20
    Warn  Level = 30
    Error Level = 40
)

var currentLevel int32 = int32(Info)

func SetLevel(l Level) { atomic.StoreInt32(&currentLevel, int32(l)) }
func SetLevelFromString(s string) {
    switch strings.ToLower(strings.TrimSpace(s)) {
    case "debug": SetLevel(Debug)
    case "info", "": SetLevel(Info)
    case "warn", "warning": SetLevel(Warn)
    case "error": SetLevel(Error)
    default: SetLevel(Info)
    }
}

func levelEnabled(l Level) bool { return Level(atomic.LoadInt32(&currentLevel)) <= l }

type Fields map[string]interface{}

type ctxKey string

const requestIDKey ctxKey = "request_id"

func WithRequestID(ctx context.Context, id string) context.Context { return context.WithValue(ctx, requestIDKey, id) }
func RequestIDFrom(ctx context.Context) string {
    if v := ctx.Value(requestIDKey); v != nil {
        if s, ok := v.(string); ok { return s }
    }
    return ""
}

func log(l Level, msg string, fields Fields) {
    if !levelEnabled(l) { return }
    payload := map[string]interface{}{
        "ts": time.Now().Format(time.RFC3339Nano),
        "level": levelString(l),
        "msg": msg,
    }
    for k, v := range fields {
        payload[k] = v
    }
    b, err := json.Marshal(payload)
    if err != nil {
        fmt.Fprintf(os.Stdout, "{\"ts\":%q,\"level\":%q,\"msg\":%q}\n", time.Now().Format(time.RFC3339Nano), levelString(l), msg)
        return
    }
    os.Stdout.Write(append(b, '\n'))
}

func levelString(l Level) string {
    switch l {
    case Debug: return "debug"
    case Info: return "info"
    case Warn: return "warn"
    case Error: return "error"
    default: return "info"
    }
}

func Debugf(msg string, f Fields) { log(Debug, msg, f) }
func Infof(msg string, f Fields)  { log(Info, msg, f) }
func Warnf(msg string, f Fields)  { log(Warn, msg, f) }
func Errorf(msg string, f Fields) { log(Error, msg, f) }
