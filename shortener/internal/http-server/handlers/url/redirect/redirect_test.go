package redirect_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"linkify/internal/http-server/handlers/url/redirect"
	mocker "linkify/internal/http-server/handlers/url/redirect/mocks"
	"linkify/internal/lib/logger/handlers/slogdiscard"
	"linkify/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRedirectHandler(t *testing.T) {
	cases := []struct {
		name       string
		alias      string
		mockError  error
		cacheError error
		cacheURL   string
		statusCode int
	}{
		{
			name:       "Success",
			alias:      "alias",
			cacheURL:   "http://example.com",
			statusCode: http.StatusFound,
		},
		{
			name:       "Empty alias",
			alias:      "",
			statusCode: http.StatusBadRequest,
		},
		{
			name:       "Alias not found",
			alias:      "non_existent_alias",
			mockError:  storage.ErrURLNotFound,
			cacheError: storage.ErrAliasNotFound,
			statusCode: http.StatusNotFound,
		},
		{
			name:       "GetURL error",
			alias:      "alias",
			cacheError: errors.New("cache error"),
			mockError:  errors.New("failed to get URL"),
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "Cache get error",
			alias:      "alias",
			mockError:  nil,
			cacheError: errors.New("cache error"),
			cacheURL:   "http://example.com",
			statusCode: http.StatusFound,
		},
	}

	t.Parallel()
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlGetterMock := mocker.NewURLGetter(t)
			cacheGetterMock := mocker.NewCacheGetter(t)
			metricsGetterMock := mocker.NewMetricsGetter(t)
			metricsGetterMock.On("IncLinksRedirected").Maybe()
			if tc.alias != "" {
				cacheGetterMock.On("Get", mock.Anything, tc.alias).
					Return(tc.cacheURL, tc.cacheError).
					Once()

				if tc.cacheError != nil || tc.cacheURL == "" {
					urlGetterMock.On("GetURL", tc.alias).
						Return(tc.cacheURL, tc.mockError).
						Once()
				}
			}

			handler := redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock, cacheGetterMock, metricsGetterMock)
			url := fmt.Sprintf("/%s", tc.alias)
			req, err := http.NewRequest(http.MethodGet, url, nil)
			require.NoError(t, err)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			urlGetterMock.AssertExpectations(t)
			cacheGetterMock.AssertExpectations(t)
		})
	}
}
