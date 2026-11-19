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
	// Optional transport liveness checks; zero preserves the existing behavior.
	PingInterval            time.Duration
	ReadTimeout             time.Duration
	DisconnectOnDecodeError bool
}

// SubscribeContext blocks until cancellation and reconnects after connection
// failures. It uses dedicated connections, separate from the command pool.
// Messages use the same JSON encoding as Publish and PublishContext.
// Pub/Sub is not durable; applications requiring recovery should reconcile on
// subscription and periodically. Use a primary endpoint for notifications.
func SubscribeContext[T any](ctx context.Context, s *Service, channel string, handler func(T), options SubscriptionOptions) error {
	return subscribeWithRetry(ctx, func() error {
		return SubscribeOnceContext(ctx, s, channel, handler, options)
	}, options)
}

// SubscribeOnceContext confirms SUBSCRIBE before OnSubscribed, then receives
// until cancellation or transport failure. It does not reconnect; callers can
// invalidate their state before opening another subscription.
func SubscribeOnceContext[T any](ctx context.Context, s *Service, channel string, handler func(T), options SubscriptionOptions) error {
	if options.PingInterval < 0 || options.ReadTimeout < 0 ||
		(options.PingInterval > 0 && options.ReadTimeout > 0 && options.ReadTimeout <= options.PingInterval) {
		return fmt.Errorf("invalid subscription heartbeat intervals")
	}
	return s.WithConnectionContext(ctx, func(conn redis.Conn) error {
		return receiveSubscription(ctx, conn, channel, func(data []byte) error {
			var message T
			if err := json.Unmarshal(data, &message); err != nil {
				return err
			}
			handler(message)
			return nil
		}, options)
	})
}

func receiveSubscription(ctx context.Context, conn redis.Conn, channel string, handler func([]byte) error, options SubscriptionOptions) error {
	defer conn.Close()
	psc := redis.PubSubConn{Conn: conn}
	if err := psc.Subscribe(channel); err != nil {
		return err
	}
	pingCtx, stopPing := context.WithCancel(ctx)
	pingDone := make(chan struct{})
	go func() {
		defer close(pingDone)
		if options.PingInterval == 0 {
			return
		}
		ticker := time.NewTicker(options.PingInterval)
		defer ticker.Stop()
		for {
			select {
			case <-pingCtx.Done():
				return
			case <-ticker.C:
				if err := psc.Ping("subscription-heartbeat"); err != nil {
					conn.Close()
					return
				}
			}
		}
	}()
	defer func() { stopPing(); conn.Close(); <-pingDone }()
	for {
		var reply interface{}
		if options.ReadTimeout > 0 {
			reply = psc.ReceiveWithTimeout(options.ReadTimeout)
		} else {
			reply = psc.ReceiveContext(ctx)
		}
		switch message := reply.(type) {
		case redis.Subscription:
			if message.Kind == "subscribe" && message.Channel == channel && options.OnSubscribed != nil {
				options.OnSubscribed()
			}
		case redis.Message:
			if message.Channel == channel {
				if err := handler(message.Data); err != nil {
					err = fmt.Errorf("decode pub/sub message: %w", err)
					if options.DisconnectOnDecodeError {
						return err
					}
					if options.OnError != nil {
						options.OnError(err)
					}
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
