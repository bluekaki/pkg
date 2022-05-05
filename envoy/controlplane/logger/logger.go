package logger

import (
	"github.com/envoyproxy/go-control-plane/pkg/log"
	"go.uber.org/zap"
)

var _ log.Logger = (*logger)(nil)

type logger struct {
	*zap.Logger
}

func (l *logger) Debugf(format string, args ...interface{}) {
	l.Logger.Sugar().Debugf(format, args...)
}

func (l *logger) Infof(format string, args ...interface{}) {
	l.Logger.Sugar().Infof(format, args...)
}

func (l *logger) Warnf(format string, args ...interface{}) {
	l.Logger.Sugar().Warnf(format, args...)
}

func (l *logger) Errorf(format string, args ...interface{}) {
	l.Logger.Sugar().Errorf(format, args...)
}

func New(_logger *zap.Logger) log.Logger {
	if _logger == nil {
		panic("logger required")
	}

	return &logger{Logger: _logger}
}
