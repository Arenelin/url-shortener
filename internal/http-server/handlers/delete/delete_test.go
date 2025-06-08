package delete_test

import (
	"encoding/json"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"url-shortener/internal/http-server/handlers/delete"
	"url-shortener/internal/http-server/handlers/delete/mocks"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
	"url-shortener/internal/storage"
)

func TestDeleteHandler(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		alias   string
		status  string
		respErr string
		mockErr error
	}{
		{
			name:   "Good request with alias",
			alias:  "awesome-project",
			status: resp.StatusOK,
		},
		{
			name:    "Empty alias",
			alias:   "",
			respErr: "invalid request",
			status:  resp.StatusError,
		},
		{
			name:    "Not found URL",
			url:     "https://site.com",
			alias:   "site",
			respErr: "url not found",
			mockErr: storage.ErrURLNotFound,
			status:  resp.StatusError,
		},
		{
			name:    "Unexpected database error",
			alias:   "site",
			respErr: "internal error",
			mockErr: errors.New("unexpected error"),
			status:  resp.StatusError,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {

			urlDeleteMock := mocks.NewURLDelete(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlDeleteMock.On("DeleteURL", mock.AnythingOfType("string")).
					Return(tc.mockErr).
					Once()
			}
			handler := delete.New(slogdiscard.NewDiscardLogger(), urlDeleteMock)
			r := chi.NewRouter()
			r.Delete("/{alias}", handler)
			r.Delete("/", handler) // for the case of an empty alias

			ts := httptest.NewServer(r)
			defer ts.Close()

			req, err := http.NewRequest(http.MethodDelete, ts.URL+"/"+tc.alias, nil)
			require.NoError(t, err)

			client := &http.Client{}
			re, err := client.Do(req)
			require.NoError(t, err)
			defer re.Body.Close()

			body, err := io.ReadAll(re.Body)
			require.NoError(t, err)
			var response resp.Response

			if tc.respErr != "" {
				require.NoError(t, json.Unmarshal([]byte(body), &response))
				require.Equal(t, tc.status, response.Status)
			}

			require.Equal(t, tc.respErr, response.Error)
		})
	}
}
