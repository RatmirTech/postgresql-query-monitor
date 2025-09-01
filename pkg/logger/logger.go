package logger

import "go.uber.org/zap"

var logger *zap.Logger

func Init() {
	var err error

	logger, err = zap.NewProduction()
	if err != nil {
		panic(err)
	}
}

func GetLogger() *zap.Logger {
	return logger
}
