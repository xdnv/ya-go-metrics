package app

import (
	"time"

	"github.com/sethvargo/go-retry"
)

func NewBackoff(maxRetries uint64) retry.Backoff {
	//init backoff
	backoff := retry.NewExponential(1 * time.Second)
	backoff = CustomExponential(1*time.Second, backoff)
	if maxRetries > 0 {
		backoff = retry.WithMaxRetries(maxRetries, backoff)
	}
	return backoff
}

// Custom backoff middleware making intervals 1,3,5,5,5,..
func CustomExponential(t time.Duration, next retry.Backoff) retry.BackoffFunc {
	return func() (time.Duration, bool) {
		val, stop := next.Next()
		if stop {
			return 0, true
		}

		switch val {
		case 1 * time.Second:
			val = 1 * time.Second
		case 2 * time.Second:
			val = 3 * time.Second
		default:
			val = 5 * time.Second
		}

		return val, false
	}
}

// func DoRetry(ctx context.Context, retry.Backoff, func) error {

// 	err := retry.Do(ctx, backoff, func(ctx context.Context) error {
// 		// Actual retry logic here
// 		return nil
// 	})

// 	if err != nil {
// 		// handle error
// 	}
// 	return err
// }
