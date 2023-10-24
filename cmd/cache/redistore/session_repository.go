package redistore

import (
	"Cataloguer/cmd/custom_errors"
	"errors"
	"github.com/gomodule/redigo/redis"
)

const (
	NoExpiration = 0
)

type SessionRepository struct {
	redisCache *RedisCache
}

func (s *SessionRepository) GetValue(key string) (string, error) {
	data, err := s.redisCache.conn.Do("GET", key)
	item, err := redis.String(data, err)
	if err == redis.ErrNil {
		return "", errors.New(custom_errors.RecordNotFoundInCache)
	} else if err != nil {
		return "", err
	}
	return item, nil
}

func (s *SessionRepository) SetValue(key string, value string, ex int) error {
	var reply interface{}
	var err error
	if NoExpiration == ex {
		reply, err = s.redisCache.conn.Do("SET", key, value)
	} else {
		reply, err = s.redisCache.conn.Do("SET", key, value, "EX", ex)
	}
	_, err = redis.String(reply, err)
	//if result != "OK"
	if err != nil {
		return errors.New(custom_errors.CantSetValueInCache)
	}
	return nil
}

func (s *SessionRepository) DeleteValue(key string) {
	_, _ = s.redisCache.conn.Do("DEL", key)
}
