// UpdateMetricV1 & UpdateMetricV2 implementation on application layer
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

// HTTP single metric update v1 processing
func UpdateMetricV1(mr *domain.MetricRequest) *domain.HandlerStatus {
	hs := new(domain.HandlerStatus)

	err := Stor.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		hs.Message = fmt.Sprintf("failed to update metric [%s][%s][%s], error: %s", mr.Type, mr.Name, mr.Value, err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusBadRequest
		return hs
	}

	//save dump if set to immediate mode
	if (Sc.StorageMode == domain.File) && (Sc.StoreInterval == 0) {
		err := Stor.SaveState(Sc.FileStoragePath)
		if err != nil {
			hs.Message = fmt.Sprintf("failed to save server state to [%s], error: %s", Sc.FileStoragePath, err.Error())
			hs.Err = err
			hs.HTTPStatus = http.StatusInternalServerError
			return hs
		}
	}

	hs.Message = "OK"
	return hs
}

// HTTP single metric update v2 processing
func UpdateMetricV2(data io.Reader) (*[]byte, *domain.HandlerStatus) {
	hs := new(domain.HandlerStatus)

	var m domain.Metrics

	//logger.Debugf("UpdateMetricV2 body: %v", data) //DEBUG

	if err := json.NewDecoder(data).Decode(&m); err != nil {
		hs.Message = fmt.Sprintf("json metric decode error: %s", err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusBadRequest
		return nil, hs
	}

	if strings.TrimSpace(m.ID) == "" {
		hs.Message = "json metric decode error: empty metric id"
		hs.Err = errors.New(hs.Message)
		hs.HTTPStatus = http.StatusBadRequest
		return nil, hs
	}

	mr := new(domain.MetricRequest)
	mr.Type = m.MType
	mr.Name = m.ID

	switch mr.Type {
	case "gauge":
		mr.Value = fmt.Sprint(*m.Value)
	case "counter":
		mr.Value = fmt.Sprint(*m.Delta)
	default:
		hs.Message = fmt.Sprintf("unsupported metric type: %s", mr.Type)
		hs.Err = errors.New(hs.Message)
		hs.HTTPStatus = http.StatusNotFound
		return nil, hs
	}

	err := Stor.UpdateMetricS(mr.Type, mr.Name, mr.Value)
	if err != nil {
		hs.Message = fmt.Sprintf("failed to update metric [%s][%s][%s], error: %s", mr.Type, mr.Name, mr.Value, err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusBadRequest
		return nil, hs
	}

	//save dump if set to immediate mode
	if (Sc.StorageMode == domain.File) && (Sc.StoreInterval == 0) {
		err = Stor.SaveState(Sc.FileStoragePath)
		if err != nil {
			hs.Message = fmt.Sprintf("failed to save server state to [%s], error: %s", Sc.FileStoragePath, err.Error())
			hs.Err = err
			hs.HTTPStatus = http.StatusInternalServerError
			return nil, hs
		}
	}

	metric, err := Stor.GetMetric(mr.Name)
	if err != nil {
		hs.Message = fmt.Sprintf("ERROR getting metric [%s]: %s", mr.Name, err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusInternalServerError
		return nil, hs
	}

	//return current metric value
	switch mr.Type {
	case "gauge":
		metricValue := metric.GetValue().(float64)
		m.Value = &metricValue
	case "counter":
		metricValue := metric.GetValue().(int64)
		m.Delta = &metricValue
	}

	resp, err := json.Marshal(m)
	if err != nil {
		hs.Message = fmt.Sprintf("json metric encode error: %s", err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusInternalServerError
		return nil, hs
	}

	return &resp, hs
}
