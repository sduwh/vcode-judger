package channel

import (
	"sync/atomic"
	"time"

	"github.com/go-redis/redis/v7"
)

type RedisChannel struct {
	client *redis.Client
	stop   int32
}

func NewRedisChannel(addr string) (Channel, error) {
	client := redis.NewClient(&redis.Options{
		Addr: addr,
	})
	return &RedisChannel{client: client}, nil
}

func (c *RedisChannel) Push(topic string, message []byte) error {
	return c.client.LPush(topic, message).Err()
}

func (c *RedisChannel) Take(topic string, timeout time.Duration) ([]byte, error) {
	r, err := c.client.BRPop(timeout, topic).Result()
	if err != nil {
		return nil, err
	}
	return []byte(r[1]), nil
}

func (c *RedisChannel) Listen(topic string, listener Listener) {
	go func() {
		for atomic.LoadInt32(&c.stop) == 0 {
			v, err := c.Take(topic, 0)
			if err != nil {
				listener.OnError(err)
				continue
			}
			listener.OnNext(v)
		}
		listener.OnComplete()
	}()
}

func (c *RedisChannel) Close() error {
	atomic.StoreInt32(&c.stop, 1)
	return c.client.Close()
}
