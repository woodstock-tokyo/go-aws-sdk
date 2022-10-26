package secretsmanager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetSecretValue test get secret value
func TestGetSecretValue(t *testing.T) {
	svc := NewService()
	svc.SetRegion("ap-northeast-1")
	resp := svc.GetSecretValue("woodstock-api-local")
	assert.NoError(t, resp.Error)
}
