package elasticache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
)

// PublishContext publishes a JSON message using the caller's deadline.
// Its encoding matches Publish.
func PublishContext[T any](ctx context.Context, s *Service, channel string, message T) error {
	data, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("marshal pub/sub message: %w", err)
	}
	conn, err := s.redisPool.GetContext(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	_, err = redis.DoContext(conn, ctx, "PUBLISH", channel, data)
	return err
}

// SubscriptionOptions controls reconnects and lifecycle notifications.
// Callbacks run synchronously and should not block. OnSubscribed is also called
// after reconnection, allowing applications to reconcile missed messages.
type SubscriptionOptions struct {
	OnSubscribed  func()
	OnError       func(error)
	RetryInterval time.Duration
}

// SubscribeContext blocks until cancellation and reconnects after connection
// failures. It uses dedicated connections, separate from the command pool.
// Messages use the same JSON encoding as Publish and PublishContext.
// Pub/Sub is not durable; applications requiring recovery should reconcile on
// subscription and periodically. Use a primary endpoint for notifications.
func SubscribeContext[T any](ctx context.Context, s *Service, channel string, handler func(T), options SubscriptionOptions) error {
	return subscribeWithRetry(ctx, func() error {
		conn, err := s.dial(ctx)
		if err != nil {
			return err
		}
		return receiveSubscription(ctx, conn, channel, func(data []byte) error {
			var message T
			if err := json.Unmarshal(data, &message); err != nil {
				return err
			}
			handler(message)
			return nil
		}, options)
	}, options)
}

func receiveSubscription(ctx context.Context, conn redis.Conn, channel string, handler func([]byte) error, options SubscriptionOptions) error {
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	if err := psc.Subscribe(channel); err != nil {
		return err
	}
	for {
		switch message := psc.ReceiveContext(ctx).(type) {
		case redis.Subscription:
			if message.Kind == "subscribe" && message.Channel == channel && options.OnSubscribed != nil {
				options.OnSubscribed()
			}
		case redis.Message:
			if message.Channel == channel {
				if err := handler(message.Data); err != nil && options.OnError != nil {
					options.OnError(fmt.Errorf("decode pub/sub message: %w", err))
				}
			}
		case error:
			return message
		}
	}
}

func subscribeWithRetry(ctx context.Context, subscribe func() error, options SubscriptionOptions) error {
	interval := options.RetryInterval
	if interval <= 0 {
		interval = time.Second
	}
	for ctx.Err() == nil {
		err := subscribe()
		if ctx.Err() != nil {
			break
		}
		if err == nil {
			err = fmt.Errorf("pub/sub connection ended")
		}
		if options.OnError != nil {
			options.OnError(err)
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
	return ctx.Err()
}
