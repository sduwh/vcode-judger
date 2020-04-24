package config

import (
	rotateLogs "github.com/lestrrat-go/file-rotatelogs"
	"github.com/rifflock/lfshook"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"log"
	"os"
	"path"
	"time"
)

func ConfInit(configName *string, configPath *string) {
	viper.SetConfigName(*configName)
	viper.AddConfigPath(*configPath)
	redisPassword := os.Getenv("REDIS_PASSWORD")
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	viper.Set("redisPassword", redisPassword)
	viper.Set("mongoPassword", mongoPassword)
	if err := viper.ReadInConfig(); err != nil {
		panic(err)
	}
}

func LogInit() {
	logConf := viper.GetStringMap("log")
	baseLogPath := path.Join(logConf["path"].(string), logConf["name"].(string))
	logFile, err := os.OpenFile(baseLogPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0766)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(logFile)
	if viper.GetBool("debug") {
		logrus.SetLevel(logrus.DebugLevel)
	}else {
		logrus.SetLevel(logrus.InfoLevel)
	}
	writer, err := rotateLogs.New(
		baseLogPath+".%Y%m%d%h%M",
		rotateLogs.WithLinkName(baseLogPath),
		rotateLogs.WithRotationCount(5),
		rotateLogs.WithRotationTime(time.Hour),
	)
	if err != nil {
		log.Panic(err)
	}
	lfHook := lfshook.NewHook(lfshook.WriterMap{
		logrus.DebugLevel: writer,
		logrus.InfoLevel:  writer,
		logrus.WarnLevel:  writer,
		logrus.ErrorLevel: writer,
		logrus.FatalLevel: writer,
		logrus.PanicLevel: writer,
	}, &logrus.TextFormatter{})

	logrus.AddHook(lfHook)

	logrus.Info("log file load success……")
}
