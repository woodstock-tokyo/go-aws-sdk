package elasticache

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
)

func TestConnectionContextCancellationInterruptsFlush(t *testing.T) {
	client, server := net.Pipe()
	defer server.Close()
	service := &Service{dial: func(context.Context) (redis.Conn, error) {
		return redis.NewConn(client, 0, 0), nil
	}}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	done := make(chan error, 1)
	go func() {
		done <- service.WithConnectionContext(ctx, func(conn redis.Conn) error {
			if err := conn.Send("GET", "blocked"); err != nil {
				return err
			}
			return conn.Flush()
		})
	}()
	select {
	case err := <-done:
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatalf("expected deadline, got %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation did not interrupt pipeline")
	}
}

func TestSubscribeOnceHeartbeatAndStrictDecoding(t *testing.T) {
	server := miniredis.RunT(t)
	service := NewService(server.Addr())
	defer service.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready, delivered := make(chan struct{}, 1), make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- SubscribeOnceContext(ctx, service, "prices", func(value string) { delivered <- value }, SubscriptionOptions{
			OnSubscribed: func() { ready <- struct{}{} },
			PingInterval: 5 * time.Millisecond, ReadTimeout: 100 * time.Millisecond,
			DisconnectOnDecodeError: true,
		})
	}()
	waitPubSub(t, ready)
	// Several read-timeout periods without market messages must stay healthy.
	time.Sleep(250 * time.Millisecond)
	if err := PublishContext(ctx, service, "prices", "quiet-but-connected"); err != nil {
		t.Fatal(err)
	}
	if got := waitPubSub(t, delivered); got != "quiet-but-connected" {
		t.Fatal(got)
	}
	server.Publish("prices", "malformed-json")
	if err := waitPubSub(t, done); err == nil {
		t.Fatal("strict subscription accepted invalid JSON")
	}
	if err := Ping(service); err != nil {
		t.Fatal("dedicated subscription closed shared command pool:", err)
	}
}
