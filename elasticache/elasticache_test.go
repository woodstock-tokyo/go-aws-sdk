package elasticache

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var svc *Service

type Person struct {
	Name string
	Age  int
}

func init() {
	svc = NewService("localhost:6379")
}

// TestPing test redis ping
func TestPing(t *testing.T) {
	err := Ping(svc)
	assert.Nil(t, err, "ping should not return error")
}

// TestSet test redis set
func TestSet(t *testing.T) {
	err := Set(svc, "test", "abc", 0)
	assert.Nil(t, err, "set should not return error")
}

func TestSetTwice(t *testing.T) {
	err := Set(svc, "test", "abcde", 30)
	assert.Nil(t, err, "set should not return error")
}

func TestGetKeys(t *testing.T) {
	keys, err := GetKeys(svc, "test")
	assert.Nil(t, err, "get kets should not return error")
	assert.Equal(t, keys[0], "test")
}

// TestGet test redis set
func TestGet(t *testing.T) {
	value, err := Get[string](svc, "test")
	assert.Nil(t, err, "get should not return error")
	assert.Equal(t, value, "abcde", "get should return expected value")
}

func TestSetAgain(t *testing.T) {
	err := Set(svc, "test1", "12345", 30)
	assert.Nil(t, err, "set should not return error")
}

// TestGet test redis set
func TestMGet(t *testing.T) {
	value, err := MGet[string](svc, []string{"test", "test1"})
	assert.Equal(t, len(value), 2, "get should return expected value")
	assert.Nil(t, err, "get should not return error")
}

// TestDelete test redis delete
func TestDelete(t *testing.T) {
	err := Delete(svc, "test")
	assert.Nil(t, err, "delete should not return error")
	err = Delete(svc, "tes1")
	assert.Nil(t, err, "delete should not return error")
}

// TestGet test redis set
func TestGetAgain(t *testing.T) {
	value, err := Get[string](svc, "test")
	assert.Error(t, err, "get should return error")
	assert.Equal(t, value, "", "get should return empty")
}

// TestSAdd test redis SADD
func TestSAdd(t *testing.T) {
	people := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
	}

	err := SAdd(svc, "test", people, 30)
	assert.Nil(t, err, "SAdd should not return error")
}

// TestSMembers test redis SMembers
func TestSMembers(t *testing.T) {
	// SADD FIFO so we need to revert the order here
	expected := []Person{
		{Name: "Jane", Age: 25},
		{Name: "John", Age: 30},
	}

	actual, err := SMembers[Person](svc, "test")
	assert.Nil(t, err, "SMembers should not return error")
	assert.Equal(t, len(actual), len(expected), "SMembers should return expected value")
}

// TestSIsMember tests SIsMember for set existence
func TestSIsMember(t *testing.T) {
	key := "test_sismember"
	Delete(svc, key)

	person := Person{Name: "TestUser", Age: 99}

	// Should return false before adding
	exists, err := SIsMember(svc, key, person)
	assert.Nil(t, err, "SIsMember should not return error before adding")
	assert.False(t, exists, "SIsMember should return false for non-existing member")

	// Add the person
	err = SAdd(svc, key, []Person{person}, 30)
	assert.Nil(t, err, "SAdd should not return error")

	// Should return true after adding
	exists, err = SIsMember(svc, key, person)
	assert.Nil(t, err, "SIsMember should not return error after adding")
	assert.True(t, exists, "SIsMember should return true for existing member")

	// Cleanup
	err = Delete(svc, key)
	assert.Nil(t, err, "Delete should not return error")
}

// TestSRem test redis SREM
func TestSRem(t *testing.T) {
	people := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
	}

	err := SRem(svc, "test", people)
	assert.Nil(t, err, "SRem should not return error")
}

// TestSMembersAgain test redis SMembers
func TestSMembersAgain(t *testing.T) {
	actual, err := SMembers[Person](svc, "test")
	assert.Nil(t, err, "SMembers should not return error")
	assert.Equal(t, len(actual), 0, "SMembers should return expected value")
}

// TestZAdd test redis ZADD
func TestZAdd(t *testing.T) {
	scores := []float64{1.12, 4.5, 3.5, 4.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}

	err := ZAdd(svc, "test", members, scores, 30)
	assert.Nil(t, err, "ZAdd should not return error")
}

