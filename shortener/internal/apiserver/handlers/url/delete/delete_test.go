package delete_test

import (
	"context"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	del "linkify/internal/apiserver/handlers/url/delete"
	mocker "linkify/internal/apiserver/handlers/url/delete/mocks"
	"linkify/internal/lib/logger/zapdiscard"
	"linkify/internal/storage"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDeleteHandler(t *testing.T) {
	cases := []struct {
		name       string
		alias      string
		mockError  error
		cacheError error
		statusCode int
	}{
		{
			name:       "Success",
			alias:      "alias",
			statusCode: http.StatusNoContent,
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
			statusCode: http.StatusNotFound,
		},
		{
			name:       "DeleteURL error",
			alias:      "alias",
			mockError:  errors.New("failed to delete URL"),
			statusCode: http.StatusInternalServerError,
		},
		{
			name:       "Cache deletion error",
			alias:      "alias",
			mockError:  nil,
			cacheError: errors.New("cache error"),
			statusCode: http.StatusInternalServerError,
		},
	}

	t.Parallel()
	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			urlDeleterMock := mocker.NewURLDeleter(t)
			cacheDeleterMock := mocker.NewCacheDeleter(t)
			metricsDeleterMock := mocker.NewMetricsDeleter(t)
			metricsDeleterMock.On("IncLinksDeleted").Maybe()
			if tc.alias != "" {
				urlDeleterMock.On("DeleteURL", tc.alias).Return(tc.mockError).Once()

				if tc.mockError == nil {
					cacheDeleterMock.On("Delete", mock.Anything, tc.alias).
						Return(tc.cacheError).
						Once()
				}
			}

			handler := del.New(zapdiscard.New(), urlDeleterMock, cacheDeleterMock, metricsDeleterMock)
			url := fmt.Sprintf("/url/%s", tc.alias)
			req, err := http.NewRequest(http.MethodDelete, url, nil)
			require.NoError(t, err)

			rctx := chi.NewRouteContext()
			rctx.URLParams.Add("alias", tc.alias)
			req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			require.Equal(t, tc.statusCode, rr.Code)

			urlDeleterMock.AssertExpectations(t)
			cacheDeleterMock.AssertExpectations(t)
		})
	}
}
