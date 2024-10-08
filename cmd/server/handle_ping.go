package main

import (
	"fmt"
	"net/http"

	"internal/app"
)

// HTTP request processing
func handlePingDBServer(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	//w.WriteHeader(http.StatusOK)

	if sc.StorageMode != app.Database {
		http.Error(w, "cannot ping DB connection: server does not run in Database mode", http.StatusBadRequest)
		return
	}

	// db, err := sql.Open("pgx", sc.DatabaseDSN)
	// if err != nil {
	// 	http.Error(w, fmt.Sprintf("error opening DB connection: %s", err), http.StatusInternalServerError)
	// 	return
	// }
	// defer db.Close()

	// dbctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// defer cancel()
	// if err = db.PingContext(dbctx); err != nil {
	// 	http.Error(w, fmt.Sprintf("error pinging DB server: %s", err), http.StatusInternalServerError)
	// 	return
	// }

	if err := stor.Ping(); err != nil {
		http.Error(w, fmt.Sprintf("error pinging DB server: %s", err), http.StatusInternalServerError)
		return
	}

	body := "Ping OK"
	w.Write([]byte(body))
}
