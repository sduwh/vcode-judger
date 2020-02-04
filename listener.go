package main

import (
	"github.com/sduwh/vcode-judger/channel"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sirupsen/logrus"
)

type taskListener struct {
	statusCh channel.Channel
}

func (l *taskListener) OnNext(message []byte) {
	if err := l.statusCh.Push(consts.TopicStatus, []byte("ok")); err != nil {
		logrus.WithError(err).Error("Push message")
	}
}

func (l *taskListener) OnError(err error) {
	logrus.WithError(err).Error("Consume failed")
}

func (l *taskListener) OnComplete() {
	logrus.Info("Consume done")
	if err := l.statusCh.Close(); err != nil {
		logrus.WithError(err).Error("Close redis channel")
	}
}

type remoteTaskListener struct {
	statusCh channel.Channel
}

func (l *remoteTaskListener) OnNext(message []byte) {
	if err := l.statusCh.Push(consts.TopicStatus, []byte("ok")); err != nil {
		logrus.WithError(err).Error("Push message")
	}
}

func (l *remoteTaskListener) OnError(err error) {
	logrus.WithError(err).Error("Consume failed")
}

func (l *remoteTaskListener) OnComplete() {
	logrus.Info("Consume done")
	if err := l.statusCh.Close(); err != nil {
		logrus.WithError(err).Error("Close redis channel")
	}
}
