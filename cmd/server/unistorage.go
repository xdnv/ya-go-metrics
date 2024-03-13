package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"internal/app"
	"internal/ports/storage"
	"time"
)

// universal metric storage
type UniStorage struct {
	config *app.ServerConfig
	ctx    context.Context
	stor   *storage.MemStorage
	db     *storage.DbStorage
}

// init metric storage
func NewUniStorage(cf *app.ServerConfig, conn *sql.DB) *UniStorage {

	if cf.StorageMode == app.Database {
		return &UniStorage{
			config: cf,
			ctx:    context.Background(),
			db:     storage.NewDbStorage(conn),
		}
	} else {
		return &UniStorage{
			config: cf,
			ctx:    context.Background(),
			stor:   storage.NewMemStorage(),
		}
	}
}

func (t UniStorage) Ping() error {
	if t.config.StorageMode == app.Database {
		dbctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
		defer cancel()
		return t.db.PingContext(dbctx)
	} else {
		return nil
	}
}

func (t UniStorage) LoadState(path string) error {
	if t.config.StorageMode == app.Database {
		return nil
	} else {
		return t.stor.LoadState(path)
	}
}

func (t UniStorage) SaveState(path string) error {
	if t.config.StorageMode == app.Database {
		return nil
	} else {
		return t.stor.SaveState(path)
	}
}

func (t UniStorage) GetMetric(metric string) (storage.Metric, error) {
	if t.config.StorageMode == app.Database {
		return nil, errors.ErrUnsupported
	} else {
		val, ok := t.stor.Metrics[metric]
		if !ok {
			return nil, fmt.Errorf("metric not found: %s", metric)
		}
		return val, nil
	}
}

func (t UniStorage) GetMetrics() storage.MetricMap {

	if t.config.StorageMode == app.Database {
		return nil
	} else {
		return t.stor.Metrics
	}
}

func (t UniStorage) UpdateMetricS(mType string, mName string, mValue string) error {
	if t.config.StorageMode == app.Database {
		return errors.ErrUnsupported
	} else {
		return t.stor.UpdateMetricS(mType, mName, mValue)
	}
}
