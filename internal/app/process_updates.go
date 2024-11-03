// UpdateMetrics implementation on application layer
package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"internal/domain"
	"io"
	"net/http"
	"strings"
)

// HTTP mass metric update processing
func UpdateMetrics(data io.Reader) (*[]byte, *domain.HandlerStatus) {
	hs := new(domain.HandlerStatus)

	var m []domain.Metrics
	var errs []error

	//logger.Debug(fmt.Sprintf("UpdateMetrics body: %v", data)) //DEBUG

	if err := json.NewDecoder(data).Decode(&m); err != nil {
		hs.Message = fmt.Sprintf("json metric decode error: %s", err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusBadRequest
		return nil, hs
	}

	mr := Stor.BatchUpdateMetrics(&m, &errs)

	//handling all errors encountered
	if len(errs) > 0 {
		strErrors := make([]string, len(errs))
		for i, err := range errs {
			strErrors[i] = err.Error()
		}
		hs.Message = fmt.Sprintf("bulk update errors: %s", strings.Join(strErrors, "\n"))
		hs.Err = errors.New(hs.Message)
		hs.HTTPStatus = http.StatusBadRequest
		return nil, hs
	}

	//save dump if set to immediate mode
	if (Sc.StorageMode == domain.File) && (Sc.StoreInterval == 0) {
		err := Stor.SaveState(Sc.FileStoragePath)
		if err != nil {
			hs.Message = fmt.Sprintf("failed to save server state to [%s], error: %s", Sc.FileStoragePath, err.Error())
			hs.Err = err
			hs.HTTPStatus = http.StatusInternalServerError
			return nil, hs
		}
	}

	resp, err := json.Marshal(mr)
	if err != nil {
		hs.Message = fmt.Sprintf("json metric encode error: %s", err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusInternalServerError
		return nil, hs
	}

	return &resp, hs
}
