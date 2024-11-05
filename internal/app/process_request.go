// RequestMetricV1 & RequestMetricV2 implementation on application layer
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

// HTTP single metric request v1 processing
func RequestMetricV1(mr *domain.MetricRequest) *domain.HandlerStatus {
	hs := new(domain.HandlerStatus)

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		hs.Message = fmt.Sprintf("unexpected metric type: %s", mr.Type)
		hs.Err = errors.New(hs.Message)
		hs.HTTPStatus = http.StatusBadRequest
		return hs
	}

	metric, err := Stor.GetMetric(mr.Name)
	if err != nil {
		hs.Message = fmt.Sprintf("ERROR getting metric [%s]: %s", mr.Name, err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusNotFound
		return hs
	}

	//write metric value in plaintext
	hs.Message = fmt.Sprintf("%v", metric.GetValue())

	return hs
}

// HTTP single metric request v2 processing
func RequestMetricV2(data io.Reader) (*[]byte, *domain.HandlerStatus) {
	hs := new(domain.HandlerStatus)

	var m domain.Metrics

	//logger.Debug(fmt.Sprintf("RequestMetricV2 body: %v", data)) //DEBUG

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

	//type validation
	switch mr.Type {
	case "gauge":
	case "counter":
	default:
		hs.Message = fmt.Sprintf("unexpected metric type: %s", mr.Type)
		hs.Err = errors.New(hs.Message)
		hs.HTTPStatus = http.StatusNotFound
		return nil, hs
	}

	metric, err := Stor.GetMetric(mr.Name)
	if err != nil {
		hs.Message = fmt.Sprintf("ERROR getting metric [%s]: %s", mr.Name, err.Error())
		hs.Err = err
		hs.HTTPStatus = http.StatusNotFound
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
