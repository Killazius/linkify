package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"linkify/internal/transport/handlers/url/save"
	mocker "linkify/internal/transport/handlers/url/save/mocks"
	"linkify/pkg/logger/zapdiscard"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		respError  string
		mockError  error
		statusCode int
		cacheError error
		cacheURL   string
		body       string
	}{
		{
			name:       "Success",
			url:        "https://google.com",
			statusCode: http.StatusOK,
			body:       fmt.Sprintf(`{"url": "%s"}`, "https://google.com"),
		},
		{
			name:       "Empty URL",
			url:        "",
			respError:  "field URL is required",
			statusCode: http.StatusBadRequest,
			body:       fmt.Sprintf(`{"url": "%s"}`, ""),
		},
		{
			name:       "Invalid URL",
			url:        "some invalid URL",
			respError:  "field URL is not a valid URL",
			statusCode: http.StatusBadRequest,
			body:       fmt.Sprintf(`{"url": "%s"}`, "some invalid URL"),
		},
		{
			name:       "Save Error",
			url:        "https://google.com",
			respError:  "failed to generate unique alias",
			mockError:  errors.New("unexpected error"),
			statusCode: http.StatusInternalServerError,
			body:       fmt.Sprintf(`{"url": "%s"}`, "https://google.com"),
		},
		{
			name:       "Invalid JSON",
			url:        "",
			respError:  "failed to decode request",
			statusCode: http.StatusBadRequest,
			body:       `{"url": "https://google.com"`,
		},
	}
	const aliasLength = 6
	t.Parallel()
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocker.NewURLSaver(t)
			cacheSaverMock := mocker.NewCacheSaver(t)
			metricsSaverMock := mocker.NewMetricsSaver(t)
			metricsSaverMock.On("IncLinksCreated").Maybe()
			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("Save", tc.url, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
					Return(tc.mockError).
					Once()

				if tc.mockError == nil {
					cacheSaverMock.On("Set", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string"), tc.url, mock.AnythingOfType("time.Duration")).
						Return(nil).
						Once()
				}
			}
			handler := save.New(zapdiscard.New(), urlSaverMock, cacheSaverMock, aliasLength, metricsSaverMock)

			req, err := http.NewRequest(http.MethodPost, "/url", bytes.NewReader([]byte(tc.body)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			urlSaverMock.AssertExpectations(t)
			cacheSaverMock.AssertExpectations(t)
			if tc.respError == "" {
				require.NotEmpty(t, resp.Alias)
				require.Len(t, resp.Alias, aliasLength)
				require.WithinDuration(t, time.Now(), resp.CreatedAt, time.Second)
			} else {
				require.Empty(t, resp.Alias)
				require.True(t, resp.CreatedAt.IsZero())
			}
		})
	}
}
