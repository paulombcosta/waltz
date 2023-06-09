package log

import (
	"os"

	"go.uber.org/zap"
)

var Logger *zap.SugaredLogger

func Init() error {
	var logger *zap.Logger
	var err error
	profile := os.Getenv("PROFILE")
	if profile == "dev" {
		logger, err = zap.NewDevelopment()
	} else {
		logger, err = zap.NewProduction()
	}
	if err != nil {
		return err
	}
	Logger = logger.Sugar()
	return nil
}
