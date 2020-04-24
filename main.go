package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"github.com/spf13/viper"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sduwh/vcode-judger/channel"
	"github.com/sduwh/vcode-judger/config"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sduwh/vcode-judger/remotejudger"
	"github.com/sirupsen/logrus"
)

func main() {
	// set time location
	loc, _ := time.LoadLocation("Asia/Chongqing")

	// config init
	configName := flag.String("configName", "config", "config file's name.")
	configPath := flag.String("configPath", "./config", "config file's path.")
	flag.Parse()
	config.ConfInit(configName, configPath)

	// log init
	config.LogInit()
	logrus.Printf("时区: %s\n", loc)

	// redis config
	redisAddr := viper.GetString("redis.host") + ":" + viper.GetString("redis.port")
	redisPass := viper.GetString("redisPassword")

	// main flow
	// remote task queue
	remoteTaskChannel, err := channel.NewRedisChannel(redisAddr, redisPass)
	if err != nil {
		logrus.WithError(err).Fatal("Create remote task channel")
	}
	// task result queue
	statusChannel, err := channel.NewRedisChannel(redisAddr, redisPass)
	if err != nil {
		logrus.WithError(err).Fatal("Create status channel channel")
	}
	// new remoteJudger
	remoteJudger, err := remotejudger.NewRemoteJudger()
	if err != nil {
		logrus.WithError(err).Fatal("Create remote remoteJudger")
	}
	// start listen task queue and judge task
	remoteTaskChannel.Listen(consts.TopicRemoteTask, &RemoteTaskListener{
		judger: remoteJudger,
		listener: &RemoteJudgeListener{
			statusChannel: statusChannel,
		},
	})

	// exit the program
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
	// create new task by message
	if err := json.Unmarshal(message, task); err != nil {
		logrus.WithError(err).Error("Unmarshal json")
		return
	}
	decodeBytes, err := base64.StdEncoding.DecodeString(task.Code)
	if err != nil {
		logrus.WithError(err).Errorf("base64 decode fail")
	}
	task.Code = string(decodeBytes)
	logrus.Debugf("getTask:%+v", task)
	logrus.WithFields(logrus.Fields{
		"ID":         task.ID,
		"ProviderOJ": task.ProviderOJ,
		"ProviderID": task.ProviderID,
		"Language":   task.Language,
	}).Info("New remote task")
	// start judge task
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
	if status.CompileError != "" {
		status.CompileError = base64.StdEncoding.EncodeToString([]byte(status.CompileError))
	}

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

func (l *RemoteJudgeListener) OnComplete(judgeTaskId string) {
	logrus.Infof("Task: %+v Judge done", judgeTaskId)
}
