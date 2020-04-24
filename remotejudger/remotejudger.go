package remotejudger

import (
	"errors"
	"github.com/sirupsen/logrus"
	"strings"
	"time"

	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sduwh/vcode-judger/remotejudger/providers"
	"github.com/sduwh/vcode-judger/util"
)

var (
	ErrOJNotSupported = errors.New("OJ not supported")
)

type RemoteJudger interface {
	Judge(task *models.RemoteJudgeTask, listener RemoteJudgeListener)
}

type RemoteJudgeListener interface {
	OnStatus(status *models.JudgeStatus)

	OnError(err error)

	OnComplete(judgeTaskId string)
}

/*
	远程判题服务接口
*/
type RemoteJudgeProvider interface {
	/*
		登陆
	*/
	Login() error

	/*
		检查是否登陆
	*/
	HasLogin() (bool, error)

	/*
		提交代码并返回目标网站的submitID
	*/
	Submit(task *models.RemoteJudgeTask) (string, error)

	/*
		更具submitID去检查判题结果
	*/
	Status(task *models.RemoteJudgeTask, submitID string) (*models.JudgeStatus, error)
}

func NewRemoteJudger() (RemoteJudger, error) {
	poj, err := providers.NewProviderPOJ()
	if err != nil {
		return nil, err
	}
	hdu, err := providers.NewProviderHDU()
	if err != nil {
		return nil, err
	}
	sdut, err := providers.NewProviderSDUT()
	if err != nil {
		return nil, err
	}
	return &remoteJudger{
		providers: map[string]RemoteJudgeProvider{
			consts.RemotePOJ:  poj,
			consts.RemoteHDU:  hdu,
			consts.RemoteSDUT: sdut,
		},
	}, nil
}

type remoteJudger struct {
	providers map[string]RemoteJudgeProvider
}

func (r *remoteJudger) Judge(task *models.RemoteJudgeTask, listener RemoteJudgeListener) {
	defer listener.OnComplete(task.ID)
	// get origin oj
	provider, ok := r.providers[task.ProviderOJ]
	if !ok {
		listener.OnError(ErrOJNotSupported)
		return
	}

	hasLogin, err := provider.HasLogin()
	if err != nil {
		listener.OnError(err)
		return
	}

	if !hasLogin {
		if err := provider.Login(); err != nil {
			listener.OnError(err)
			return
		}
	}
	// submit the problem
	submitID, err := provider.Submit(task)
	if err != nil {
		listener.OnError(err)
		return
	}

	for {
		status, err := provider.Status(task, submitID)
		if err != nil {
			listener.OnError(err)
			return
		}
		// Modify results
		logrus.Debugf("[judge]status: %+v", status)
		status.Status = util.JudgeStatusMapper(status.Status)
		// send the result to queue
		listener.OnStatus(status)

		if strings.Contains(status.Status, "ing") {
			time.Sleep(1 * time.Second)
			continue
		}
		return
	}
}
