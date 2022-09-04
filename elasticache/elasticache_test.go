package elasticache

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var svc *Service

func init() {
	svc = NewService("redis-11725.c294.ap-northeast-1-2.ec2.cloud.redislabs.com:11725", DialPassword(os.Getenv("WS_ELASTICACHE_PASSWORD")))
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
