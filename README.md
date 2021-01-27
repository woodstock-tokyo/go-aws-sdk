# GO AWS SDK

AWS client designed for easy use.

## Install

```bash
go get -u github.com/woodstock-tokyo/go-aws-sdk
```

## S3

example:

```Go
import github.com/woodstock-tokyo/go-aws-sdk

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

## SQS

### SQS Send Message

```Go
import bitbucket.org/indiesquare/indiesquare-aws/sqs

func main() {
    svc := NewService("<<access key id>>", "<<secret key>>")
    svc.SetRegion("ap-northeast-1")
    svc.SetQueueURL("<<queue url>>")

    message := &TestMessage{
        Foo: "test",
        Bar: Bar{
            Name: "Oda Nobunaga",
            Age:  34,
            Time: time.Now(),
        },
    }

    encodedMessage, _ := json.Marshal(message)
    opts := &SendOptions{
        Message:        encodedMessage,
        MessageGroupID: "Unit-Testing",
    }

    resp := svc.Send(opts)
    if resp.Error != nil {
        t.Error(resp.Error)
    }
}
```

### SQS Receive Message

```Go
import bitbucket.org/indiesquare/indiesquare-aws/sqs

func main() {
    svc := NewService("<<access key id>>", "<<secret key>>")
    svc.SetRegion("ap-northeast-1")
    svc.SetQueueURL("<<queue url>>")

    svc := NewService("<<access key id>>", "<<secret access key>>")
    svc.SetRegion("ap-northeast-1")
    svc.SetQueueURL(Q)

    opts := new(ReceiveOptions)

    resp := svc.Receive(opts)
    if resp.Error != nil {
        t.Error(resp.Error)
    } else {
        receiptHandle = resp.ReceiptHandle
    }
}
```

### SQS Delete Message

```Go
import bitbucket.org/indiesquare/indiesquare-aws/sqs

func main() {
    svc := NewService("<<access key id>>", "<<secret key>>")
    svc.SetRegion("ap-northeast-1")
    svc.SetQueueURL("<<queue url>>")

    opts := &DeleteOptions{
        ReceiptHandle: receiptHandle,
    }

    resp := svc.Delete(opts)
    if resp.Error != nil {
        t.Error(resp.Error)
    }
}
```

## Async

Use _AsyncSend_, _AsyncReceive_, _AsyncDelete_ to handle SQS operation request concurrently
