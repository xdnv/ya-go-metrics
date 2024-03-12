package main

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// HTTP request processing
func pingDBServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	db, err := sql.Open("pgx", sc.DatabaseDSN)
	if err != nil {
		http.Error(w, fmt.Sprintf("error opening DB connection: %s", err), http.StatusInternalServerError)
		return
	}
	defer db.Close()

	dbctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err = db.PingContext(dbctx); err != nil {
		http.Error(w, fmt.Sprintf("error pinging DB server: %s", err), http.StatusInternalServerError)
		return
	}

	body := "Ping OK"
	_, _ = w.Write([]byte(body))

	//w.WriteHeader(http.StatusOK)
}
