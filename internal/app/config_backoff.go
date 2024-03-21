package app

import (
	"context"
	"errors"
	"fmt"
	"internal/adapters/logger"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
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

func DoRetry(ctx context.Context, max uint64, f func(ctx context.Context) error) error {

	backoff := NewBackoff(max)
	if err := retry.Do(ctx, backoff, f); err != nil {
		return err
	}
	return nil
}

func HandleRetriableWeb(err error, retryMessage string) error {
	if err != nil {
		logger.Error(fmt.Sprintf("%s, retry: %v", retryMessage, err))
		return retry.RetryableError(err)
	}
	return nil
}

func HandleRetriableDB(err error, retryMessage string) error {
	if err != nil {

		var pgErr *pgconn.PgError
		//if errors.As(err, &pgErr) && pgerrcode.IsInvalidCatalogName(pgErr.Code) { ////debug line with wrong database name error subclass
		if errors.As(err, &pgErr) && pgerrcode.IsConnectionException(pgErr.Code) {
			logger.Error(fmt.Sprintf("%s, retry: %v", retryMessage, err))
			return retry.RetryableError(err)
		} else {
			logger.Error(fmt.Sprintf("%s, FATAL: %v", retryMessage, err))
			return err
		}

	}
	return nil
}
