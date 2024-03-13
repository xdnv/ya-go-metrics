package storage

import (
	"context"
	"database/sql"
	"time"
)

// main metric storage
type DbStorage struct {
	conn *sql.DB
}

// NewDbStorage returns new PostgreSQL Metric storage
func NewDbStorage(conn *sql.DB) *DbStorage {
	return &DbStorage{conn: conn}
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
	tableName := "public.config"
	dbKey := "DBVersion"
	dbVersion := "20240313"

	// has, err := t.TableExists(ctx, tx, tableName)
	// if err != nil {
	// 	return err
	// }

	// if !has {
	// config table stores app config entries
	//TODO: add version update procedure
	tx.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS @tableName (
			key VARCHAR(128) NOT NULL PRIMARY KEY,
			value TEXT
		);
		INSERT INTO @tableName (key, value)
			VALUES (@dbKey, @dbVersion)
		ON CONFLICT (key)
			DO UPDATE SET value = excluded.value;
	`,
		sql.Named("tableName", tableName),
		sql.Named("dbKey", dbKey),
		sql.Named("dbVersion", dbVersion),
	)
	// }

	// gauge metrics
	tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS public.gauges (
            id  VARCHAR(128) NOT NULL PRIMARY KEY,
            value DOUBLE PRECISION NOT NULL 
        )
    `)

	// counter metrics
	tx.ExecContext(ctx, `
        CREATE TABLE IF NOT EXISTS public.counters (
            id VARCHAR(128) NOT NULL PRIMARY KEY,
            value BIGINT NOT NULL
        )
    `)

	// commit transaction
	return tx.Commit()
}

func (t DbStorage) PingContext(ctx context.Context) error {

	dbctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	return t.conn.PingContext(dbctx)
}
