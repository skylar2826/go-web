package session_store

import (
	"awesomeProject2/web/context"
	"awesomeProject2/web/custom_error"
	"awesomeProject2/web/session"
	"errors"
	"fmt"
	"github.com/redis/go-redis/v9"
	"time"
)

type redisSession struct {
	cmd      redis.Cmdable
	id       string
	redisKey string
}

func (r *redisSession) Get(c *context.Context, key string) (interface{}, error) {
	return r.cmd.HGet(c.R.Context(), r.redisKey, key).Result()
}

func (r *redisSession) Set(c *context.Context, key string, val interface{}) error {
	const lua = `
	if redis.call("exists", KEYS[1])
	then
		return redis.call("hset", KEYS[1], ARGV[1],ARGV[2])
	else
		return -1
	end
	`

	res, err := r.cmd.Eval(c.R.Context(), lua, []string{r.redisKey}, key, val).Int()
	if err != nil {
		return err
	}
	if res < 0 {
		return fmt.Errorf("session: session 创建失败")
	}
	return nil
}

func (r *redisSession) ID() string {
	return r.id
}

type redisStore struct {
	prefix  string
	cmd     redis.Cmdable
	expired time.Duration
}

func (r *redisStore) Generate(c *context.Context, id string) (session.Session, error) {
	key := getRedisKey(r.prefix, id)
	_, err := r.cmd.HSet(c.R.Context(), key, id, id).Result()
	if err != nil {
		return nil, err
	}
	_, err = r.cmd.Expire(c.R.Context(), key, r.expired).Result()
	if err != nil {
		return nil, err
	}
	return &redisSession{
		cmd:      r.cmd,
		id:       id,
		redisKey: key,
	}, nil
}

func (r *redisStore) Get(c *context.Context, id string) (session.Session, error) {
	key := getRedisKey(r.prefix, id)
	cnt, err := r.cmd.Exists(c.R.Context(), key).Result()
	if err != nil {
		return nil, err
	}
	if cnt != 1 {
		return nil, custom_error.ErrorNotFound(fmt.Sprintf("session %s\n", id))
	}
	return &redisSession{
		id:       id,
		cmd:      r.cmd,
		redisKey: key,
	}, nil
}

func (r *redisStore) Remove(c *context.Context, id string) error {
	key := getRedisKey(r.prefix, id)
	_, err := r.cmd.Del(c.R.Context(), key).Result()
	return err
}

func (r *redisStore) Refresh(c *context.Context, id string) error {
	key := getRedisKey(r.prefix, id)
	ok, err := r.cmd.Expire(c.R.Context(), key, r.expired).Result()
	if !ok {
		return errors.New(fmt.Sprintf("更新过期时间失败，session key 不存在: %s\n", id))
	}
	return err
}

func NewRedisStore(cmd redis.Cmdable, prefix string, expired time.Duration) session.Store {
	return &redisStore{
		cmd:     cmd,
		prefix:  prefix,
		expired: expired,
	}
}

func getRedisKey(prefix string, id string) string {
	return fmt.Sprintf("%s-%s", prefix, id)
}
