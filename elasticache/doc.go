// Package elasticache wraps ElastiCache related services.
//
// PublishContext and SubscribeContext provide JSON Pub/Sub with caller-controlled
// cancellation. SubscribeContext blocks, uses a dedicated connection, and
// reconnects automatically. Run it in a caller-owned goroutine. Its
// SubscriptionOptions.OnSubscribed callback runs after every subscription
// acknowledgement, including reconnects, so callers can reload their state.
// Pub/Sub messages are not durable: callers should also reconcile periodically.
//
// Configure connection/write timeouts with DialConnectTimeout and
// DialWriteTimeout. Use a primary endpoint for update notifications rather than
// relying on delivery through a replica/reader endpoint. Existing Publish and
// Subscribe remain available with their original behavior.
package elasticache
