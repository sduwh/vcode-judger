package providers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProviderPOJ_Login(t *testing.T) {
	p, err := NewProviderPOJ()
	assert.NoError(t, err)

	err = p.Login()
	assert.NoError(t, err)
}

func TestProviderPOJ_HasLogin(t *testing.T) {

}

func TestProviderPOJ_Submit(t *testing.T) {

}

func TestProviderPOJ_Status(t *testing.T) {

}

func TestProviderPOJ_fetchCompileError(t *testing.T) {

}
