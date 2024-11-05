// PingDB implementation on application layer
package app

import (
	"errors"
	"fmt"
	"internal/domain"
	"net/http"
)

func PingDBServer() *domain.HandlerStatus {
	hs := new(domain.HandlerStatus)

	if Sc.StorageMode != domain.Database {
		hs.Message = "cannot ping DB connection: server does not run in Database mode"
		hs.Err = errors.New(hs.Message)
		hs.HTTPStatus = http.StatusBadRequest
		return hs
	}

	if err := Stor.Ping(); err != nil {
		hs.Message = fmt.Sprintf("error pinging DB server: %s", err)
		hs.Err = err
		hs.HTTPStatus = http.StatusInternalServerError
		return hs
	}

	hs.Message = "Ping OK"
	return hs
}
