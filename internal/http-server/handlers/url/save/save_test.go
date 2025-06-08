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
	"testing"
	_ "url-shortener/internal/config"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/http-server/handlers/url/save/mocks"
	resp "url-shortener/internal/lib/api/response"
	"url-shortener/internal/lib/logger/handlers/slogdiscard"
)

func TestSaveHandler(t *testing.T) {
	testCases := []struct {
		name    string
		url     string
		alias   string
		respErr string
		status  string
		mockErr error
	}{
		{
			name:   "Good request with alias",
			url:    "https://google.com",
			alias:  "awesome-project",
			status: resp.StatusOK,
		},
		{
			name:    "Empty URL",
			url:     "",
			alias:   "some_alias",
			respErr: "field URL is a required field",
			status:  resp.StatusError,
		},
		{
			name:   "Empty alias",
			url:    "https://site.com",
			alias:  "",
			status: resp.StatusOK,
		},
		{
			name:    "Invalid URL with empty alias",
			url:     "Hello world",
			alias:   "",
			respErr: "field URL is not a valid URL",
			status:  resp.StatusError,
		},
		{
			name:    "Invalid URL with correct alias",
			url:     "Hello world",
			alias:   "hey-hey",
			respErr: "field URL is not a valid URL",
			status:  resp.StatusError,
		},
		{
			name:    "Duplicate url",
			url:     "https://google.com",
			alias:   "lala",
			mockErr: errors.New("duplicate url"),
			respErr: "failed to add url",
			status:  resp.StatusError,
		},
		{
			name:    "Unexpected database error",
			url:     "https://test.com",
			alias:   "hello!",
			mockErr: errors.New("unexpected error"),
			respErr: "failed to add url",
			status:  resp.StatusError,
		},
	}

	for _, tc := range testCases {

		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("CONFIG_PATH", "../../../../../config/local.yaml")

			urlSaverMock := mocks.NewURLSaver(t)

			if tc.respErr == "" || tc.mockErr != nil {
				urlSaverMock.On("SaveURL", tc.url, mock.AnythingOfType("string")).
					Return(int64(1), tc.mockErr).
					Once()
			}
			handler := save.New(slogdiscard.NewDiscardLogger(), urlSaverMock)

			input := fmt.Sprintf(`{"url": "%s", "alias": "%s"}`, tc.url, tc.alias)

			req, err := http.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(input)))
			require.NoError(t, err)

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			body := rr.Body.String()
			var response save.Response

			require.NoError(t, json.Unmarshal([]byte(body), &response))

			if tc.respErr == "" && tc.mockErr == nil {
				require.True(t, response.Alias != "")
			}

			if tc.alias == "" && tc.respErr == "" && tc.mockErr == nil {
				require.True(t, len(response.Alias) == 6)
			}

			require.Equal(t, tc.respErr, response.Error)
			require.Equal(t, tc.status, response.Status)
		})
	}
}
