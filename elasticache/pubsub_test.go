package elasticache

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
)

func waitPubSub[T any](t *testing.T, ch <-chan T) T {
	t.Helper()
	select {
	case value := <-ch:
		return value
	case <-time.After(5 * time.Second):
		t.Fatal("timed out waiting for pub/sub event")
		var zero T
		return zero
	}
}

func TestContextPubSubDeliveryAndCancellation(t *testing.T) {
	mr := miniredis.RunT(t)
	mr.RequireAuth("test-password")
	s := NewService(mr.Addr(), DialPassword("test-password"), DialConnectTimeout(time.Second), DialWriteTimeout(time.Second))
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{}, 1)
	messages := make(chan string, 2)
	done := make(chan error, 1)
	go func() {
		done <- SubscribeContext(ctx, s, "updates", func(value string) { messages <- value }, SubscriptionOptions{
			OnSubscribed: func() { ready <- struct{}{} },
		})
	}()
	waitPubSub(t, ready)
	if err := PublishContext(ctx, s, "updates", "new-value"); err != nil {
		t.Fatal(err)
	}
	if got := waitPubSub(t, messages); got != "new-value" {
		t.Fatalf("message = %q, want new-value", got)
	}
	// Existing Publish uses the same wire format as the new context API.
	if err := Publish(s, "updates", "legacy-value"); err != nil {
		t.Fatal(err)
	}
	if got := waitPubSub(t, messages); got != "legacy-value" {
		t.Fatalf("legacy message = %q", got)
	}
	cancel()
	if err := waitPubSub(t, done); !errors.Is(err, context.Canceled) {
		t.Fatalf("subscription returned %v, want context cancellation", err)
	}
	// Cancelling a dedicated subscriber must not close the command pool.
	if err := Ping(s); err != nil {
		t.Fatal(err)
	}
}

func TestContextPubSubReconnectsAndReconciles(t *testing.T) {
	mr := miniredis.RunT(t)
	s := NewService(mr.Addr(), DialConnectTimeout(time.Second))
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{}, 10)
	reported := make(chan error, 100)
	messages := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- SubscribeContext(ctx, s, "updates", func(value string) { messages <- value }, SubscriptionOptions{
			OnSubscribed:  func() { ready <- struct{}{} },
			OnError:       func(err error) { reported <- err },
			RetryInterval: 10 * time.Millisecond,
		})
	}()
	waitPubSub(t, ready)
	mr.Close()
	waitPubSub(t, reported)
	if err := mr.Restart(); err != nil {
		t.Fatal(err)
	}
	waitPubSub(t, ready)
	if err := PublishContext(ctx, s, "updates", "after-reconnect"); err != nil {
		t.Fatal(err)
	}
	if got := waitPubSub(t, messages); got != "after-reconnect" {
		t.Fatalf("message after reconnect = %q", got)
	}
	cancel()
	waitPubSub(t, done)
}

func TestContextPubSubMalformedMessageDoesNotDisconnect(t *testing.T) {
	mr := miniredis.RunT(t)
	s := NewService(mr.Addr())
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{}, 1)
	reported := make(chan error, 1)
	messages := make(chan string, 1)
	done := make(chan error, 1)
	go func() {
		done <- SubscribeContext(ctx, s, "updates", func(value string) { messages <- value }, SubscriptionOptions{
			OnSubscribed: func() { ready <- struct{}{} },
			OnError:      func(err error) { reported <- err },
		})
	}()
	waitPubSub(t, ready)
	mr.Publish("updates", "not JSON")
	waitPubSub(t, reported)
	if err := PublishContext(ctx, s, "updates", "valid"); err != nil {
		t.Fatal(err)
	}
	if got := waitPubSub(t, messages); got != "valid" {
		t.Fatalf("message = %q, want valid", got)
	}
	cancel()
	waitPubSub(t, done)
}

func TestContextPubSubAlreadyCancelled(t *testing.T) {
	s := NewService("127.0.0.1:1")
	defer s.Close()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := PublishContext(ctx, s, "updates", "unused"); !errors.Is(err, context.Canceled) {
		t.Fatalf("publish returned %v, want cancellation", err)
	}
	if err := SubscribeContext(ctx, s, "updates", func(string) {}, SubscriptionOptions{}); !errors.Is(err, context.Canceled) {
		t.Fatalf("subscribe returned %v, want cancellation", err)
	}
}

func TestSubscribeContextCancellationDuringReconnectWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	reported := make(chan error, 1)
	done := make(chan error, 1)
	go func() {
		done <- subscribeWithRetry(ctx, func() error { return errors.New("offline") }, SubscriptionOptions{
			OnError:       func(err error) { reported <- err },
			RetryInterval: time.Hour,
		})
	}()
	waitPubSub(t, reported)
	cancel()
	if err := waitPubSub(t, done); !errors.Is(err, context.Canceled) {
		t.Fatalf("returned %v, want cancellation", err)
	}
}
