package providers

import (
	"encoding/base64"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sirupsen/logrus"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"
)

type ProviderHDU struct {
	client         *http.Client
	languages      map[string]string
	statuses       map[string]string
	accounts       []*account
	currentAccount *account
}

func NewProviderHDU() (*ProviderHDU, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	return &ProviderHDU{
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
			consts.LanguageC:    "3",
			consts.LanguageCPP:  "2",
			consts.LanguageJava: "5",
		},
		accounts: []*account{
			{username: "jinchengwu001", password: "jcw666"},
		},
	}, nil
}

func (h *ProviderHDU) Login() error {
	h.currentAccount = h.accounts[rand.Intn(len(h.accounts))]
	data := url.Values{
		"username": []string{h.currentAccount.username},
		"userpass": []string{h.currentAccount.password},
		"login":    []string{"Sign In"},
	}
	loginRequest, err := http.NewRequest("POST",
		"http://acm.hdu.edu.cn/userloginex.php?action=login",
		strings.NewReader(data.Encode()))
	if err != nil {
		return err
	}
	loginRequest.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	loginRequest.Header.Set("Host", "acm.hdu.edu.cn")
	loginRequest.Header.Set("Referer", "http://acm.hdu.edu.cn/")
	loginRequest.Header.Set("Origin", "http://acm.hdu.edu.cn")

	resp, err := h.client.Do(loginRequest)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil
	}
	text := doc.Find("table").First().
		Find("tr").First().Next().
		Find("table").First().
		Find("tr").First().Next().
		Find("td").Last().
		Find("a").First().Text()
	logrus.Debugf("login: username %+v", text)
	if text != h.currentAccount.username {
		return ErrLoginFailed
	}
	return nil
}

func (h *ProviderHDU) HasLogin() (bool, error) {
	if h.currentAccount == nil {
		return false, nil
	}

	resp, err := h.client.Get("http://acm.hdu.edu.cn/")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, err
	}

	text := doc.Find("table").First().
		Find("tr").First().Next().
		Find("table").First().
		Find("tr").First().Next().
		Find("td").Last().
		Find("a").First().Text()
	logrus.Debugf("check login: username %+v", text)
	return text == h.currentAccount.username, nil
}

func (h *ProviderHDU) Submit(task *models.RemoteJudgeTask) (string, error) {
	id := task.ProviderID
	language, ok := h.languages[strings.ToUpper(task.Language)]
	if !ok {
		return "", ErrLanguageNotSupported
	}

	resp, err := h.client.PostForm("http://acm.hdu.edu.cn/submit.php?action=submit", url.Values{
		"check":     []string{"0"},
		"problemid": []string{id},
		"language":  []string{language},
		"usercode":  []string{task.Code},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resp, err = h.client.Get(
		fmt.Sprintf("http://acm.hdu.edu.cn/status.php?first=&pid=&user=%s&lang=%s&status=0",
			h.currentAccount.username,
			task.Language))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	submitID := doc.Find("table").First().
		Find("table").First().
		Find("tr").First().Next().
		Find("td").First().Text()
	if submitID == "" {
		return "", ErrSubmissionNotFound
	}
	logrus.Debugf("submitId: %v", submitID)
	return submitID, nil
}

func (h *ProviderHDU) Status(task *models.RemoteJudgeTask, submitID string) (*models.JudgeStatus, error) {
	resp, err := h.client.Get(
		fmt.Sprintf("http://acm.hdu.edu.cn/status.php?first=%s&pid=&user=&lang=0&status=0", submitID))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	curRow := doc.Find("table").First().
		Find("table").First().
		Find("tr").First().Next()
	for {
		curSubmitID := curRow.Find("td").First()
		if curSubmitID.Text() == "" {
			break
		}
		if curSubmitID.Text() != submitID {
			curRow = curRow.Next()
			continue
		}
		remoteStatus := curSubmitID.Next().Next()
		timeUsed := remoteStatus.Next().Next()
		memoryUsed := timeUsed.Next()
		status := &models.JudgeStatus{
			TaskID:   task.ID,
			SubmitID: submitID,
			Status:   strings.TrimSpace(remoteStatus.Text()),
		}
		logrus.Debugf("hdu judge status: %+v", status)
		if status.Status == "" {
			return nil, ErrStatusNotFound
		}
		if status.Status == "Compilation Error" {
			status.CompileError, err = h.FetchCompileError(submitID)
			if err != nil {
				logrus.WithError(err).Error("Fetch compile error")
			}
		}
		_, _ = fmt.Sscanf(timeUsed.Text(), "%dMS", &status.TimeUsed)
		_, _ = fmt.Sscanf(memoryUsed.Text(), "%dB", &status.TimeUsed)
	}
	return nil, nil
}

func (h *ProviderHDU) FetchCompileError(submitID string) (string, error) {
	resp, err := h.client.Get(fmt.Sprintf("http://acm.hdu.edu.cn/viewerror.php?rid=%s", submitID))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", nil
	}
	compileError := []byte(doc.Find("pre").Text())
	return base64.StdEncoding.EncodeToString(compileError), nil
}
