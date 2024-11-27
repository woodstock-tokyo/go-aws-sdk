# GO AWS SDK

AWS client designed for easy use.

## Install

```bash
go get -u github.com/woodstock-tokyo/go-aws-sdk
```

## S3

example:

```Go
import "github.com/woodstock-tokyo/go-aws-sdk/s3"

func main() {
    svc := NewService("<<access key id>>", "<<secret key>>")
    svc.SetRegion("ap-northeast-1")
    svc.SetBucket("my bucket")
    opts := &UploadOptions{
        FileName: "./test.png",
        Public:   true,
    }
    resp := svc.AsyncUpload(opts)
    if <-resp.Error != nil {
        t.Error(resp.Error)
    }
}
```

## Dynamodb

example:

```Go
package dynamo

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/woodstock-tokyo/go-aws-sdk/dynamo"
)

// Use struct tags much like the standard JSON library,
// you can embed anonymous structs too!
type widget struct {
	UserID int       // Hash key, a.k.a. partition key
	Time   time.Time // Range key, a.k.a. sort key

	Msg       string              `dynamo:"Message"`    // Change name in the database
	Count     int                 `dynamo:",omitempty"` // Omits if zero value
	Children  []widget            // Lists
	Friends   []string            `dynamo:",set"` // Sets
	Set       map[string]struct{} `dynamo:",set"` // Map sets, too!
	SecretKey string              `dynamo:"-"`    // Ignored
}

func main() {
  svc := dynamo.NewService(config.S3.AccessKeyID, config.S3.SecretAccessKey)
	svc.SetRegion(config.S3.Region)
  db := svc.Instance()

  table := db.Table("Widgets")

	// put item
	w := widget{UserID: 613, Time: time.Now(), Msg: "hello"}
	err := table.Put(w).Run()

	// get the same item
	var result widget
	err = table.Get("UserID", w.UserID).
		Range("Time", dynamo.Equal, w.Time).
		One(&result)

	// get all items
	var results []widget
	err = table.Scan().All(&results)

	// use placeholders in filter expressions (see Expressions section below)
	var filtered []widget
	err = table.Scan().Filter("'Count' > ?", 10).All(&filtered)
}
```

## Async

Use _AsyncSend_, _AsyncReceive_, _AsyncDelete_ to handle SQS operation request concurrently

### Tobe added: SES, Agcod, CloudSearch ...
