package elasticache

import (
	"encoding/json"
	"fmt"
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

// ///////////////////////////////// We use functions instead of methods because of the generic type /////////////////////////////////
// Ping ping
func Ping(s *Service) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := redis.String(conn.Do("PING"))
	if err != nil {
		return fmt.Errorf("cannot 'PING' db: %v", err)
	}
	return nil
}

// Get get
func Get[T any](s *Service, key string) (data T, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	value, err := redis.Bytes(conn.Do("GET", key))
	if err != nil {
		return
	}

	err = json.Unmarshal(value, &data)
	return
}

// Set set
func Set[T any](s *Service, key string, value T, ttlSeconds uint) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = conn.Do("SETEX", key, ttlSeconds, jsonBytes)
	return err
}

// Delete delete
func Delete(s *Service, key string) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("DEL", key)
	return err
}

// Exists exist
func Exists(s *Service, key string) (bool, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	ok, err := redis.Bool(conn.Do("EXISTS", key))
	if err != nil {
		return ok, fmt.Errorf("error checking if key %s exists: %v", key, err)
	}
	return ok, err
}

// GetKeys get keys by pattern
func GetKeys(s *Service, pattern string) ([]string, error) {
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
func Incr(s *Service, counterKey string) (int, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	return redis.Int(conn.Do("INCR", counterKey))
}

// SAdd sadd
func SAdd[T any](s *Service, key string, members []T, ttlSeconds uint) (err error) {
	// convert structs to strings (JSON)
	var memberStrings []string
	for _, member := range members {
		jsonBytes, marshalErr := json.Marshal(member)
		if marshalErr != nil {
			err = marshalErr
			return
		}
		memberStrings = append(memberStrings, string(jsonBytes))
	}

	args := redis.Args{}.Add(key).AddFlat(memberStrings)
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("SADD", args...)
	if err != nil {
		return err
	}

	if ttlSeconds > 0 {
		_, err = conn.Do("EXPIRE", key, ttlSeconds)
	}
	return
}

// SMembers smembers
func SMembers[T any](s *Service, key string) (members []T, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	members = []T{}
	memberStrings, err := redis.Strings(conn.Do("SMEMBERS", key))
	if err != nil {
		return []T{}, err
	}

	for _, member := range memberStrings {
		var t T
		if err = json.Unmarshal([]byte(member), &t); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}
		members = append(members, t)
	}

	return
}

// SRem srem
// have to make it as a function instead of a method because of the generic type
func SRem[T any](s *Service, key string, membersToRemove []T) (err error) {
	// convert structs to strings (JSON)
	var memberStrings []string
	for _, member := range membersToRemove {
		jsonBytes, marshalErr := json.Marshal(member)
		if marshalErr != nil {
			err = marshalErr
			return
		}
		memberStrings = append(memberStrings, string(jsonBytes))
	}

	args := redis.Args{}.Add(key).AddFlat(memberStrings)

	conn := s.redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("SREM", args...)
	return
}

// ZAdd zadd
func ZAdd[T any, U comparable](s *Service, key string, members []T, scores []U, ttlSeconds uint) (err error) {
	// convert structs to strings (JSON)
	args := redis.Args{}.Add(key)
	for i, member := range members {
		jsonBytes, err := json.Marshal(member)
		if err != nil {
			return err
		}
		args = args.AddFlat(map[U]string{scores[i]: string(jsonBytes)})
	}

	conn := s.redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ZADD", args...)
	if err != nil {
		return err
	}

	if ttlSeconds > 0 {
		_, err = conn.Do("EXPIRE", key, ttlSeconds)
	}
	return
}

// ZRem zrem
func ZRem[T any](s *Service, key string, membersToRemove []T) (err error) {
	// convert structs to strings (JSON)
	var memberStrings []string
	for _, member := range membersToRemove {
		jsonBytes, marshalErr := json.Marshal(member)
		if marshalErr != nil {
			err = marshalErr
			return
		}
		memberStrings = append(memberStrings, string(jsonBytes))
	}

	args := redis.Args{}.Add(key).AddFlat(memberStrings)

	conn := s.redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("ZREM", args...)
	return
}

// ZRangeWithScore zrange with score
func ZRangeWithScore[T comparable, U any](s *Service, key string, start, end uint) (members []T, scores []U, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	args := redis.Args{}.Add(key).Add(start).Add(end).Add("WITHSCORES")

	strs, err := redis.Strings(conn.Do("ZRANGE", args...))
	if err != nil {
		return []T{}, []U{}, err
	}

	total := len(strs) / 2
	for i := 0; i < total; i++ {
		var t T
		if err = json.Unmarshal([]byte(strs[2*i]), &t); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}
		var u U
		if err = json.Unmarshal([]byte(strs[2*i+1]), &u); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}

		members = append(members, t)
		scores = append(scores, u)
	}

	return
}

// ZRevRangeWithScore zrevrange with score
func ZRevRangeWithScore[T comparable, U any](s *Service, key string, start, end uint) (members []T, scores []U, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	args := redis.Args{}.Add(key).Add(start).Add(end).Add("WITHSCORES")

	strs, err := redis.Strings(conn.Do("ZREVRANGE", args...))
	if err != nil {
		return []T{}, []U{}, err
	}

	total := len(strs) / 2
	for i := 0; i < total; i++ {
		var t T
		if err = json.Unmarshal([]byte(strs[2*i]), &t); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}
		var u U
		if err = json.Unmarshal([]byte(strs[2*i+1]), &u); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}

		members = append(members, t)
		scores = append(scores, u)
	}

	return
}

// ZScore zscore
func ZScore[T any, U any](s *Service, key string, member T) (score *U, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	jsonBytes, err := json.Marshal(member)
	if err != nil {
		return
	}
	memberString := string(jsonBytes)

	args := redis.Args{}.Add(key).AddFlat(memberString)

	bytes, err := redis.Bytes(conn.Do("ZSCORE", args...))
	if err != nil {
		return
	}

	if err = json.Unmarshal(bytes, &score); err != nil {
		return nil, err
	}

	return score, err
}

// ZRank zrank with score
func ZRank[T any](s *Service, key string, member T) (rank int, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	jsonBytes, err := json.Marshal(member)
	if err != nil {
		return
	}
	memberString := string(jsonBytes)

	args := redis.Args{}.Add(key).AddFlat(memberString)

	rank, err = redis.Int(conn.Do("ZRANK", args...))
	if err != nil {
		return
	}

	return rank, err
}

// ZCount zcount
func ZCount[U comparable](s *Service, key string, min *U, max *U) (count int, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	args := redis.Args{}.Add(key)

	if min != nil {
		args = args.Add(*min)
	} else {
		args = args.Add("-inf")
	}

	if max != nil {
		args = args.Add(*max)
	} else {
		args = args.Add("+inf")
	}

	count, err = redis.Int(conn.Do("ZCOUNT", args...))
	if err != nil {
		return
	}

	return
}

// /////////////////////////////// PRIVATE ///////////////////////////////////////
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
