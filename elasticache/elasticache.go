package elasticache

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Service service includes context and credentials
type Service struct {
	redisPool *redis.Pool
}

// NewService service initializer
func NewService(host string) *Service {
	pool := newRedisPool(host)
	return &Service{
		redisPool: pool,
	}
}

// Close close pool
func (s *Service) Close() {
	s.redisPool.Close()
}

// Ping ping
func (s *Service) Ping() error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		return fmt.Errorf("cannot 'PING' db: %v", err)
	}
	return nil
}

// Get get
func (s *Service) Get(key string) ([]byte, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	var data []byte
	data, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return data, fmt.Errorf("error getting key %s: %v", key, err)
	}
	return data, err
}

// Set set
func (s *Service) Set(key string, value []byte) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}
		return fmt.Errorf("error setting key %s to %s: %v", key, v, err)
	}
	return err
}

// Exists exist
func (s *Service) Exists(key string) (bool, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}
	return ok, err
}

// Delete delete
func (s *Service) Delete(key string) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}

// GetKeys get keys by pattern
func (s *Service) GetKeys(pattern string) ([]string, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	iter := 0
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern))
		if err != nil {
			return keys, fmt.Errorf("error retrieving '%s' keys", pattern)
		}

		iter, _ = redis.Int(arr[0], nil)
		k, _ := redis.Strings(arr[1], nil)
		keys = append(keys, k...)

		if iter == 0 {
			break
		}
	}

	return keys, nil
}

// Incr incr
func (s *Service) Incr(counterKey string) (int, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	return redis.Int(conn.Do("INCR", counterKey))
}

func newRedisPool(host string) *redis.Pool {
	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", host)
			if err != nil {
				return nil, err
			}
			return c, err
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
