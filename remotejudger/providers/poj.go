package providers

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/sirupsen/logrus"
	"golang.org/x/exp/rand"
)

type ProviderPOJ struct {
	client         *http.Client
	languages      map[string]string
	statuses       map[string]string
	accounts       []*account
	currentAccount *account
}

func NewProviderPOJ() (*ProviderPOJ, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &ProviderPOJ{
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
			consts.LanguageC:    "1",
			consts.LanguageCPP:  "0",
			consts.LanguageJava: "2",
		},
		accounts: []*account{
			{username: "wuyanzu001", password: "wyz666"},
		},
	}, nil
}

func (p *ProviderPOJ) Login() error {
	p.currentAccount = p.accounts[rand.Intn(len(p.accounts))]

	resp, err := p.client.PostForm("http://poj.org/login", url.Values{
		"user_id1":  []string{p.currentAccount.username},
		"password1": []string{p.currentAccount.password},
		"B1":        []string{"login"},
		"url":       []string{"%2F"},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return err
	}

	text := doc.Find("table").
		First().
		Find("tr").
		Last().
		Find("td").
		Last().
		Find("a").
		First().
		Children().
		Text()
	if text != p.currentAccount.username {
		return ErrLoginFailed
	}
	return nil
}

func (p *ProviderPOJ) HasLogin() (bool, error) {
	resp, err := p.client.Get("http://poj.org/")
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return false, err
	}

	text := doc.Find("table").
		First().
		Find("tr").
		Last().
		Find("td").
		Last().
		Find("a").
		First().
		Children().
		Text()
	return text == p.currentAccount.username, nil
}

func (p *ProviderPOJ) Submit(task *models.RemoteJudgeTask) (string, error) {
	language, ok := p.languages[task.Language]
	if !ok {
		return "", ErrLanguageNotSupported
	}

	resp, err := p.client.PostForm("http://poj.org/submit", url.Values{
		"problem_id": []string{task.ProviderID},
		"language":   []string{language},
		"source":     []string{task.Code},
		"submit":     []string{"Submit"},
		"encoded":    []string{"1"},
	})
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	resp, err = p.client.Get("http://poj.org/status?user_id=" + p.currentAccount.username)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}

	submitID := doc.Find("table").
		Last().
		Find("tr").
		First().
		Next().
		Find("td").
		First().
		Text()
	if submitID == "" {
		return "", ErrSubmissionNotFound
	}
	return submitID, nil
}

func (p *ProviderPOJ) Status(submitID string) (*models.JudgeStatus, error) {
	resp, err := p.client.Get("http://poj.org/status?user_id=" + p.currentAccount.username)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return nil, err
	}

	curRow := doc.Find("table").Last().Find("tr").First().Next()
	for {
		curSubmitID := curRow.Find("td").First()
		if curSubmitID.Text() == "" {
			break
		}
		if curSubmitID.Text() != submitID {
			curRow = curRow.Next()
			continue
		}

		remoteStatus := curSubmitID.Next().Next().Next()
		memoryUsed := remoteStatus.Next()
		timeUsed := memoryUsed.Next()
		status := &models.JudgeStatus{
			SubmitID: submitID,
			Status:   strings.TrimSpace(remoteStatus.Text()),
		}
		if status.Status == "" {
			return nil, ErrStatusNotFound
		}
		if status.Status == "Compile Error" {
			status.CompileError, err = p.fetchCompileError(submitID)
			if err != nil {
				logrus.WithError(err).Error("Fetch compile error")
			}
		}

		_, _ = fmt.Sscanf(memoryUsed.Text(), "%dK", &status.MemoryUsed)
		_, _ = fmt.Sscanf(timeUsed.Text(), "%dMS", &status.TimeUsed)
		return status, nil
	}

	return nil, ErrSubmissionNotFound
}

func (p *ProviderPOJ) fetchCompileError(submitID string) (string, error) {
	resp, err := p.client.Get("http://poj.org/showcompileinfo?solution_id=" + submitID)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", err
	}
	return doc.Find("pre").Find("font").Text(), nil
}
