package elasticache

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"

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
