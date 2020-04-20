package remotejudger

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/stretchr/testify/assert"
)

func TestRemoteJudger_Judge(t *testing.T) {
	judger, err := NewRemoteJudger()
	assert.NoError(t, err)

	judger.Judge(&models.RemoteJudgeTask{
		ID:         "a",
		ProviderOJ: consts.RemotePOJ,
		ProviderID: "1000",
		Language:   consts.LanguageC,
		Code: base64.StdEncoding.EncodeToString([]byte(`
			// Code here
			int main() {
				aaa // Should compile error
				return 0;
			}
		`)),
	}, &mockRemoteJudgeListener{t: t})

	judger.Judge(&models.RemoteJudgeTask{
		ID:         "b",
		ProviderOJ: consts.RemotePOJ,
		ProviderID: "1000",
		Language:   consts.LanguageC,
		Code: base64.StdEncoding.EncodeToString([]byte(`
			// Code here
			# include <stdio.h>
			int main() {
				int a, b;    
				scanf("%d %d", &a, &b);
				printf("%d\n", a+b);
				return 0;
			}
		`)),
	}, &mockRemoteJudgeListener{t: t})
}

type mockRemoteJudgeListener struct {
	t *testing.T
}

func (l *mockRemoteJudgeListener) OnStatus(status *models.JudgeStatus) {
	if status.SubmitID == "a" {
		if strings.Contains(status.Status, "ing") {
			assert.NotEmpty(l.t, status.Status)
			assert.Empty(l.t, status.TimeUsed)
			assert.Empty(l.t, status.MemoryUsed)
			assert.Empty(l.t, status.CompileError)
		} else {
			assert.Equal(l.t, "Compile Error", status.Status)
			assert.Empty(l.t, status.TimeUsed)
			assert.Empty(l.t, status.MemoryUsed)
			assert.NotEmpty(l.t, status.CompileError)
		}
		return
	}

	if status.SubmitID == "b" {
		if strings.Contains(status.Status, "ing") {
			assert.NotEmpty(l.t, status.Status)
			assert.Empty(l.t, status.TimeUsed)
			assert.Empty(l.t, status.MemoryUsed)
			assert.Empty(l.t, status.CompileError)
		} else {
			assert.Equal(l.t, "Accepted", status.Status)
			assert.Empty(l.t, status.TimeUsed)
			assert.Empty(l.t, status.MemoryUsed)
			assert.NotEmpty(l.t, status.CompileError)
		}
		return
	}
}

func (l *mockRemoteJudgeListener) OnError(err error) {
	assert.Fail(l.t, "unexpected listener error", err.Error())
}

func (l *mockRemoteJudgeListener) OnComplete() {

}
