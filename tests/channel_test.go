package tests

import (
	"fmt"
	"testing"

	"github.com/sduwh/vcode-judger/channel"
	"github.com/stretchr/testify/assert"
)

func TestRedisChannel(t *testing.T) {
	ch, err := channel.NewRedisChannel("127.0.0.1:6379")
	assert.NoError(t, err)
	defer func() {
		_ = ch.Close()
	}()

	err = ch.Push("topic-1", []byte("aaa"))
	assert.NoError(t, err)

	v, err := ch.Take("topic-1", 0)
	assert.NoError(t, err)

	fmt.Println(string(v))
}
