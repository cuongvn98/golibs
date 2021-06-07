package utils

import (
	"github.com/pkg/errors"
	"os"
	"os/signal"
	"syscall"
)

func GracefullShutdown() error {
	exitChan := make(chan os.Signal)
	signal.Notify(exitChan, syscall.SIGTERM, syscall.SIGINT)
	c := <-exitChan

	return errors.New(c.String())
}
