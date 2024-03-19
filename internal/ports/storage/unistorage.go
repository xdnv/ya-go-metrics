package storage

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"internal/adapters/logger"
	"internal/app"
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

func (t UniStorage) Bootstrap() error {
	if t.config.StorageMode == app.Database {
		return t.db.Bootstrap(t.ctx)
	}

	return nil
}

func (t UniStorage) Close() {
	if t.config.StorageMode == app.Database {
		t.db.Close()
	}
}

func (t UniStorage) Ping() error {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()
		return t.db.Ping(dbctx)
	} else {
		return nil
	}
}

func (t UniStorage) SetMetric(name string, metric Metric) {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()
		err := t.db.SetMetric(dbctx, name, metric)
		if err != nil {
			logger.Error(fmt.Sprintf("UniStorage.SetMetric error: %s\n", err))
		}
	} else {
		t.stor.SetMetric(name, metric)
	}
}

func (t UniStorage) LoadState(path string) error {
	if t.config.StorageMode == app.Database {
		//not needed in DB mode
		return nil
	} else {
		return t.stor.LoadState(path)
	}
}

func (t UniStorage) SaveState(path string) error {
	if t.config.StorageMode == app.Database {
		//not needed in DB mode
		return nil
	} else {
		return t.stor.SaveState(path)
	}
}

func (t UniStorage) GetMetric(id string) (Metric, error) {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()
		return t.db.GetMetric(dbctx, id)
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
	var err error

	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()
		targetMap, err = t.db.GetMetrics(dbctx)
		if err != nil {
			logger.Error(fmt.Sprintf("UniStorage.GetMetrics error: %s\n", err))
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

func (t UniStorage) UpdateMetricS(mType string, mName string, mValue string) error {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, t.timeout)
		defer cancel()
		return t.db.UpdateMetricS(dbctx, mType, mName, mValue)
	} else {
		return t.stor.UpdateMetricS(mType, mName, mValue)
	}
}