package providers

import (
	"testing"

	"github.com/sduwh/vcode-judger/consts"
	"github.com/sduwh/vcode-judger/models"
	"github.com/stretchr/testify/assert"
)

func TestProviderPOJ_Login(t *testing.T) {
	p, err := NewProviderPOJ()
	assert.NoError(t, err)

	err = p.Login()
	assert.NoError(t, err)
}

func TestProviderPOJ_HasLogin(t *testing.T) {
	p, err := NewProviderPOJ()
	assert.NoError(t, err)

	ok, err := p.HasLogin()
	assert.NoError(t, err)
	assert.False(t, ok)

	err = p.Login()
	assert.NoError(t, err)

	ok, err = p.HasLogin()
	assert.NoError(t, err)
	assert.True(t, ok)
}

func TestProviderPOJ_Submit(t *testing.T) {
	p, err := NewProviderPOJ()
	assert.NoError(t, err)

	err = p.Login()
	assert.NoError(t, err)

	task := &models.RemoteJudgeTask{
		ID:         "a",
		ProviderID: "1000",
		Language:   consts.LanguageC,
		Code: `// Code here
			int main() {
				return 0;
			}`,
	}

	id, err := p.Submit(task)
	assert.NoError(t, err)

	status, err := p.Status(task, id)
	assert.NoError(t, err)
	assert.NotEmpty(t, status.Status)
	assert.Equal(t, "a", status.TaskID)
	assert.Equal(t, id, status.SubmitID)
	assert.Empty(t, status.TimeUsed)
	assert.Empty(t, status.MemoryUsed)
	assert.Empty(t, status.CompileError)
}
