package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
	//_ "github.com/lib/pq"
)

// main metric storage
type DbStorage struct {
	conn *sql.DB
}

// NewDbStorage returns new PostgreSQL Metric storage
func NewDbStorage(conn *sql.DB) *DbStorage {
	return &DbStorage{conn: conn}
}

func (t DbStorage) Close() {
	t.conn.Close()
}

// // Check if table exists
// func (t DbStorage) TableExists(ctx context.Context, tx *sql.Tx, tableName string) (bool, error) {
// 	row := tx.QueryRowContext(ctx, `
// 		SELECT to_regclass('@tableName');
// 		`,
// 		sql.Named("tableName", tableName))

// 	var (
// 		result sql.NullString
// 	)
// 	err := row.Scan(&result)
// 	if err != nil {
// 		return false, err
// 	}

// 	return result.Valid, nil
// }

// prepare database
func (t DbStorage) Bootstrap(ctx context.Context) error {

	// begin transaction
	tx, err := t.conn.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	//we get db name in DSN spec
	//dbName := "ya_metrics"

	// a := `
	// 	INSERT INTO the_table (id, column_1, column_2)
	// 		VALUES (1, 'A', 'X'), (2, 'B', 'Y'), (3, 'C', 'Z')
	// 	ON CONFLICT (id) DO UPDATE
	// 		SET column_1 = excluded.column_1,
	//   			column_2 = excluded.column_2;
	//   `

	//Check if db exists
	// row := tx.QueryRowContext(ctx, `
	// 	SELECT datname FROM pg_catalog.pg_database WHERE datname=@dbname
	// `,
	// 	sql.Named("dbname", dbName))

	//check config
	//tableName := "public.config"
	dbKey := "DBVersion"
	dbVersion := "20240313"

	// has, err := t.TableExists(ctx, tx, tableName)
	// if err != nil {
	// 	return err
	// }

	//Important! pgx does not support sql.Named(), use pgx.NamedArgs{} instead

	// if !has {
	// config table stores app config entries
	//TODO: add version update procedure
	_, err = tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS public.config (
			key VARCHAR(128) NOT NULL PRIMARY KEY,
			value TEXT
		);
	`) //,
	//sql.Named("tableName", tableName),
	//)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO public.config (key, value)
			VALUES (@dbKey::text, @dbVersion::text)
		ON CONFLICT (key)
			DO UPDATE SET value = excluded.value;
	`,
		//sql.Named("tableName", tableName),
		//sql.Named("dbkey", dbKey),
		pgx.NamedArgs{
			"dbKey":     dbKey,
			"dbVersion": dbVersion,
		},
	)
	if err != nil {
		return err
	}
	// }

	// gauge metrics
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS public.gauges (
            id  VARCHAR(128) NOT NULL PRIMARY KEY,
            value DOUBLE PRECISION NOT NULL 
        );
    `)
	if err != nil {
		return err
	}

	// counter metrics
	_, err = tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS public.counters (
            id VARCHAR(128) NOT NULL PRIMARY KEY,
            value BIGINT NOT NULL
        );
    `)
	if err != nil {
		return err
	}

	// commit transaction
	return tx.Commit()
}

func (t DbStorage) PingContext(ctx context.Context) error {

	dbctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return t.conn.PingContext(dbctx)
}

// assign metric object to certain name. use with caution, TODO: replace with safer API
func (t DbStorage) SetMetric(ctx context.Context, name string, metric Metric) error {

	mType := metric.GetType()
	query := ""

	switch mType {
	case "gauge":
		query = `
		INSERT INTO public.gauges (id, value)
			VALUES (@id::text, @value::double precision)
		ON CONFLICT (id)
			DO UPDATE SET value = excluded.value;
	`
	case "counter":
		query = `
		INSERT INTO public.counters (id, value)
			VALUES (@id::text, @value::bigint)
		ON CONFLICT (id)
			DO UPDATE SET value = excluded.value;
	`
	default:
		return fmt.Errorf("unexpected metric type: %s", mType)
	}

	_, err := t.conn.ExecContext(ctx, query,
		pgx.NamedArgs{
			"id":    name,
			"value": metric.GetValue(),
		},
	)

	return err
}

func (t DbStorage) UpdateMetricS(ctx context.Context, mType string, mName string, mValue string) error {

	var val interface{}
	var err error
	query := ""

	switch mType {
	case "gauge":
		val, err = strconv.ParseFloat(mValue, 64)
		if err != nil {
			return err
		}
		query = `
		INSERT INTO public.gauges (id, value)
			VALUES (@id::text, @value::double precision)
		ON CONFLICT (id)
			DO UPDATE SET value = excluded.value;
	`
	case "counter":
		val, err = strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return err
		}
		query = `
		INSERT INTO public.counters (id, value)
			VALUES (@id::text, @value::bigint)
		ON CONFLICT (id)
			DO UPDATE SET value = public.counters.value + excluded.value;
	`
	default:
		return fmt.Errorf("unexpected metric type: %s", mType)
	}

	_, err = t.conn.ExecContext(ctx, query,
		pgx.NamedArgs{
			"id":    mName,
			"value": val,
		},
	)

	return err
}

func (t DbStorage) GetMetric(ctx context.Context, id string) (Metric, error) {

	query := `
		SELECT
			'gauge' AS mtype,
			id AS id,
			value AS floatvalue,
			NULL as intvalue
		FROM
			public.gauges
		WHERE
			id = @id
		UNION ALL
		SELECT
			'counter' AS mtype,
			id AS id,
			NULL as floatvalue,
			value AS intvalue
		FROM
			public.counters
		WHERE
			id = @id;
	`

	row := t.conn.QueryRowContext(ctx, query,
		pgx.NamedArgs{
			"id": id,
		},
	)

	var (
		mType       string
		mId         string
		mFloatValue sql.NullFloat64
		mIntValue   sql.NullInt64
	)

	err := row.Scan(&mType, &mId, &mFloatValue, &mIntValue)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("metric not found: %s", id)
		}

		fmt.Printf("Scan error %s", err)
		return nil, err
	}

	var metric Metric

	switch mType {
	case "gauge":
		metric = &Gauge{Value: mFloatValue.Float64}
	case "counter":
		metric = &Counter{Value: mIntValue.Int64}
	default:
		return nil, fmt.Errorf("unexpected metric type: %s", mType)
	}

	return metric, nil
}
