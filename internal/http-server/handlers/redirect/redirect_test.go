package redirect_test

import (
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/go-playground/assert/v2"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/redirect"
	"url-shortener/internal/http-server/handlers/redirect/mocks"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestRedirectHandler(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		alias   string
		respErr error
		mockErr error
	}{
		{
			name:  "Good request with alias",
			url:   "https://google.com",
			alias: "awesome-project",
		},
		{
			name:    "Empty alias",
			url:     "https://google.com",
			alias:   "",
			respErr: fmt.Errorf("%s: %w: %d", api.GetRedirectOp, api.ErrInvalidStatusCode, http.StatusNotFound),
		},
		{
			name:    "Not found URL",
			url:     "https://site.com",
			alias:   "site",
			respErr: fmt.Errorf("%s: %w: %d", api.GetRedirectOp, api.ErrInvalidStatusCode, http.StatusOK),
			mockErr: errors.New("url not found"),
		},
		{
			name:    "Unexpected database error",
			url:     "https://google.com",
			alias:   "site",
			respErr: fmt.Errorf("%s: %w: %d", api.GetRedirectOp, api.ErrInvalidStatusCode, http.StatusOK),
			mockErr: errors.New("unexpected error"),
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			urlGetterMock := mocks.NewURLGetter(t)

			if tc.respErr == nil || tc.mockErr != nil {
				urlGetterMock.On("GetURL", mock.AnythingOfType("string")).
					Return("https://google.com", tc.mockErr).
					Once()
			}
			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetterMock))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)

			//if err != nil {
			require.Equal(t, tc.respErr, err)
			//} else {
			assert.Equal(t, tc.url, redirectedToURL)
			//}
		})
	}
}
