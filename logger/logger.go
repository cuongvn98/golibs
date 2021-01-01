package logger

import "github.com/sirupsen/logrus"

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
