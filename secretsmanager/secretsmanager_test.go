package secretsmanager

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestGetSecretValue test get secret value
func TestGetSecretValue(t *testing.T) {
	svc := NewService(os.Getenv("WS_SECRETS_MANAGER_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SECRETS_MANAGER_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	resp := svc.GetSecretValue("woodstock-api-local")
	assert.NoError(t, resp.Error)
}
