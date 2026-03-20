package logwrapper

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path"
	"path/filepath"
	"runtime"

	"go.opentelemetry.io/contrib/bridges/otelslog"
)

const (
	ONE_FUNC_UP       = 1
	DEFAULT_FORMATTER = " %s:%v"
)

type Logger struct {
	inner    slog.Logger
	format   string
	hostname string
}

func (m *Logger) Enabled(level slog.Level) bool {
	return m.inner.Enabled(context.Background(), level)
}

func lineInfo() string {
	FUNCTION_DEPTH := 3
	pc, file, line, ok := runtime.Caller(FUNCTION_DEPTH)
	if !ok {
		return ""
	}

	funcName := runtime.FuncForPC(pc).Name()

	pkgName := path.Dir(funcName)

	return fmt.Sprintf("%s/%s:%d", pkgName, path.Base(file), line)
}

func getSourceFilename() string {
	FUNCTION_DEPTH := 2
	_, filename, _, ok := runtime.Caller(FUNCTION_DEPTH)
	if !ok {
		return "Unknown file"
	}
	return filepath.Base(filename)
}

func NewLogger() Logger {
	rtn := Logger{
		inner:  *slog.New(otelslog.NewHandler(getSourceFilename())),
		format: DEFAULT_FORMATTER,
	}

	hostname, err := os.Hostname()
	if err != nil {
		rtn.Error("Error getting hostname, setting to", "hostname", hostname, "err", err)
	}
	rtn.hostname = hostname

	return rtn
}

func (m *Logger) keyValAid(msg string, keyVals ...any) string {
	composite := []byte(msg)
	for i := 0; i < len(keyVals); i += 2 {
		var pair string
		if i+1 >= len(keyVals) {
			pair = fmt.Sprintf(m.format, keyVals[i], "Value not provided")
		} else {
			pair = fmt.Sprintf(m.format, keyVals[i], keyVals[i+1])
		}
		composite = append(composite, []byte(pair)...)
	}
	return string(composite)
}

func (m *Logger) wrapperAid(
	callerLevel slog.Level, ctx context.Context,
	msg string, keyvals ...any,
) {
	var call func(string, ...any)
	if !slog.Default().Enabled(ctx, callerLevel) {
		return
	}
	switch callerLevel {

	case slog.LevelDebug:
		call = m.inner.Debug
	case slog.LevelInfo:
		call = m.inner.Info
	case slog.LevelWarn:
		call = m.inner.Warn
	case slog.LevelError:
		call = m.inner.Error

	default:
		return
	}
	call(
		m.keyValAid(msg, keyvals...),
		"host_name", m.hostname,
		"line_info", lineInfo(),
	)
}

// ================================= Wrappers =================================

func (m *Logger) Debug(msg string, keyvals ...any) {
	m.wrapperAid(slog.LevelDebug, context.Background(), msg, keyvals...)
}

func (m *Logger) Info(msg string, keyVals ...any) {
	m.wrapperAid(slog.LevelInfo, context.Background(), msg, keyVals...)
}

func (m *Logger) Warn(msg string, keyVals ...any) {
	m.wrapperAid(slog.LevelWarn, context.Background(), msg, keyVals...)
}

func (m *Logger) Error(msg string, keyVals ...any) {
	m.wrapperAid(slog.LevelError, context.Background(), msg, keyVals...)
}
