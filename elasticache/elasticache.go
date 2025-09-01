package elasticache

import (
	"context"
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
func Set[T any](s *Service, key string, value T, ttlSeconds uint, nx ...bool) error {
	_nx := false
	if len(nx) == 1 {
		_nx = nx[0]
	}

	conn := s.redisPool.Get()
	defer conn.Close()

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	if !_nx {
		_, err = conn.Do("SET", key, jsonBytes)
	} else {
		_, err = conn.Do("SET", key, jsonBytes, "NX")
	}

	if ttlSeconds > 0 {
		_, err = conn.Do("EXPIRE", key, ttlSeconds)
	}
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

// MGet gets multiple keys and unmarshals JSON values into the provided type
func MGet[T any](s *Service, keys []string) (map[string]T, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	// Convert keys slice to interface slice for the Do method
	interfaceKeys := make([]interface{}, len(keys))
	for i, key := range keys {
		interfaceKeys[i] = key
	}

	// Perform MGET command
	values, err := redis.ByteSlices(conn.Do("MGET", interfaceKeys...))
	if err != nil {
		return nil, fmt.Errorf("error retrieving values for keys: %v", err)
	}

	// Create a map to hold key-value pairs
	result := make(map[string]T)
	for i, key := range keys {
		if values[i] != nil {
			var data T
			if err := json.Unmarshal(values[i], &data); err != nil {
				return nil, fmt.Errorf("error unmarshalling value for key %s: %v", key, err)
			}
			result[key] = data
		}
	}

	return result, nil
}

// GetKeys get keys by pattern
func GetKeys(s *Service, pattern string) ([]string, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	iter := 0
	count := 10000
	keys := []string{}
	for {
		arr, err := redis.Values(conn.Do("SCAN", iter, "MATCH", pattern, "COUNT", count))
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

func ZIncrBy[T any](s *Service, key string, increment float64, member T, ttlSeconds uint) (float64, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	// Marshal the member to JSON string
	jsonBytes, err := json.Marshal(member)
	if err != nil {
		return 0, fmt.Errorf("failed to marshal member for ZINCRBY: %w", err)
	}
	memberStr := string(jsonBytes)

	newScore, err := redis.Float64(conn.Do("ZINCRBY", key, increment, memberStr))
	if err != nil {
		return 0, fmt.Errorf("ZINCRBY failed: %w", err)
	}

	if ttlSeconds > 0 {
		_, _ = conn.Do("EXPIRE", key, ttlSeconds)
	}

	return newScore, nil
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

// SIsMember checks if a given member exists in the Redis set
func SIsMember[T any](s *Service, key string, member T) (bool, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	// Serialize the member to JSON
	jsonBytes, err := json.Marshal(member)
	if err != nil {
		return false, fmt.Errorf("failed to marshal member for SISMEMBER: %w", err)
	}

	exists, err := redis.Bool(conn.Do("SISMEMBER", key, jsonBytes))
	if err != nil {
		return false, fmt.Errorf("SISMEMBER command failed: %w", err)
	}

	return exists, nil
}

// SCard returns the number of elements in the set
// return -1 if the set does not exist
func SCard(s *Service, key string) (int, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	count, err := redis.Int(conn.Do("SCARD", key))
	if err != nil {
		return -1, fmt.Errorf("SCARD command failed: %w", err)
	}

	return count, nil
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
func ZRangeWithScore[T comparable, U any](s *Service, key string, start, end int) (members []T, scores []U, err error) {
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
func ZRevRangeWithScore[T comparable, U any](s *Service, key string, start, end int) (members []T, scores []U, err error) {
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

// ZRangeByScoreWithScore zrangebyscore with score
func ZRangeByScoreWithScore[T comparable, U any](s *Service, key string, min, max int64) (members []T, scores []U, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	args := redis.Args{}.Add(key).Add(min).Add(max).Add("WITHSCORES")

	strs, err := redis.Strings(conn.Do("ZRANGEBYSCORE", args...))
	if err != nil {
		return []T{}, []U{}, err
	}

	total := len(strs) / 2
	for i := range total {
		var t T
		if err = json.Unmarshal([]byte(strs[2*i]), &t); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}
		var u U
		if err = json.Unmarshal([]byte(strs[2*i+1]), &u); err != nil {
			fmt.Println("Failed to unmarshal score:", err)
			continue
		}

		members = append(members, t)
		scores = append(scores, u)
	}

	return
}

// ZRevRangeByScoreWithScore zrevrangebyscore with score
func ZRevRangeByScoreWithScore[T comparable, U any](s *Service, key string, max, min int64) (members []T, scores []U, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	args := redis.Args{}.Add(key).Add(max).Add(min).Add("WITHSCORES")

	strs, err := redis.Strings(conn.Do("ZREVRANGEBYSCORE", args...))
	if err != nil {
		return []T{}, []U{}, err
	}

	total := len(strs) / 2
	for i := range total {
		var t T
		if err = json.Unmarshal([]byte(strs[2*i]), &t); err != nil {
			fmt.Println("Failed to unmarshal member:", err)
			continue
		}
		var u U
		if err = json.Unmarshal([]byte(strs[2*i+1]), &u); err != nil {
			fmt.Println("Failed to unmarshal score:", err)
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

// ZCard zcard
func ZCard(s *Service, key string) (count int, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	count, err = redis.Int(conn.Do("ZCARD", key))
	if err != nil {
		return -1, err
	}

	return count, nil
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

// ZRemRangeByScore removes members in a sorted set within the given score range.
func ZRemRangeByScore(s *Service, key string, min, max float64) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("ZREMRANGEBYSCORE", key, min, max)
	if err != nil {
		return fmt.Errorf("ZREMRANGEBYSCORE command failed: %w", err)
	}

	return nil
}

// HSet sets a field in a hash, overwriting if it already exists.
func HSet[T any](s *Service, key, field string, value T) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	// Serialize the value to JSON
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value for HSET: %w", err)
	}

	_, err = conn.Do("HSET", key, field, jsonBytes)
	if err != nil {
		return fmt.Errorf("failed to set field in hash: %w", err)
	}

	return nil
}

// HGet retrieves a field value from a Redis hash.
func HGet[T any](s *Service, key, field string) (data T, err error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	// Get the value from Redis
	value, err := redis.Bytes(conn.Do("HGET", key, field))
	if err == redis.ErrNil {
		return data, fmt.Errorf("field %s not found in hash %s", field, key) // Handle missing field case
	} else if err != nil {
		return data, fmt.Errorf("failed to retrieve field from hash: %w", err)
	}

	// Deserialize JSON into struct
	err = json.Unmarshal(value, &data)
	if err != nil {
		return data, fmt.Errorf("failed to unmarshal HGET value: %w", err)
	}

	return data, nil
}

// HGetAll retrieves all key-value pairs from a Redis hash.
func HGetAll[T any](s *Service, key string) (map[string]T, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	// Retrieve all fields and values from the hash
	data, err := redis.StringMap(conn.Do("HGETALL", key))
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve hash: %w", err)
	}

	// Convert values to the desired struct type
	result := make(map[string]T)
	for field, jsonValue := range data {
		var value T
		err := json.Unmarshal([]byte(jsonValue), &value)
		if err != nil {
			return nil, fmt.Errorf("failed to unmarshal field %s: %w", field, err)
		}
		result[field] = value
	}

	return result, nil
}

// LPush pushes a value to the left of a Redis list.
func LPush[T any](s *Service, key string, value T) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	jsonBytes, err := json.Marshal(value)
	if err != nil {
		return err
	}

	_, err = conn.Do("LPUSH", key, jsonBytes)
	return err
}

// LRange retrieves a range of values from a Redis list and unmarshals them into the provided type.
func LRange[T any](s *Service, key string, start, stop int) ([]T, error) {
	conn := s.redisPool.Get()
	defer conn.Close()

	values, err := redis.ByteSlices(conn.Do("LRANGE", key, start, stop))
	if err != nil {
		return nil, err
	}

	var result []T
	for _, v := range values {
		var item T
		if err := json.Unmarshal(v, &item); err == nil {
			result = append(result, item)
		}
	}
	return result, nil
}

// Copy copy
func Copy(s *Service, fromKey, toKey string) (err error) {
	args := redis.Args{}.Add(fromKey).Add(toKey)
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err = conn.Do("COPY", args...)
	if err != nil {
		return err
	}
	return
}

// Rename renames a key in
func Rename(s *Service, oldKey, newKey string) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	_, err := conn.Do("RENAME", oldKey, newKey)
	if err != nil {
		return fmt.Errorf("failed to rename key from %s to %s: %w", oldKey, newKey, err)
	}
	return nil
}

// Publish sends a message to a Redis Pub/Sub channel
func Publish[T any](s *Service, channel string, message T) error {
	conn := s.redisPool.Get()
	defer conn.Close()

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	_, err = conn.Do("PUBLISH", channel, msgBytes)
	if err != nil {
		return fmt.Errorf("failed to publish message to channel %s: %w", channel, err)
	}

	return nil
}

// Subscribe listens for messages on a Redis Pub/Sub channel and processes them with a callback
func Subscribe[T any](s *Service, channel string, ctx context.Context, handler func(T)) error {
	conn := s.redisPool.Get()
	psc := redis.PubSubConn{Conn: conn}

	// Subscribe to the channel
	if err := psc.Subscribe(channel); err != nil {
		conn.Close()
		return fmt.Errorf("failed to subscribe to channel %s: %w", channel, err)
	}

	// Keep listening for messages
	// Listen for messages in a separate goroutine
	pubsubChan := make(chan redis.Message, 100)
	// Subscriber
	go func() {
		defer conn.Close()
		defer psc.Unsubscribe(channel)
		for {
			select {
			case <-ctx.Done():
				return
			case v, ok := <-pubsubChan:
				if !ok {
					return // Channel closed, and release the conn
				}
				var data T
				if err := json.Unmarshal(v.Data, &data); err != nil {
					fmt.Printf("Failed to unmarshal message: %v\n", err)
					continue
				}
				handler(data)
			}
		}
	}()

	// Publisher
	go func() {
		defer close(pubsubChan)
		for {
			_v := psc.Receive()
			switch v := _v.(type) {
			case redis.Message:
				pubsubChan <- v
			case redis.Subscription:
				fmt.Printf("Subscription message to %s (kind: %s, count: %d)\n", v.Channel, v.Kind, v.Count)
				if v.Kind == "unsubscribe" || v.Kind == "punsubscribe" {
					return
				}
			case error:
				fmt.Printf("Redis Pub/Sub error: %v\n", v)
				return
			}
		}
	}()

	return nil
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
