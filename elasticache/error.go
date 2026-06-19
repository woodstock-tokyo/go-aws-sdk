package elasticache

import (
	"fmt"

	"github.com/gomodule/redigo/redis"
)

// FieldNotFoundError wraps redis.ErrNil with the specific key/field that was missing.
// Sentinel check — works because Unwrap returns ErrNil
// Use below to check for this condition.
// ```
// if errors.Is(err, redis.ErrNil) { ... }
// ```
// Typed check — works regardless of Unwrap
// Use below to check for this condition.
// ```
// var nfErr *elasticache.FieldNotFoundError
// if errors.As(err, &nfErr) { ... }
// ```

type FieldNotFoundError struct {
	Key   string
	Field string
}

func (e *FieldNotFoundError) Error() string {
	return fmt.Sprintf("field %s not found in hash %s", e.Field, e.Key)
}

func (e *FieldNotFoundError) Unwrap() error { return redis.ErrNil }
