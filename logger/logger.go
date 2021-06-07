package logger

import (
	"github.com/sirupsen/logrus"
	"go.uber.org/zap"
)

type Logger struct {
	*logrus.Logger
}

func New(level string, caller bool) (*Logger, error) {
	l, err := logrus.ParseLevel(level)
	if err != nil {
		return nil, err
	}
	logger := logrus.New()
	logger.SetLevel(l)
	logger.SetReportCaller(caller)
	return &Logger{Logger: logger}, nil
}

func NewZap(isDev bool) (*zap.Logger, error) {
	var logger *zap.Logger
	var err error
	if isDev {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	return logger, err
}
