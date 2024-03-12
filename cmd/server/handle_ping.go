package main

import (
	"database/sql"
	"fmt"
	"net/http"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// HTTP request processing
func pingDBServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")

	conn, err := sql.Open("pgx", sc.DatabaseDSN)
	if err != nil {
		http.Error(w, fmt.Sprintf("error connecting to DB server: %s", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	body := "Ping OK"
	_, _ = w.Write([]byte(body))

	//w.WriteHeader(http.StatusOK)
}
