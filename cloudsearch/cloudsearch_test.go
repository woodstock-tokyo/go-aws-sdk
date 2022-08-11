package cloudsearch

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSearch test cloud search
func TestSearch(t *testing.T) {
	svc := NewService(os.Getenv("WS_CLOUDSEARCH_AWS_ACCESS_KEY_ID"), os.Getenv("WS_CLOUDSEARCH_AWS_SECRET_ACCESS_KEY"))
	svc.SetRegion("ap-northeast-1")
	svc.SetEndpoint("search-woodstock-stg-4xctkbk7zdvtmh35j7hsaqaiwi.ap-northeast-1.cloudsearch.amazonaws.com")

	opts := &CloudSearchOptions{
		Query: "welcome",
	}

	resp := svc.Search(opts)
	assert.NoError(t, resp.Error)

	fmt.Println(resp)
}
