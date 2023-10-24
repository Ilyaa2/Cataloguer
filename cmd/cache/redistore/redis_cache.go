package redistore

import (
	"Cataloguer/cmd/cache"
	"github.com/gomodule/redigo/redis"
	"log"
)

type RedisCache struct {
	conn              redis.Conn
	sessionRepository *SessionRepository
}

func New(url string) (*RedisCache, error) {
	conn, err := redis.DialURL(url)
	if err != nil {
		log.Fatal(err)
		return nil, err
	}
	return &RedisCache{conn: conn}, nil
}

func (r *RedisCache) Session() cache.SessionRepository {
	if r.sessionRepository == nil {
		r.sessionRepository = &SessionRepository{redisCache: r}
	}
	return r.sessionRepository
}
