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

	taskListener, err := newRemoteTaskListener(newRemoteJudgeListener(statusChannel))
	if err != nil {
		logrus.WithError(err).Fatal("Create remote task listener")
	}
	remoteTaskChannel.Listen(consts.TopicRemoteTask, taskListener)

	logrus.Info("Started")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh
	logrus.Info("Stopped")
}

type remoteTaskListener struct {
	judger   remotejudger.RemoteJudger
	listener remotejudger.RemoteJudgeListener
}

func newRemoteTaskListener(listener remotejudger.RemoteJudgeListener) (*remoteTaskListener, error) {
	judger, err := remotejudger.NewRemoteJudger()
	if err != nil {
		return nil, err
	}
	return &remoteTaskListener{
		judger:   judger,
		listener: listener,
	}, nil
}

func (l *remoteTaskListener) OnNext(message []byte) {
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

func (l *remoteTaskListener) OnError(err error) {
	logrus.WithError(err).Error("Consume failed")
}

func (l *remoteTaskListener) OnComplete() {
	logrus.Info("Consume done")
}

type remoteJudgeListener struct {
	statusChannel channel.Channel
}

func newRemoteJudgeListener(statusChannel channel.Channel) *remoteJudgeListener {
	return &remoteJudgeListener{statusChannel: statusChannel}
}

func (l *remoteJudgeListener) OnStatus(status *models.JudgeStatus) {
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

func (l *remoteJudgeListener) OnError(err error) {
	logrus.WithError(err).Error("Judge failed")
}

func (l *remoteJudgeListener) OnComplete() {
	logrus.Info("Judge done")
}
