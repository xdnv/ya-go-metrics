package main

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_index(t *testing.T) {
	type want struct {
		contentType string
		statusCode  int
		bodyHeader  string
	}
	tests := []struct {
		name    string
		request string
		want    want
	}{
		{
			name: "001 positive root test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  200,
				bodyHeader:  "<html>",
			},
			request: "/",
		},
		{
			name: "002 negative root test",
			want: want{
				contentType: "text/plain; charset=utf-8",
				statusCode:  http.StatusNotFound,
				bodyHeader:  "",
			},
			request: "/bla",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := httptest.NewRequest(http.MethodPost, tt.request, nil)
			w := httptest.NewRecorder()
			index(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			bodyBytes, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			bodyString := string(bodyBytes)

			assert.True(t, strings.Contains(bodyString, tt.want.bodyHeader))
		})
	}

}
