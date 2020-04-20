package remotejudger

import (
	"errors"
	"strings"
	"time"

	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sduwh/vcode-judger/remotejudger/providers"
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

	OnComplete()
}

type RemoteJudgeProvider interface {
	Login() error

	HasLogin() (bool, error)

	Submit(task *models.RemoteJudgeTask) (string, error)

	Status(task *models.RemoteJudgeTask, submitID string) (*models.JudgeStatus, error)
}

func NewRemoteJudger() (RemoteJudger, error) {
	poj, err := providers.NewProviderPOJ()
	if err != nil {
		return nil, err
	}
	return &remoteJudger{
		providers: map[string]RemoteJudgeProvider{
			consts.RemotePOJ: poj,
		},
	}, nil
}

type remoteJudger struct {
	providers map[string]RemoteJudgeProvider
}

func (r *remoteJudger) Judge(task *models.RemoteJudgeTask, listener RemoteJudgeListener) {
	defer listener.OnComplete()

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
		listener.OnStatus(status)

		if strings.Contains(status.Status, "ing") {
			time.Sleep(1 * time.Second)
			continue
		}
		return
	}
}
