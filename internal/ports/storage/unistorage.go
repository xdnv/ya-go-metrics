// the (uni))storage module provides abstraction interface for both in-memory and DB metric storage
package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"internal/adapters/logger"
	"internal/app"
	"internal/domain"
)

// universal metric storage
type UniStorage struct {
	config  *app.ServerConfig
	ctx     context.Context
	stor    *MemStorage
	db      *DbStorage
	timeout time.Duration
}

// init metric storage
func NewUniStorage(cf *app.ServerConfig) *UniStorage {

	if cf.StorageMode == app.Database {
		var (
			conn *sql.DB
			err  error
		)

		conn, err = sql.Open("pgx", cf.DatabaseDSN)
		if err != nil {
			logger.Fatal(err.Error())
		}

		return &UniStorage{
			config:  cf,
			ctx:     context.Background(),
			db:      NewDbStorage(conn),
			timeout: 5 * time.Second,
		}
	} else {
		return &UniStorage{
			config: cf,
			ctx:    context.Background(),
			stor:   NewMemStorage(),
		}
	}
}

// prepare database (if applicable)
func (t UniStorage) Bootstrap() error {
	if t.config.StorageMode == app.Database {
		return t.db.Bootstrap(t.ctx)
	}

	return nil
}

// close DB connection or file
func (t UniStorage) Close() {
	if t.config.StorageMode == app.Database {
		t.db.Close()
	}
}

// ping database
func (t UniStorage) Ping() error {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()

		errMsg := "UniStorage.Ping error"
		backoff := func(ctx context.Context) error {
			err := t.db.Ping(dbctx)
			return app.HandleRetriableDB(err, errMsg)
		}
		err := app.DoRetry(dbctx, t.config.MaxConnectionRetries, backoff)
		//return t.db.Ping(dbctx)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: %s\n", errMsg, err))
		}
		return err
	} else {
		return nil
	}
}

// set metric by metric object provided
func (t UniStorage) SetMetric(name string, metric Metric) {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()

		errMsg := "UniStorage.SetMetric error"
		backoff := func(ctx context.Context) error {
			err := t.db.SetMetric(dbctx, name, metric)
			return app.HandleRetriableDB(err, errMsg)
		}
		err := app.DoRetry(dbctx, t.config.MaxConnectionRetries, backoff)
		//err := t.db.SetMetric(dbctx, name, metric)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: %s\n", errMsg, err))
		}
	} else {
		t.stor.SetMetric(name, metric)
	}
}

// load metric storage state from file
func (t UniStorage) LoadState(path string) error {
	if t.config.StorageMode == app.Database {
		//not needed in DB mode
		return nil
	} else {
		return t.stor.LoadState(path)
	}
}

// save metric storage state to file
func (t UniStorage) SaveState(path string) error {
	if t.config.StorageMode == app.Database {
		//not needed in DB mode
		return nil
	} else {
		return t.stor.SaveState(path)
	}
}

// get metric by its name
func (t UniStorage) GetMetric(id string) (Metric, error) {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()
		var metric Metric

		errMsg := "UniStorage.GetMetric error"
		backoff := func(ctx context.Context) error {
			var err error
			metric, err = t.db.GetMetric(dbctx, id)
			return app.HandleRetriableDB(err, errMsg)
		}
		err := app.DoRetry(dbctx, t.config.MaxConnectionRetries, backoff)
		//return t.db.GetMetric(dbctx, id)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: %s\n", errMsg, err))
			return nil, err
		}
		return metric, err
	} else {
		val, ok := t.stor.Metrics[id]
		if !ok {
			return nil, fmt.Errorf("metric not found: %s", id)
		}
		return val, nil
	}
}

// Get a copy of Metric storage
func (t UniStorage) GetMetrics() map[string]Metric {

	// Create the target map
	targetMap := make(map[string]Metric)

	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()

		errMsg := "UniStorage.GetMetrics error"
		backoff := func(ctx context.Context) error {
			var err error
			targetMap, err = t.db.GetMetrics(dbctx)
			return app.HandleRetriableDB(err, errMsg)
		}
		err := app.DoRetry(dbctx, t.config.MaxConnectionRetries, backoff)
		//targetMap, err = t.db.GetMetrics(dbctx) //original w/o retries
		if err != nil {
			logger.Error(fmt.Sprintf("%s: %s\n", errMsg, err))
			// return empty map
			return make(map[string]Metric)
		}
		return targetMap
	} else {
		// Get copy of original map
		for key, value := range t.stor.Metrics {
			targetMap[key] = value
		}
	}
	return targetMap
}

// update metric using string-represented value
func (t UniStorage) UpdateMetricS(mType string, mName string, mValue string) error {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()

		errMsg := "UniStorage.UpdateMetricS error"
		backoff := func(ctx context.Context) error {
			err := t.db.UpdateMetricS(dbctx, mType, mName, mValue)
			return app.HandleRetriableDB(err, errMsg)
		}
		err := app.DoRetry(dbctx, t.config.MaxConnectionRetries, backoff)
		//return t.db.UpdateMetricS(dbctx, mType, mName, mValue)
		if err != nil {
			logger.Error(fmt.Sprintf("%s: %s\n", errMsg, err))
		}
		return err
	} else {
		return t.stor.UpdateMetricS(mType, mName, mValue)
	}
}

// mass metric update
func (t UniStorage) BatchUpdateMetrics(m *[]domain.Metrics, errs *[]error) *[]domain.Metrics {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()

		var mr *[]domain.Metrics

		errMsg := "UniStorage.UpdateMetricS error"
		backoff := func(ctx context.Context) error {
			var err error
			mr, err = t.db.BatchUpdateMetrics(dbctx, m, errs)
			return app.HandleRetriableDB(err, errMsg)
		}
		err := app.DoRetry(dbctx, t.config.MaxConnectionRetries, backoff)
		//return t.db.BatchUpdateMetrics(dbctx, m)
		if err != nil {
			*errs = append(*errs, err)
			logger.Error(fmt.Sprintf("%s: %s\n", errMsg, err))
		}

		return mr
	} else {
		return t.stor.BatchUpdateMetrics(m, errs)
	}
}
