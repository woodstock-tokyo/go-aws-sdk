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

## Secrets Manager (AWS SDK v2)

Use `github.com/woodstock-tokyo/go-aws-sdk/secretsmanager/v2` for new
context-aware integrations. It supports the standard AWS credential/region
chains, including IAM roles and temporary session tokens; static access keys
are not required.

```go
import (
    "context"

    "github.com/woodstock-tokyo/go-aws-sdk/secretsmanager/v2"
)

func loadSecret(ctx context.Context, secretID string) (string, error) {
    client, err := secretsmanager.New(ctx)
    if err != nil {
        return "", err
    }
    return client.SecretString(ctx, secretID)
}
```

Create and reuse a client rather than constructing one per call. `New` accepts
the SDK's config load options; `NewFromConfig` accepts explicit configuration
and service options, and `NewWithAPI` enables testing with an injected client.
Each client has independent region/credential configuration, not a global
singleton. `SecretString` returns the raw secret without assuming a JSON schema,
rejects empty/binary-only values and forwards request cancellation/deadlines.
It does not log secret values. Errors wrap SDK causes, so callers should redact
provider details before logging or returning them to untrusted clients.

The legacy `secretsmanager.NewService` API is unchanged. Tests for the new
package use mocks and a loopback HTTP server, not live AWS credentials:

```sh
go test -race ./secretsmanager/v2
```

### Tobe added: SES, Agcod, CloudSearch ...
