package elasticache

import (
	"testing"

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
	err := Set(svc, "test", "abc", 30)
	assert.Nil(t, err, "set should not return error")
}

func TestSetTwice(t *testing.T) {
	err := Set(svc, "test", "abcde", 30)
	assert.Nil(t, err, "set should not return error")
}

// TestGet test redis set
func TestGet(t *testing.T) {
	value, err := Get[string](svc, "test")
	assert.Nil(t, err, "get should not return error")
	assert.Equal(t, value, "abcde", "get should return expected value")
}

// TestDelete test redis delete
func TestDelete(t *testing.T) {
	err := Delete(svc, "test")
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

	count, err := ZCount[float64](svc, "test", &min, &max)
	assert.Nil(t, err, "ZCount should not return error")
	assert.Equal(t, 3, count, "ZCount should return expected count")
}

// TestZRem test redis ZREM
func TestZRemAgain(t *testing.T) {
	members := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
		{Name: "Jano", Age: 20},
		{Name: "Jene", Age: 15},
	}

	err := ZRem(svc, "test", members)
	assert.Nil(t, err, "ZRem should not return error")
}

// TestZRevRankWithScore test redis ZREVRANK
func TestZRankWithScore(t *testing.T) {
	scores := []float64{1.12, 3.5, 2.0, 3.5, 4.5, 2.0}
	members := []Person{ // rank => reverse rank
		{Name: "John", Age: 30}, // 0 => 6-0 => 6
		{Name: "Jana", Age: 25}, // 4 => 6-4 => 2
		{Name: "Jano", Age: 20}, // 2 => 6-2 => 4
		{Name: "Jona", Age: 20}, // 4 => 6-4 => 2
		{Name: "Jono", Age: 20}, // 5 => 6-5 => 1
		{Name: "Jene", Age: 15}, // 2 => 6-2 => 4
	}
	_ = ZAdd(svc, "test", members, scores, 30)

	rank, score, err := ZRankWithScore[Person, float64](svc, "test", Person{Name: "Jene", Age: 15})
	assert.Nil(t, err, "ZRankWithScore should not return error")
	assert.Equal(t, 2, rank, "ZRankWithScore should return expected rank")
	assert.Equal(t, 2.0, *score, "ZRankWithScore should return expected score")
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
