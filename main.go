package main

import (
	"encoding/json"
	"os"
	"os/signal"
	"syscall"

	"github.com/sduwh/vcode-judger/channel"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sduwh/vcode-judger/remotejudger"
	"github.com/sirupsen/logrus"
)

const RedisAddr = "127.0.0.1:6379"

func main() {
	remoteTaskChannel, err := channel.NewRedisChannel(RedisAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Create remote task channel")
	}

	statusChannel, err := channel.NewRedisChannel(RedisAddr)
	if err != nil {
		logrus.WithError(err).Fatal("Create status channel channel")
	}

	judger, err := remotejudger.NewRemoteJudger()
	if err != nil {
		logrus.WithError(err).Fatal("Create remote judger")
	}

	remoteTaskChannel.Listen(consts.TopicRemoteTask, &RemoteTaskListener{
		judger: judger,
		listener: &RemoteJudgeListener{
			statusChannel: statusChannel,
		},
	})

	logrus.Info("Started")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logrus.Info("Stopped")
}

type RemoteTaskListener struct {
	judger   remotejudger.RemoteJudger
	listener remotejudger.RemoteJudgeListener
}

func (l *RemoteTaskListener) OnNext(message []byte) {
	task := &models.RemoteJudgeTask{}
	if err := json.Unmarshal(message, task); err != nil {
		logrus.WithError(err).Error("Unmarshal json")
		return
	}

	logrus.WithFields(logrus.Fields{
		"ID":         task.ID,
		"ProviderOJ": task.ProviderOJ,
		"ProviderID": task.ProviderID,
		"Language":   task.Language,
	}).Info("New remote task")

	l.judger.Judge(task, l.listener)
}

func (l *RemoteTaskListener) OnError(err error) {
	logrus.WithError(err).Error("Consume failed")
}

func (l *RemoteTaskListener) OnComplete() {
	logrus.Info("Consume done")
}

type RemoteJudgeListener struct {
	statusChannel channel.Channel
}

func (l *RemoteJudgeListener) OnStatus(status *models.JudgeStatus) {
	message, err := json.Marshal(status)
	if err != nil {
		logrus.WithError(err).Error("Marshal json")
		return
	}
	if err := l.statusChannel.Push(consts.TopicStatus, message); err != nil {
		logrus.WithError(err).Error("Push message")
		return
	}
}

func (l *RemoteJudgeListener) OnError(err error) {
	logrus.WithError(err).Error("Judge failed")
}

func (l *RemoteJudgeListener) OnComplete() {
	logrus.Info("Judge done")
}
