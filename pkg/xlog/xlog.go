package xlog

import (
	"log/slog"
	"os"
	"runtime"
	"strconv"
)

var logger *slog.Logger

func init() {
	var level = slog.LevelDebug
	if v, err := strconv.Atoi(os.Getenv("LOG_LEVEL")); err == nil {
		level = slog.Level(v)

	}
	var opts = &slog.HandlerOptions{
		AddSource: true,
		Level:     level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				pc, f, l, _ := runtime.Caller(7)
				a.Value = slog.GroupValue(
					slog.Attr{
						Key:   "file",
						Value: slog.StringValue(f),
					},
					slog.Attr{
						Key:   "line",
						Value: slog.IntValue(l),
					},
					slog.Attr{
						Key:   "function",
						Value: slog.StringValue(runtime.FuncForPC(pc).Name()),
					},
				)
			}
			return a
		},
	}
	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger = slog.New(handler)
}

func Info(msg string, args ...any) {
	logger.Info(msg, args...)
}

func Debug(msg string, args ...any) {
	logger.Debug(msg, args...)
}

func Error(msg string, args ...any) {
	logger.Error(msg, args...)
}

func Warn(msg string, args ...any) {
	logger.Warn(msg, args...)
}
