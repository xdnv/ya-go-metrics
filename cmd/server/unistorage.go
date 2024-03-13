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
func NewUniStorage(cf *app.ServerConfig) *UniStorage {

	if cf.StorageMode == app.Database {
		var (
			conn *sql.DB
			err  error
		)
		if sc.StorageMode == app.Database {
			conn, err = sql.Open("pgx", sc.DatabaseDSN)
			if err != nil {
				panic(err)
			}

		}

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

func (t UniStorage) Close() {
	if t.config.StorageMode == app.Database {
		t.db.Close()
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

func (t UniStorage) SetMetric(name string, metric storage.Metric) {
	if t.config.StorageMode == app.Database {
		//TODO: implement SQL logic
	} else {
		t.stor.SetMetric(name, metric)
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
		return errors.ErrUnsupported
	} else {
		return t.stor.UpdateMetricS(mType, mName, mValue)
	}
}
