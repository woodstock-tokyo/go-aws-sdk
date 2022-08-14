package cloudsearch

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

type Document struct {
	PostID       string `json:"post_id"`
	ContentEN    string `json:"content_en"`
	ContentJA    string `json:"content_ja"`
	UserID       string `json:"user_id"`
	ParentPostID string `json:"parent_post_id"`
	TopicID      string `json:"topic_id"`
	NewsID       string `json:"news_id"`
	PickID       string `json:"pick_id"`
	PortfolioID  string `json:"portfolio_id"`
}

func TestUploadAdd(t *testing.T) {
	svc := initService()
	opts := &CloudSearchDocumentUploadOptions{
		Content: []CloudSearchDocumentUploadContent{
			{
				ID:   1,
				Type: CloudSearchDocumentUploadTypeAdd,
				Fields: Document{
					PostID:       "1",
					ContentEN:    "welcome to woodstock! this is the genesis message of the timeline",
					ContentJA:    "woodstockへようこそ",
					UserID:       "1",
					ParentPostID: "",
					TopicID:      "",
					NewsID:       "",
					PickID:       "",
					PortfolioID:  "",
				},
			},
			{
				ID:   2,
				Type: CloudSearchDocumentUploadTypeAdd,
				Fields: Document{
					PostID:       "2",
					ContentEN:    "welcome to hotel california",
					ContentJA:    "hotel californiaへようこそ",
					UserID:       "2",
					ParentPostID: "",
					TopicID:      "",
					NewsID:       "",
					PickID:       "",
					PortfolioID:  "",
				},
			},
		},
	}

	resp := svc.Upload(opts)
	assert.NoError(t, resp.Error)
	assert.Equal(t, resp.Adds, int64(2))
}

// TestSearch test cloud search
func TestSearch(t *testing.T) {
	svc := initService()
	opts := &CloudSearchOptions{
		Query: "welcome",
		Highlight: []CloudSearchHighlightOption{
			{
				Field:  "content_ja",
				Format: CloudSearchHighlightFormatHTML,
			},
			{
				Field:  "content_en",
				Format: CloudSearchHighlightFormatHTML,
			},
		},
	}

	resp := svc.Search(opts)
	assert.EqualValues(t, 2, resp.Found)
	assert.NoError(t, resp.Error)
}

func TestUploadDelete(t *testing.T) {
	svc := initService()
	opts := &CloudSearchDocumentUploadOptions{
		Content: []CloudSearchDocumentUploadContent{
			{
				ID:     1,
				Type:   CloudSearchDocumentUploadTypeDelete,
				Fields: Document{},
			},
			{
				ID:     2,
				Type:   CloudSearchDocumentUploadTypeDelete,
				Fields: Document{},
			},
		},
	}

	resp := svc.Upload(opts)
	assert.NoError(t, resp.Error)
	assert.EqualValues(t, resp.Deletes, 2)
}

// TestSearch test cloud search
func TestSearchAgain(t *testing.T) {
	svc := initService()
	opts := &CloudSearchOptions{
		Query: "welcome",
	}

	resp := svc.Search(opts)
	assert.EqualValues(t, 0, resp.Found)
	assert.NoError(t, resp.Error)
}

func initService() *Service {
	svc := NewService(os.Getenv("WS_CLOUDSEARCH_AWS_ACCESS_KEY_ID"), os.Getenv("WS_CLOUDSEARCH_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	svc.SetSearchEndpoint("search-woodstock-stg-4xctkbk7zdvtmh35j7hsaqaiwi.ap-northeast-1.cloudsearch.amazonaws.com")
	svc.SetDocumentEndpoint("doc-woodstock-stg-4xctkbk7zdvtmh35j7hsaqaiwi.ap-northeast-1.cloudsearch.amazonaws.com")
	return svc
}
