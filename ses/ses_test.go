package ses

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	. "github.com/woodstock-tokyo/woodstock-utils"
)

func TestCreateTemplate(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	opts := &CreateTemplateOptions{
		TemplateName: "test_template",
		Subject:      "test",
		HTML:         Vtop("<html><body><h1>Hello {{name}}</h1></body></html>"),
	}

	resp := svc.CreateTemplate(opts)
	assert.NoError(t, resp.Error)
}

func TestUpdateTemplate(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	opts := &UpdateTemplateOptions{
		TemplateName: "test_template",
		Subject:      "test",
		HTML:         Vtop("<html><body><h1>Hello {{name}} - New Template</h1></body></html>"),
	}

	resp := svc.UpdateTemplate(opts)
	assert.NoError(t, resp.Error)
}

// TestSendMessage test send message
func TestSendMail(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	opts := &SendEmailOptions{
		Sender:     "contact@woodstock.club",
		Recipients: []string{"min@woodstock.club"},
		CCs:        []string{"brian@woodstock.club"},
		BCCs:       []string{"daisuke@woodstock.club"},
		Template:   "test_template",
		TemplateData: map[string]string{
			"name": "Dear Valued Customer",
		},
		ConfigurationSet: Vtop("managed-dedicated-ip"),
	}

	resp := svc.SendEmail(opts)
	assert.NoError(t, resp.Error)
}

func TestListTemplates(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	resp := svc.ListTemplates(new(ListTemplatesOptions))
	assert.NoError(t, resp.Error)
	assert.Equal(t, 1, len(resp.Templates))
}

func TestGetTemplate(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	resp := svc.GetTemplate(&GetTemplateOptions{
		TemplateName: "test_template",
	})

	assert.NoError(t, resp.Error)
	assert.Equal(t, *resp.TemplateName, "test_template")
	assert.Equal(t, *resp.SubjectPart, "test")
}

func TestDeleteTemplate(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	resp := svc.DeleteTemplate(&DeleteTemplateOptions{
		TemplateName: "test_template",
	})

	assert.NoError(t, resp.Error)
}

func TestListTemplatesAgain(t *testing.T) {
	svc := NewService(os.Getenv("WS_SES_AWS_ACCESS_KEY_ID"), os.Getenv("WS_SES_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")

	resp := svc.ListTemplates(new(ListTemplatesOptions))
	assert.NoError(t, resp.Error)
	assert.Equal(t, 0, len(resp.Templates))
}
