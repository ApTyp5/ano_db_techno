package logs

import (
	"go.uber.org/zap"
)

var logger *zap.Logger
var sugar *zap.SugaredLogger
var err error

func init() {
	logger, err = zap.NewDevelopment()
	if err != nil {
		panic(err)
	}

	sugar = logger.Sugar()
	sugar.Info("Logging started")
}

func Fatal(args ...interface{}) {
	//sugar.Fatal(args...)
}

func Info(args ...interface{}) {
	//sugar.Info(args...)
}

func Error(err error) {
	//sugar.Info("\nerror: ", "\nmore: ", err.Error(), "\nless: ", errors.Cause(err).Error())
}
