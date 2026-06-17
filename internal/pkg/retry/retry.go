package retry

import (
	"context"

	"k8s.io/apimachinery/pkg/util/wait"
	kretry "k8s.io/client-go/util/retry"
)

// OnError retries fn on error, respecting context cancellation between attempts.
func OnError(ctx context.Context, backoff wait.Backoff, retriable func(error) bool, fn func() error) error {
	return kretry.OnError(backoff,
		func(err error) bool {
			if ctx.Err() != nil {
				return false
			}
			return retriable(err)
		},
		func() error {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fn()
		})
}
