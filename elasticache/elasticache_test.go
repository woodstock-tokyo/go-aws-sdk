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
	err := svc.Ping()
	assert.Nil(t, err, "ping should not return error")
}

// TestSet test redis set
func TestSet(t *testing.T) {
	err := svc.Set("test", []byte("abc"))
	assert.Nil(t, err, "set should not return error")
}

func TestSetTwice(t *testing.T) {
	err := svc.Set("test", []byte("abcde"))
	assert.Nil(t, err, "set should not return error")
}

// TestGet test redis set
func TestGet(t *testing.T) {
	value, err := svc.Get("test")
	assert.Nil(t, err, "get should not return error")
	assert.Equal(t, value, []byte("abcde"), "get should return expected value")
}

// TestDelete test redis delete
func TestDelete(t *testing.T) {
	err := svc.Delete("test")
	assert.Nil(t, err, "delete should not return error")
}

// TestSAdd test redis SADD
func TestSAdd(t *testing.T) {
	people := []Person{
		{Name: "John", Age: 30},
		{Name: "Jane", Age: 25},
	}

	err := SAdd(svc, "test", people, 5*60)
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
	assert.Equal(t, actual, expected, "SMembers should return expected value")
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
