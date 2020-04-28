package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"time"

	"github.com/sduwh/vcode-judger/channel"
	"github.com/sduwh/vcode-judger/config"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sduwh/vcode-judger/remotejudger"
	"github.com/sduwh/vcode-judger/web"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// TODO 添加sandbox，和c,c++,java,Python运行环境
// TODO 添加本地判题功能
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

	port := viper.GetString("port")

	router := web.GetRouter()
	logrus.Info("Router load success...")
	logrus.Info("Start web server...")
	if err := router.Run(port); err != nil {
		logrus.Panicf("Web server start fail: %s", err)
	}

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
