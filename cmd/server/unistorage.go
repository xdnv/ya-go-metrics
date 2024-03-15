package main

import (
	"context"
	"database/sql"
	"fmt"
	"internal/adapters/logger"
	"internal/app"
	"internal/ports/storage"
	"time"
)

// universal metric storage
type UniStorage struct {
	config  *app.ServerConfig
	ctx     context.Context
	stor    *storage.MemStorage
	db      *storage.DbStorage
	timeout time.Duration
}

// init metric storage
func NewUniStorage(cf *app.ServerConfig) *UniStorage {

	if cf.StorageMode == app.Database {
		var (
			conn *sql.DB
			err  error
		)

		conn, err = sql.Open("pgx", sc.DatabaseDSN)
		if err != nil {
			logger.Fatal(err.Error())
		}

		return &UniStorage{
			config:  cf,
			ctx:     context.Background(),
			db:      storage.NewDbStorage(conn),
			timeout: 5 * time.Second,
		}
	} else {
		return &UniStorage{
			config: cf,
			ctx:    context.Background(),
			stor:   storage.NewMemStorage(),
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
		return t.db.PingContext(dbctx)
	} else {
		return nil
	}
}

func (t UniStorage) SetMetric(name string, metric storage.Metric) {
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

func (t UniStorage) GetMetric(id string) (storage.Metric, error) {
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
func (t UniStorage) GetMetrics() map[string]storage.Metric {

	// Create the target map
	targetMap := make(map[string]storage.Metric)

	if t.config.StorageMode == app.Database {
		//TODO: implement SQL logic
	} else {
		// Copy from the original map to the target map
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
