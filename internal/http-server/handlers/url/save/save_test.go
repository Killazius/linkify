package save_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"shorturl/internal/http-server/handlers/url/save"
	"shorturl/internal/http-server/handlers/url/save/mocks"
	"shorturl/internal/lib/logger/handlers/slogdiscard"
	"testing"
)

func TestSaveHandler(t *testing.T) {
	cases := []struct {
		name       string
		url        string
		respError  string
		mockError  error
		statusCode int
	}{
		{
			name:       "Success",
			url:        "https://google.com",
			statusCode: http.StatusOK,
		},
		{
			name:       "Empty URL",
			url:        "",
			respError:  "field URL is required",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Invalid URL",
			url:        "some invalid URL",
			respError:  "field URL is not a valid URL",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "SaveURL Error",
			url:        "https://google.com",
			respError:  "failed to save url",
			mockError:  errors.New("unexpected error"),
			statusCode: http.StatusInternalServerError,
		},
	}
	const aliasLength = 6
	t.Parallel()
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlSaverMock := mocks.NewURLSaver(t)
			cacheSaverMock := mocks.NewCacheSaver(t)

			if tc.respError == "" || tc.mockError != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string"), mock.AnythingOfType("time.Time")).
					Return(tc.mockError).
					Once()

				if tc.mockError == nil {
					cacheSaverMock.On("Set", mock.AnythingOfType("context.backgroundCtx"), mock.AnythingOfType("string"), tc.url, mock.AnythingOfType("time.Duration")).
						Return(nil).
						Once()
				}
			}
			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock, cacheSaverMock, aliasLength)

			input := fmt.Sprintf(`{"url": "%s"}`, tc.url)

			req, err := http.NewRequest(http.MethodPost, "/save", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			body := rr.Body.String()

			var resp save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &resp))

			require.Equal(t, tc.respError, resp.Error)

			// TODO: add more checks
		})
	}
}
