package logger

import (
	"proxy_forwarder/gost/core/logger"
)

var (
	nop = &nopLogger{}
)

func Nop() logger.Logger {
	return nop
}

type nopLogger struct{}

func (l *nopLogger) WithFields(fields map[string]any) logger.Logger {
	return l
}

func (l *nopLogger) Trace(args ...any) {
}

func (l *nopLogger) Tracef(format string, args ...any) {
}

func (l *nopLogger) Debug(args ...any) {
}

func (l *nopLogger) Debugf(format string, args ...any) {
}

func (l *nopLogger) Info(args ...any) {
}

func (l *nopLogger) Infof(format string, args ...any) {
}

func (l *nopLogger) Warn(args ...any) {
}

func (l *nopLogger) Warnf(format string, args ...any) {
}

func (l *nopLogger) Error(args ...any) {
}

func (l *nopLogger) Errorf(format string, args ...any) {
}

func (l *nopLogger) Fatal(args ...any) {
}

func (l *nopLogger) Fatalf(format string, args ...any) {
}

func (l *nopLogger) GetLevel() logger.LogLevel {
	return ""
}

func (l *nopLogger) IsLevelEnabled(level logger.LogLevel) bool {
	return false
}
