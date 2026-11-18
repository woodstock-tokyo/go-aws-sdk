package secretsmanager

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	awssecrets "github.com/aws/aws-sdk-go-v2/service/secretsmanager"
)

type fakeAPI func(context.Context, *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error)

func (f fakeAPI) GetSecretValue(ctx context.Context, input *awssecrets.GetSecretValueInput, _ ...func(*awssecrets.Options)) (*awssecrets.GetSecretValueOutput, error) {
	return f(ctx, input)
}

func TestClientSecretString(t *testing.T) {
	secret := `{"certificate":"AQID","unrelated":{"nested":true}}`
	ctx := context.WithValue(context.Background(), testKey{}, "forwarded")
	client := NewWithAPI(fakeAPI(func(actual context.Context, input *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error) {
		if actual != ctx || aws.ToString(input.SecretId) != "certificate-secret" {
			t.Fatal("context or secret ID not forwarded")
		}
		return &awssecrets.GetSecretValueOutput{SecretString: aws.String(secret)}, nil
	}))
	actual, err := client.SecretString(ctx, "certificate-secret")
	if err != nil || actual != secret {
		t.Fatalf("raw secret not preserved: %v", err)
	}
}

type testKey struct{}

func TestClientMissingValues(t *testing.T) {
	for _, output := range []*awssecrets.GetSecretValueOutput{
		nil, {}, {SecretString: aws.String("")}, {SecretString: aws.String(" \n\t")},
		{SecretBinary: []byte("binary-only")},
	} {
		client := NewWithAPI(fakeAPI(func(context.Context, *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error) {
			return output, nil
		}))
		if _, err := client.SecretString(context.Background(), "id"); !errors.Is(err, ErrEmptySecret) {
			t.Fatalf("expected empty-secret error: %v", err)
		}
	}
	var absent *Client
	for _, client := range []*Client{absent, NewWithAPI(nil)} {
		if _, err := client.SecretString(context.Background(), "id"); !errors.Is(err, ErrClientRequired) {
			t.Fatal(err)
		}
	}
	client := NewWithAPI(fakeAPI(func(context.Context, *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error) {
		t.Fatal("empty ID reached AWS")
		return nil, nil
	}))
	if _, err := client.SecretString(context.Background(), " "); !errors.Is(err, ErrSecretID) {
		t.Fatal(err)
	}
}

func TestClientCancellationAndErrors(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := New(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	client := NewWithAPI(fakeAPI(func(context.Context, *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error) {
		t.Fatal("canceled context reached AWS")
		return nil, nil
	}))
	if _, err := client.SecretString(ctx, "id"); !errors.Is(err, context.Canceled) {
		t.Fatal(err)
	}
	cause := errors.New("provider error")
	client = NewWithAPI(fakeAPI(func(context.Context, *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error) {
		return nil, cause
	}))
	if _, err := client.SecretString(context.Background(), "id"); !errors.Is(err, cause) {
		t.Fatal("provider cause not preserved")
	}
	ctx, stop := context.WithTimeout(context.Background(), time.Millisecond)
	defer stop()
	client = NewWithAPI(fakeAPI(func(ctx context.Context, _ *awssecrets.GetSecretValueInput) (*awssecrets.GetSecretValueOutput, error) {
		<-ctx.Done()
		return nil, ctx.Err()
	}))
	if _, err := client.SecretString(ctx, "id"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal(err)
	}
}

func TestClientRegionsAndSessionTokensAreIndependent(t *testing.T) {
	var mu sync.Mutex
	regions := make(map[string]int)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Amz-Security-Token") != "synthetic-session-token" {
			t.Error("session token not forwarded")
		}
		region := ""
		for _, candidate := range []string{"ap-northeast-1", "us-east-1"} {
			if strings.Contains(r.Header.Get("Authorization"), "/"+candidate+"/secretsmanager/aws4_request") {
				region = candidate
			}
		}
		if region == "" {
			t.Error("request not signed for expected region")
		}
		mu.Lock()
		regions[region]++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/x-amz-json-1.1")
		_, _ = w.Write([]byte(`{"SecretString":"{\"certificate\":\"AQID\"}"}`))
	}))
	defer server.Close()
	for _, region := range []string{"ap-northeast-1", "us-east-1"} {
		cfg := aws.Config{Region: region, Credentials: credentials.NewStaticCredentialsProvider("synthetic-access-key", "synthetic-secret-key", "synthetic-session-token")}
		client := NewFromConfig(cfg, func(options *awssecrets.Options) { options.BaseEndpoint = aws.String(server.URL) })
		if _, err := client.SecretString(context.Background(), "id"); err != nil {
			t.Fatal(err)
		}
	}
	if regions["ap-northeast-1"] != 1 || regions["us-east-1"] != 1 {
		t.Fatal("clients reused another instance's configuration")
	}
}

func TestClientUsesStandardEnvironmentCredentials(t *testing.T) {
	t.Setenv("AWS_ACCESS_KEY_ID", "synthetic-access-key")
	t.Setenv("AWS_SECRET_ACCESS_KEY", "synthetic-secret-key")
	t.Setenv("AWS_SESSION_TOKEN", "synthetic-session-token")
	t.Setenv("AWS_REGION", "ap-northeast-1")
	t.Setenv("AWS_EC2_METADATA_DISABLED", "true")
	t.Setenv("AWS_CONFIG_FILE", t.TempDir()+"/absent-config")
	t.Setenv("AWS_SHARED_CREDENTIALS_FILE", t.TempDir()+"/absent-credentials")
	t.Setenv("AWS_PROFILE", "")
	t.Setenv("AWS_DEFAULT_PROFILE", "")
	// Inspect the same default config chain used by New, without calling AWS.
	cfg, err := config.LoadDefaultConfig(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	value, err := cfg.Credentials.Retrieve(context.Background())
	if err != nil || value.SessionToken != "synthetic-session-token" || cfg.Region != "ap-northeast-1" {
		t.Fatal("default credential/region chain was not retained")
	}
	if _, err := New(context.Background()); err != nil {
		t.Fatal(err)
	}
}

func TestClientHTTPDeadline(t *testing.T) {
	release := make(chan struct{})
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {
		case <-r.Context().Done():
		case <-release:
		}
	}))
	defer server.Close()
	defer close(release)
	client := NewFromConfig(aws.Config{
		Region: "ap-northeast-1", HTTPClient: server.Client(),
		Credentials: credentials.NewStaticCredentialsProvider("synthetic-id", "synthetic-key", ""),
	}, func(options *awssecrets.Options) { options.BaseEndpoint = aws.String(server.URL) })
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	if _, err := client.SecretString(ctx, "synthetic-secret"); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("HTTP deadline not preserved: %v", err)
	}
}
