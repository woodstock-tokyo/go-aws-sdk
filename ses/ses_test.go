package ses

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/woodstock-tokyo/woodstock-utils"
)

// TestSendMessage test send message
func TestSendMail(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	opts := &SendEmailOptions{
		Sender:     "contact@woodstock.club",
		Recipients: []string{"min@woodstock.club"},
		CCs:        []string{"brian@woodstock.club"},
		BCCs:       []string{"daisuke@woodstock.club"},
		Template:   "test",
		TemplateData: map[string]string{
			"name": "Dear Valued Customer",
		},
		ConfigurationSet: Vtop("managed-dedicated-ip"),
	}

	resp := svc.SendTampleteEmail(opts)
	assert.NoError(t, resp.Error)
}