// TestZRem test redis ZREM
func TestZRem(t *testing.T) {
	scores := []float64{1.12, 3.5, 3.5, 4.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	err := ZRem(svc, "test", members[0:2])
	assert.Nil(t, err, "ZRem should not return error")
}

// TestZCount test redis ZCOUNT
func TestZCount(t *testing.T) {
	scores := []float64{1.12, 3.5, 3.5, 4.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	var min float64 = 1.12
	var max float64 = 3.5

	count, err := ZCount(svc, "test", &min, &max)
	assert.Nil(t, err, "ZCount should not return error")
	assert.Equal(t, 3, count, "ZCount should return expected count")
}

// TestZRank test redis ZRANK
func TestZRank(t *testing.T) {
	Delete(svc, "test")
	scores := []float64{1.12, 3.5, 2.0, 3.5, 4.5, 2.0}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jana", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jona", Age: 20},
		{Name: "Jono", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	rank, err := ZRank(svc, "test", Person{Name: "Jene", Age: 15})
	assert.Nil(t, err, "ZRank should not return error")
	assert.Equal(t, 2, rank, "ZRank should return expected rank")
}

// TestZScore test redis ZSCORE
func TestZScore(t *testing.T) {
	Delete(svc, "test")
	scores := []float64{1.12, 3.5, 4.5, 5.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	score, err := ZScore[Person, float64](svc, "test", Person{Name: "Jene", Age: 15})
	assert.Nil(t, err, "ZScore should not return error")
	assert.Equal(t, 5.5, *score, "ZScore should return expected score")
}

// TestZRangeWithScore test redis ZRANGE
func TestZRangeWithScore(t *testing.T) {
	Delete(svc, "test")
	scores := []float64{1.12, 3.5, 4.5, 5.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	_members, _scores, err := ZRangeWithScore[Person, float64](svc, "test", 0, 10)
	assert.Nil(t, err, "ZRangeWithScore should not return error")
	assert.Equal(t, 4, len(_members), "ZRangeWithScore should return expected length")

	for i, _member := range _members {
		assert.Equal(t, members[i], _member, "ZRangeWithScore should return expected member")
		assert.Equal(t, scores[i], _scores[i], "ZRangeWithScore should return expected score")
		i++
	}
}

// TestZRangeWithScore test redis ZRANGE
func TestZRangeByScoreWithScore(t *testing.T) {
	Delete(svc, "test")
	scores := []float64{1.12, 3.5, 4.5, 5.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	_members, _scores, err := ZRangeByScoreWithScore[Person, float64](svc, "test", 0, 10)
	assert.Nil(t, err, "ZRangeByScoreWithScore should not return error")
	assert.Equal(t, 4, len(_members), "ZRangeByScoreWithScore should return expected length")

	for i, _member := range _members {
		assert.Equal(t, members[i], _member, "ZRangeByScoreWithScore should return expected member")
		assert.Equal(t, scores[i], _scores[i], "ZRangeByScoreWithScore should return expected score")
		i++
	}
}

// TestZRevRangeWithScore test redis ZREVRANGE
func TestZRevRangeWithScore(t *testing.T) {
	scores := []float64{1.12, 3.5, 4.5, 5.5}
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}
	_ = ZAdd(svc, "test", members, scores, 30)
	// reverse scores and members
	for i, j := 0, len(members)-1; i < j; i, j = i+1, j-1 {
		scores[i], scores[j] = scores[j], scores[i]
		members[i], members[j] = members[j], members[i]
	}

	_members, _scores, err := ZRevRangeWithScore[Person, float64](svc, "test", 0, 2)
	assert.Nil(t, err, "ZRevRangeWithScore should not return error")
	assert.Equal(t, 3, len(_members), "ZRevRangeWithScore should return expected length")

	for i, _member := range _members {
		assert.Equal(t, members[i], _member, "ZRevRangeWithScore should return expected member")
		assert.Equal(t, scores[i], _scores[i], "ZRevRangeWithScore should return expected score")
		i++
	}
}

// TestCopy test redis COPY
func TestCopy(t *testing.T) {
	_ = Set(svc, "test", "abc", 30)
	err := Copy(svc, "test", "test-2")
	assert.Nil(t, err, "copy should not return error")

	value, _ := Get[string](svc, "test-2")
	assert.Equal(t, value, "abc", "get should return abc")
}

// TestGet test redis set
func TestRename(t *testing.T) {
	value, err := Get[string](svc, "test-2")
	assert.Nil(t, err, "get should not return error")
	fmt.Println(value)

	Rename(svc, "test-2", "test-3")
	value, err = Get[string](svc, "test-3")
	assert.Nil(t, err, "get should not return error")
	fmt.Println(value)

	Rename(svc, "test-3", "test-2")
	value, err = Get[string](svc, "test-2")
	assert.Nil(t, err, "get should not return error")
	fmt.Println(value)
}

func TestHSet(t *testing.T) {
	person := Person{Name: "Alice", Age: 28}

	err := HSet(svc, "test_hash", "Alice", person)
	assert.Nil(t, err, "HSet should not return an error")
}

func TestHGet(t *testing.T) {
	person, err := HGet[Person](svc, "test_hash", "Alice")
	assert.Nil(t, err, "HGet should not return an error")
	assert.Equal(t, "Alice", person.Name, "HGet should return the correct name")
	assert.Equal(t, 28, person.Age, "HGet should return the correct age")
}

func TestHGetAll(t *testing.T) {
	// Adding multiple entries
	_ = HSet(svc, "test_hash", "Bob", Person{Name: "Bob", Age: 35})
	_ = HSet(svc, "test_hash", "Charlie", Person{Name: "Charlie", Age: 40})

	// Fetch all fields
	data, err := HGetAll[Person](svc, "test_hash")
	assert.Nil(t, err, "HGetAll should not return an error")
	assert.Equal(t, 3, len(data), "HGetAll should return all stored items")
	assert.Equal(t, 28, data["Alice"].Age, "HGetAll should return correct age for Alice")
	assert.Equal(t, 35, data["Bob"].Age, "HGetAll should return correct age for Bob")
	assert.Equal(t, 40, data["Charlie"].Age, "HGetAll should return correct age for Charlie")
}

func TestHSetOverwrite(t *testing.T) {
	// Overwriting Alice's data
	newPerson := Person{Name: "Alice", Age: 30}
	err := HSet(svc, "test_hash", "Alice", newPerson)
	assert.Nil(t, err, "HSet should not return an error when overwriting")

	// Verify the new value
	person, err := HGet[Person](svc, "test_hash", "Alice")
	assert.Nil(t, err, "HGet should not return an error")
	assert.Equal(t, 30, person.Age, "HSet should overwrite the previous value")
}

// TestPublishSubscribe tests Redis Pub/Sub
func TestPublishSubscribe(t *testing.T) {
	channel := "test_pubsub"

	// Use a wait group to synchronize message receiving
	var wg sync.WaitGroup
	wg.Add(1)

	// Start subscriber
	go func() {
		err := Subscribe(svc, channel, func(msg string) {
			assert.Equal(t, "Hello, Redis!", msg, "Received message should match published message")
			wg.Done()
		})
		assert.Nil(t, err, "Subscribe should not return an error")
	}()

	// Wait briefly to ensure the subscriber is ready
	time.Sleep(500 * time.Millisecond)

	// Publish message
	err := Publish(svc, channel, "Hello, Redis!")
	assert.Nil(t, err, "Publish should not return an error")

	// Wait for the message to be received
	wg.Wait()
}

// TestSubscribeMultipleMessages tests Redis Pub/Sub with multiple messages
func TestSubscribeMultipleMessages(t *testing.T) {
	channel := "test_pubsub_multiple"
	var receivedMessages []string
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Expecting 3 messages
	wg.Add(3)

	// Start subscriber
	go func() {
		err := Subscribe(svc, channel, func(msg string) {
			mu.Lock()
			receivedMessages = append(receivedMessages, msg)
			mu.Unlock()
			wg.Done()
		})
		assert.Nil(t, err, "Subscribe should not return an error")
	}()

	time.Sleep(500 * time.Millisecond)

	// Publish multiple messages
	messages := []string{"Message 1", "Message 2", "Message 3"}
	for _, msg := range messages {
		err := Publish(svc, channel, msg)
		assert.Nil(t, err, "Publish should not return an error")
	}

	// Wait for all messages to be received
	wg.Wait()

	// Validate received messages
	mu.Lock()
	defer mu.Unlock()
	assert.ElementsMatch(t, messages, receivedMessages, "All messages should be received correctly")
}

// TestSubscribeWithStruct tests Redis Pub/Sub with a struct (e.g., Person)
func TestSubscribeWithStruct(t *testing.T) {
	channel := "test_pubsub_struct"
	expectedPerson := Person{Name: "Alice", Age: 28}
	var receivedPerson Person
	var wg sync.WaitGroup
	wg.Add(1)

	// Start subscriber
	go func() {
		err := Subscribe(svc, channel, func(person Person) {
			receivedPerson = person
			wg.Done()
		})
		assert.Nil(t, err, "Subscribe should not return an error")
	}()

	time.Sleep(500 * time.Millisecond)

	// Publish struct message
	err := Publish(svc, channel, expectedPerson)
	assert.Nil(t, err, "Publish should not return an error")

	// Wait for the message to be received
	wg.Wait()

	// Validate received struct
	assert.Equal(t, expectedPerson, receivedPerson, "Received struct should match the published struct")
}
