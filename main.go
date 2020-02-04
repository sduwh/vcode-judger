package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/sduwh/vcode-judger/channel"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sirupsen/logrus"
)

func main() {
	taskCh, err := channel.NewRedisChannel("127.0.0.1:6379")
	if err != nil {
		logrus.WithError(err).Fatal("Create redis channel")
	}

	remoteTaskCh, err := channel.NewRedisChannel("127.0.0.1:6379")
	if err != nil {
		logrus.WithError(err).Fatal("Create redis channel")
	}

	statusCh, err := channel.NewRedisChannel("127.0.0.1:6379")
	if err != nil {
		logrus.WithError(err).Fatal("Create redis channel")
	}

	taskCh.Listen(consts.TopicTask, &taskListener{statusCh: statusCh})
	remoteTaskCh.Listen(consts.TopicRemoteTask, &remoteTaskListener{statusCh: statusCh})

	logrus.Info("Started")

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	<-sigC

	logrus.Info("Stopped")
}
