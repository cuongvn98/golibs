package config

import (
	"errors"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"os"
	"reflect"
)

func Load(paths []string, name string, cfg interface{}) error {
	if reflect.ValueOf(cfg).Kind() != reflect.Ptr {
		return errors.New("required a pointer but receive a value")
	}
	if err := os.Setenv("TZ", "Asia/Ho_Chi_Minh"); err != nil {
		return err
	}
	viper.SetConfigName(name)
	viper.SetConfigType("yaml")
	for _, path := range paths {
		viper.AddConfigPath(path)
	}
	viper.SetDefault("log_level", logrus.DebugLevel)
	if err := viper.ReadInConfig(); err != nil {
		return err
	}
	if err := viper.Unmarshal(&cfg); err != nil {
		return err
	}
	return nil
}
