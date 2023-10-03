package elasticache

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gomodule/redigo/redis"
)

// Service service includes context and credentials
type Service struct {
	redisPool *redis.Pool
}

// dialOptions internal dial options
type dialOptions struct {
	username string
	password string
}

// DialOption specifies an option for dialing a Redis server.
type DialOption struct {
	f func(*dialOptions)
}

// DialUserName specifies the username to use when connecting to elastcache
func DialUserName(username string) DialOption {
	return DialOption{func(do *dialOptions) {
		do.username = username
	}}
}

// DialPassword specifies the password to use when connecting to elastcache
func DialPassword(password string) DialOption {
	return DialOption{func(do *dialOptions) {
		do.password = password
	}}
}

// NewService service initializer
func NewService(host string, options ...DialOption) *Service {
	pool := newRedisPool(host, options...)
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

// MGet get
func (s *Service) MGet(ctx context.Context, keys ...string) (map[string][]byte, error) {
	conn, err := s.redisPool.GetContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("error getting connection: %w", err)
	}
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	args := make([]any, len(keys))
	for i, key := range keys {
		args[i] = key
	}

	values, err := redis.Values(conn.Do("MGET", args...))
	if err != nil {
		return nil, fmt.Errorf("error getting keys %s: %w", strings.Join(keys, " "), err)
	}

	data := make(map[string][]byte, len(keys))
	if len(keys) != len(values) {
		return nil, errors.New("key and value length mismatch")
	}

	for i := 0; i < len(keys); i++ {
		if valueBites, ok := values[i].([]byte); ok {
			data[keys[i]] = valueBites
		}
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

// SetContext set with a context
func (s *Service) SetContext(ctx context.Context, key string, value []byte) error {
	conn, err := s.redisPool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection: %w", err)
	}
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	_, err = conn.Do("SET", key, value)
	if err != nil {
		return fmt.Errorf("error setting key %s : %w", key, err)
	}

	return err
}

// SetExpiry set with expiry
func (s *Service) SetExpiry(key string, value []byte, expireSecond uint) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("SET", key, value, "EX", expireSecond)
	if err != nil {
		v := string(value)
		if len(v) > 15 {
			v = v[0:12] + "..."
		}

		return fmt.Errorf("error setting key with expire %s to %s: %v", key, v, err)
	}

	return err
}

// SetExpiryContext set with expiry and context
func (s *Service) SetExpiryContext(ctx context.Context, key string, value []byte, expireSecond uint) error {
	conn, err := s.redisPool.GetContext(ctx)
	if err != nil {
		return fmt.Errorf("error getting connection: %w", err)
	}
	defer func(conn redis.Conn) {
		_ = conn.Close()
	}(conn)

	_, err = conn.Do("SET", key, value, "EX", expireSecond)
	if err != nil {
		return fmt.Errorf("error setting key with expire %s : %w", key, err)
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

func newRedisPool(host string, options ...DialOption) *redis.Pool {
	do := dialOptions{}
	for _, option := range options {
		option.f(&do)
	}

	return &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			dailOptions := []redis.DialOption{}
			if do.username != "" {
				dailOptions = append(dailOptions, redis.DialUsername(do.username))
			}
			if do.password != "" {
				dailOptions = append(dailOptions, redis.DialPassword(do.password))
			}

			return redis.Dial("tcp", host, dailOptions...)
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			return err
		},
	}
}
