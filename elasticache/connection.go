package elasticache

import (
	"context"

	"github.com/gomodule/redigo/redis"
)

// WithConnectionContext runs a bounded pipeline or subscription on a dedicated
// connection using this Service's existing endpoint and authentication. It never
// exposes or closes the shared command pool. The callback must not retain conn.
// Cancellation interrupts Send/Flush as well as reads; the connection is not
// returned to the pool with potentially unacknowledged commands.
func (s *Service) WithConnectionContext(ctx context.Context, use func(redis.Conn) error) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	conn, err := s.dial(ctx)
	if err != nil {
		return err
	}
	defer conn.Close()
	closed := make(chan struct{})
	stop := context.AfterFunc(ctx, func() {
		defer close(closed)
		conn.Close()
	})
	defer func() {
		if !stop() {
			<-closed
		}
	}()
	err = use(conn)
	if ctx.Err() != nil {
		return ctx.Err()
	}
	return err
}
