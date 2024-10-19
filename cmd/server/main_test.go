package main

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"internal/app"
	"internal/ports/storage"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var _ = func() bool {
	var testSc = app.ServerConfig{StorageMode: app.Memory}
	stor = storage.NewUniStorage(&testSc)
	var tm = &storage.Gauge{Value: 4.5}
	stor.SetMetric("main_test", tm)

	testing.Init()
	return true
}()

// test suite for index() handler
func Test_index(t *testing.T) {

	type want struct {
		contentType string
		bodyHeader  string
		statusCode  int
	}

	tests := []struct {
		name    string
		handler func(http.ResponseWriter, *http.Request)
		request string
		method  string
		params  map[string]string
		body    string
		want    want
	}{
		{
			name: "001-01 ROOT positive test",
			want: want{
				contentType: "text/html; charset=utf-8",
				statusCode:  200,
				bodyHeader:  "<html>",
			},
			handler: handleIndex,
			request: "/",
			method:  http.MethodGet,
		},
		{
			name: "001-02 ROOT negative test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				bodyHeader:  "",
			},
			handler: handleIndex,
			request: "/bla",
			method:  http.MethodGet,
		},
		{
			name: "002-01 PING negative test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusBadRequest,
				bodyHeader:  "",
			},
			handler: handlePingDBServer,
			request: "/ping",
			method:  http.MethodGet,
		},
		{
			name: "003-01 UPDATEv1 positive Gauge test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV1,
			request: "/update/gauge/DemoGauge/1.1",
			method:  http.MethodPost,
			params: map[string]string{
				"type":  "gauge",
				"name":  "DemoGauge",
				"value": "1.1",
			},
		},
		{
			name: "003-02 UPDATEv1 positive Counter test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV1,
			request: "/update/counter/DemoCounter/25",
			method:  http.MethodPost,
			params: map[string]string{
				"type":  "counter",
				"name":  "DemoCounter",
				"value": "25",
			},
		},
		{
			name: "003-03 UPDATEv1 negative Gauge test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV1,
			request: "/update/gaugeZ/DemoGauge/1.1",
			method:  http.MethodPost,
			params: map[string]string{
				"type":  "gaugeZ",
				"name":  "DemoGauge",
				"value": "1.1",
			},
		},
		{
			name: "003-04 UPDATEv1 negative Counter test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV1,
			request: "/update/counterZ/DemoCounter/25",
			method:  http.MethodPost,
			params: map[string]string{
				"type":  "counterZ",
				"name":  "DemoCounter",
				"value": "25",
			},
		},
		{
			name: "004-01 REQUESTv1 positive Gauge test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				bodyHeader:  "1.1",
			},
			handler: handleRequestMetricV1,
			request: "/value/gauge/DemoGauge",
			method:  http.MethodGet,
			params: map[string]string{
				"type": "gauge",
				"name": "DemoGauge",
			},
		},
		{
			name: "004-02 REQUESTv1 positive Counter test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				bodyHeader:  "25",
			},
			handler: handleRequestMetricV1,
			request: "/value/counter/DemoCounter",
			method:  http.MethodGet,
			params: map[string]string{
				"type": "counter",
				"name": "DemoCounter",
			},
		},
		{
			name: "004-03 REQUESTv1 negative Gauge test (bad type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV1,
			request: "/value/gaugeZ/DemoGauge",
			method:  http.MethodGet,
			params: map[string]string{
				"type": "gaugeZ",
				"name": "DemoGauge",
			},
		},
		{
			name: "004-04 REQUESTv1 negative Counter test (bad type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV1,
			request: "/value/counterZ/DemoCounter",
			method:  http.MethodGet,
			params: map[string]string{
				"type": "counterZ",
				"name": "DemoCounter",
			},
		},
		{
			name: "004-05 REQUESTv1 negative Counter test (bad name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV1,
			request: "/value/counterZ/DemoCounter",
			method:  http.MethodGet,
			params: map[string]string{
				"type": "counter",
				"name": "DemoCounterMissing",
			},
		},
		{
			name: "005-01 UPDATEv2 positive Gauge test",
			want: want{
				contentType: "application/json",
				statusCode:  200,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `{"type": "gauge", "id": "DemoGaugev2", "value": 3.3}`,
		},
		{
			name: "005-02 UPDATEv2 positive Counter test",
			want: want{
				contentType: "application/json",
				statusCode:  200,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `{"type": "counter", "id": "DemoCounterV2", "delta": 30}`,
		},
		{
			name: "005-03 UPDATEv2 negative Gauge test (wrong type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `{"type": "gaugeZ", "id": "DemoGaugev2", "value": 3.3}`,
		},
		{
			name: "005-04 UPDATEv2 negative Counter test (wrong type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `{"type": "counterZ", "id": "DemoCounterV2", "delta": 30}`,
		},
		{
			name: "005-05 UPDATEv2 negative Gauge test (empty deserialized name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `{"type": "gaugeZ", "name": "DemoGaugev2", "value": 3.3}`,
		},
		{
			name: "005-06 UPDATEv2 negative Counter test (empty deserialized name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `{"type": "counterZ", "name": "DemoCounterV2", "delta": 30}`,
		},
		{
			name: "005-07 UPDATEv2 negative Counter test (bad JSON)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetricV2,
			request: "/update/",
			method:  http.MethodPost,
			body:    `this is just poorly-formed JSON`,
		},
		{
			name: "006-01 REQUESTv2 positive Gauge test",
			want: want{
				contentType: "application/json",
				statusCode:  200,
				bodyHeader:  `"value":3.3`,
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "gauge", "id": "DemoGaugev2"}`,
		},
		{
			name: "006-02 REQUESTv2 positive Counter test",
			want: want{
				contentType: "application/json",
				statusCode:  200,
				bodyHeader:  `"delta":30`,
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "counter", "id": "DemoCounterV2"}`,
		},
		{
			name: "006-03 REQUESTv2 negative Gauge test (wrong type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "gaugeZ", "id": "DemoGaugev2"}`,
		},
		{
			name: "006-04 REQUESTv2 negative Counter test (wrong type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "counterZ", "id": "DemoCounterV2"}`,
		},
		{
			name: "006-05 REQUESTv2 negative Gauge test (empty deserialized name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "gaugeZ", "name": "DemoGaugev2"}`,
		},
		{
			name: "006-06 REQUESTv2 negative Counter test (empty deserialized name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "counterZ", "name": "DemoCounterV2"}`,
		},
		{
			name: "006-07 REQUESTv2 negative Gauge test (wrong name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "gauge", "id": "DemoGaugev2_ERR"}`,
		},
		{
			name: "006-08 REQUESTv2 negative Counter test (wrong name)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  404,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `{"type": "counter", "id": "DemoCounterV2_ERR"}`,
		},
		{
			name: "006-09 REQUESTv2 negative Counter test (bad JSON)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleRequestMetricV2,
			request: "/value/",
			method:  http.MethodPost,
			body:    `this is just poorly-formed JSON`,
		},
		{
			name: "007-01 MASS UPDATE positive test",
			want: want{
				contentType: "application/json",
				statusCode:  200,
				bodyHeader:  `"delta":100`,
			},
			handler: handleUpdateMetrics,
			request: "/updates/",
			method:  http.MethodPost,
			body:    `[{"type": "gauge", "id": "DemoGaugeMASS", "value": 5.5},{"type": "counter", "id": "DemoCounterMASS", "delta": 100}]`,
		},
		{
			name: "007-02 MASS UPDATE negative test (bad type)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetrics,
			request: "/updates/",
			method:  http.MethodPost,
			body:    `[{"type": "gaugeZ", "id": "DemoGaugeMASS", "value": 5.5},{"type": "counterZ", "id": "DemoCounterMASS", "delta": 100}]`,
		},
		{
			name: "007-03 MASS UPDATE negative test (bad JSON)",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  400,
				bodyHeader:  "",
			},
			handler: handleUpdateMetrics,
			request: "/updates/",
			method:  http.MethodPost,
			body:    `this is just poorly-formed JSON`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := httptest.NewRequest(tt.method, tt.request, strings.NewReader(tt.body))

			rctx := chi.NewRouteContext()
			for k, v := range tt.params {
				rctx.URLParams.Add(k, v)
			}

			r = r.WithContext(context.WithValue(r.Context(), chi.RouteCtxKey, rctx))

			tt.handler(w, r)
			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			bodyString := string(bodyBytes)
			bodyMatch := strings.Contains(bodyString, tt.want.bodyHeader)
			if !bodyMatch {
				fmt.Printf("body mismatch: expected [%s] in [%s]\n", tt.want.bodyHeader, bodyString)
			}
			assert.True(t, bodyMatch)
		})
	}

}
