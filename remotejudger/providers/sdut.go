package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"time"
)

type ProviderSDUT struct {
	client         *http.Client
	languages      map[string]string
	statuses       map[string]string
	accounts       []*account
	currentAccount *account
}

func NewProviderSDUT() (*ProviderSDUT, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &ProviderSDUT{
		client: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns:        500,
				MaxIdleConnsPerHost: 100,
				IdleConnTimeout:     2 * time.Minute,
			},
			Jar:     jar,
			Timeout: 10 * time.Second,
		},
		languages: map[string]string{
			consts.LanguageC:    "gcc",
			consts.LanguageCPP:  "g++",
			consts.LanguageJava: "java",
		},
		accounts: []*account{
			{username: "jinchengwu001", password: "jcw666"},
		},
	}, nil
}

func (s *ProviderSDUT) Login() error {
	s.currentAccount = s.accounts[rand.Intn(len(s.accounts))]
	loginDataMap := make(map[string]string)
	loginDataMap["loginName"] = s.currentAccount.username
	loginDataMap["password"] = s.currentAccount.password
	loginData, _ = json.Marshal(loginDataMap)
	nowTimeStamp := time.Now().Unix()
	loginUrl := fmt.Sprintf("https://acm.sdut.edu.cn/onlinejudge2/index.php/API_ng/session?_t=%d", nowTimeStamp)
	loginRequest, err := http.NewRequest("POST",
		loginUrl,
		bytes.NewBuffer(loginDataMap))
	return nil
}

func (s *ProviderSDUT) HasLogin() (bool, error) {
	return false, nil
}

func (s *ProviderSDUT) Submit(task *models.RemoteJudgeTask) (string, error) {
	return "", nil
}

func (s *ProviderSDUT) Status(task *models.RemoteJudgeTask, submitID string) (*models.JudgeStatus, error) {
	return nil, ErrSubmissionNotFound
}

func (s *ProviderSDUT) FetchCompileError(submitID string) (string, error) {
	return "", nil
}
