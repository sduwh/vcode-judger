package providers

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sirupsen/logrus"
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

	_, err := h.client.PostForm("http://acm.hdu.edu.cn/userloginex.php?action=login", url.Values{
		"username": []string{h.currentAccount.username},
		"userpass": []string{h.currentAccount.password},
		"login":    []string{"Sign In"},
	})
	if err != nil {
		if !strings.Contains(err.Error(), "response missing Location header") {
			return err
		}
	}

	resp, err := h.client.Get("http://acm.hdu.edu.cn/")
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}
	text := strings.TrimSpace(
		doc.Find("table").First().
			Find("tr").First().Next().
			Find("table").First().
			Find("tr").First().Next().
			Find("td").Last().
			Find("a").First().Text())
	logrus.Debugf("[login] username: %+v target username: %+v", text, h.currentAccount.username)
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

	text := strings.TrimSpace(
		doc.Find("table").First().
			Find("tr").First().Next().
			Find("table").First().
			Find("tr").First().Next().
			Find("td").Last().
			Find("a").First().Text())
	logrus.Debugf("[check login] username: %+v, target username: %+v", text, h.currentAccount.username)
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
	// 查询时语言数值向大偏移一位
	searchLanguageInt, _ := strconv.Atoi(language)
	searchLanguage := strconv.Itoa(searchLanguageInt + 1)
	statusUrl := fmt.Sprintf("http://acm.hdu.edu.cn/status.php?first=&pid=&user=%s&lang=%s&status=0",
		h.currentAccount.username,
		searchLanguage)
	resp, err = h.client.Get(
		statusUrl)
	logrus.Debugf("[submit status url] %s", statusUrl)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	submitID := doc.Find("table").First().
		Find("tr").First().Next().Next().Next().
		Find("td").First().
		Find("table").First().
		Find("tr").First().Next().Next().
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
		Find("tr").First().Next().Next().Next().
		Find("table").First().
		Find("tr").First().Next().Next()
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
		logrus.Debugf("status: %+v", status)
		return status, nil
	}
	return nil, ErrSubmissionNotFound
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
