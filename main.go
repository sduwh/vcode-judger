package main

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/sduwh/vcode-judger/channel"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sirupsen/logrus"
	"github.com/urfave/cli/v2"
)

const RedisAddr = "127.0.0.1:6379"

func main() {
	app := cli.NewApp()
	app.Name = "vcode-judger"
	app.Usage = "Judge service of VCode."
	app.Version = consts.Version
	app.Flags = []cli.Flag{
		&cli.StringFlag{
			Name:     "config",
			Aliases:  []string{"c"},
			Required: true,
		},
	}
	app.Action = func(c *cli.Context) error {
		statusCh, err := channel.NewRedisChannel(RedisAddr)
		if err != nil {
			return err
		}

		taskCh, err := channel.NewRedisChannel(RedisAddr)
		if err != nil {
			return err
		}

		remoteTaskCh, err := channel.NewRedisChannel(RedisAddr)
		if err != nil {
			return err
		}

		taskCh.Listen(consts.TopicTask, &taskListener{statusCh: statusCh})
		remoteTaskCh.Listen(consts.TopicRemoteTask, &remoteTaskListener{statusCh: statusCh})

		logrus.Info("Started")
		sigC := make(chan os.Signal, 1)
		signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
		<-sigC
		logrus.Info("Stopped")

		return nil
	}

	if err := app.Run(os.Args); err != nil {
		logrus.WithError(err).Fatal("Failed to start")
	}

	taskCh.Listen(consts.TopicTask, &taskListener{statusCh: statusCh})
	remoteTaskCh.Listen(consts.TopicRemoteTask, &remoteTaskListener{statusCh: statusCh})

	if out, err:=exec.Command("docker", "version").CombinedOutput(); err != nil {
		logrus.Fatalf("fail to connect docker: %v, %s", err, out)
	}

	logrus.Info("Started")

	sigC := make(chan os.Signal, 1)
	signal.Notify(sigC, syscall.SIGINT, syscall.SIGTERM)
	<-sigC

	logrus.Info("Stopped")
}
