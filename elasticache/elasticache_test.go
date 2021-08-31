package elasticache

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

var svc *Service

func init() {
	// localhost
	svc = NewService(":6379")
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

// TestGet test redis set
func TestGet(t *testing.T) {
	value, err := svc.Get("test")
	assert.Nil(t, err, "get should not return error")
	assert.Equal(t, value, []byte("abc"), "get should return expected value")
}

// TestDelete test redis delete
func TestDelete(t *testing.T) {
	err := svc.Delete("test")
	assert.Nil(t, err, "delete should not return error")
}
