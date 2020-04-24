package providers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"strconv"
	"strings"
	"time"
)

type ProviderSDUT struct {
	client         *http.Client
	languages      map[string]string
	statuses       map[string]string
	accounts       []*account
	currentAccount *account
}
type loginResponseData struct {
	UserId   int    `json:"userId"`
	Username string `json:"username"`
	Nickname string `json:"nickname"`
}

type loginResponse struct {
	Success bool              `json:"success"`
	Code    int               `json:"code"`
	Data    loginResponseData `json:"data"`
}

type submitResponseData struct {
	SolutionId int `json:"solutionId"`
}

type submitResponse struct {
	Success bool               `json:"success"`
	Code    int                `json:"code"`
	Data    submitResponseData `json:"data"`
}

var judgeResultMap = map[int]string{
	0: consts.StatusJudging,
	1: consts.StatusAccepted,
	2: consts.StatusTimeLimitExceeded,
	3: consts.StatusMemoryLimitExceeded,
	4: consts.StatusWrongAnswer,
	5: consts.StatusRuntimeError,
	6: consts.StatusOutputLimitExceeded,
	7: consts.StatusCompileError,
	8: consts.StatusPresentationError,
}

type statusResponseData struct {
	Memory      int64  `json:"memory"`
	Time        int64  `json:"time"`
	Result      int    `json:"result"`
	SolutionId  int    `json:"solutionId"`
	CompileInfo string `json:"compileInfo"`
}

type statusResponse struct {
	Code    int                `json:"code"`
	Success bool               `json:"success"`
	Data    statusResponseData `json:"data"`
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
	loginData, _ := json.Marshal(loginDataMap)
	loginUrl := fmt.Sprintf("https://acm.sdut.edu.cn/onlinejudge2/index.php/API_ng/session?_t=%d", time.Now().Unix())
	loginRequest, err := http.NewRequest("POST",
		loginUrl,
		bytes.NewBuffer(loginData))
	if err != nil {
		return err
	}
	loginRequest.Header.Set("content-type", "application/json;charset=UTF-8")
	resp, err := s.client.Do(loginRequest)
	if err != nil {
		return err
	}
	bodyData, _ := ioutil.ReadAll(resp.Body)
	responseData := loginResponse{}
	err = json.Unmarshal(bodyData, &responseData)
	if err != nil {
		return err
	}
	if responseData.Data.Username != s.currentAccount.username {
		return ErrLoginFailed
	}
	logrus.Debugf("[login] login success: username: %s", s.currentAccount.username)
	return nil
}

func (s *ProviderSDUT) HasLogin() (bool, error) {
	checkLoginUrl := fmt.Sprintf("https://acm.sdut.edu.cn/onlinejudge2/index.php/API_ng/session?_t=%d",
		time.Now().Unix())
	logrus.Debugf("[HasLogin] checkUrl: %s", checkLoginUrl)
	resp, err := s.client.Get(checkLoginUrl)
	if err != nil {
		return false, err
	}
	bodyData, _ := ioutil.ReadAll(resp.Body)
	logrus.Debugf("[HasLogin] body data: %s", string(bodyData))
	responseData := loginResponse{}
	err = json.Unmarshal(bodyData, &responseData)
	if err != nil {
		return false, err
	}
	logrus.Debugf("[HasLogin] responseData: %+v", responseData)
	return responseData.Success, nil
}

func (s *ProviderSDUT) Submit(task *models.RemoteJudgeTask) (string, error) {
	id := task.ProviderID
	language, ok := s.languages[strings.ToUpper(task.Language)]
	if !ok {
		return "", ErrLanguageNotSupported
	}
	submitDataMap := make(map[string]string)
	submitDataMap["code"] = task.Code
	submitDataMap["language"] = language
	submitDataMap["problemId"] = id
	submitData, _ := json.Marshal(submitDataMap)
	submitUrl := fmt.Sprintf("https://acm.sdut.edu.cn/onlinejudge2/index.php/API_ng/solutions?_t=%d",
		time.Now().Unix())
	submitRequest, err := http.NewRequest("POST", submitUrl, bytes.NewBuffer(submitData))
	if err != nil {
		return "", err
	}
	submitRequest.Header.Set("content-type", "application/json;charset=UTF-8")
	resp, err := s.client.Do(submitRequest)
	if err != nil {
		return "", err
	}
	bodyData, _ := ioutil.ReadAll(resp.Body)
	logrus.Debugf("[Submit] body data: %s", string(bodyData))
	responseData := submitResponse{}
	err = json.Unmarshal(bodyData, &responseData)
	if err != nil {
		return "", err
	}
	logrus.Debugf("[Submit] submitId: %+v", responseData.Data.SolutionId)
	submitID := ""
	if responseData.Success {
		submitID = strconv.Itoa(responseData.Data.SolutionId)
		logrus.Debugf("submitID: %v", submitID)
		return submitID, nil
	} else {
		return "", ErrSubmissionNotFound
	}
}

func (s *ProviderSDUT) Status(task *models.RemoteJudgeTask, submitID string) (*models.JudgeStatus, error) {
	statusUrl := fmt.Sprintf("https://acm.sdut.edu.cn/onlinejudge2/index.php/API_ng/solutions/%s?_t=%d",
		submitID,
		time.Now().Unix())
	resp, err := s.client.Get(statusUrl)
	if err != nil {
		return nil, err
	}
	bodyData, _ := ioutil.ReadAll(resp.Body)
	responseData := statusResponse{}
	logrus.Debugf("[Status] bodyData data: %s", string(bodyData))
	err = json.Unmarshal(bodyData, &responseData)
	if err != nil {
		return nil, err
	}
	logrus.Debugf("[Status] response data: %+v", responseData)
	if responseData.Success == false {
		return nil, ErrSubmissionNotFound
	}
	status := &models.JudgeStatus{
		TaskID:   task.ID,
		SubmitID: submitID,
		Status:   judgeResultMap[responseData.Data.Result],
	}
	if status.Status == consts.StatusCompileError {
		status.CompileError = responseData.Data.CompileInfo
		logrus.Debugf("[Status] compileInfo: %s", status.CompileError)
	}
	status.TimeUsed = responseData.Data.Time
	status.MemoryUsed = responseData.Data.Memory
	logrus.Debugf("status: %+v", status)
	return status, nil
}
