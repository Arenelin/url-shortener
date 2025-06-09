package tests

import (
	"fmt"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/url"
	"path"
	"testing"
	"url-shortener/internal/http-server/handlers/url/save"
	"url-shortener/internal/lib/api"
	"url-shortener/internal/lib/random"
)

const (
	host = "localhost:8087"
)

func TestURLShortener_HappyPath(t *testing.T) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
	}

	e := httpexpect.Default(t, u.String())
	tempAlias := random.NewRandomString(10)

	e.POST("/url").WithJSON(save.Request{
		URL:   gofakeit.URL(),
		Alias: tempAlias,
	}).
		WithBasicAuth("user", "userpass").
		Expect().
		Status(http.StatusOK).
		JSON().
		Object().
		ContainsKey("alias").
		Value("alias").
		IsEqual(tempAlias)
}

func TestURLShortener_SaveRedirectRemove(t *testing.T) {
	duplicatedAlias := gofakeit.Word() + gofakeit.Word()

	testcases := []struct {
		name  string
		url   string
		alias string
		error string
	}{
		{
			name:  "Valid URL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
		{
			name:  "Empty alias",
			url:   gofakeit.URL(),
			alias: "",
		},
		{
			name:  "Empty URL",
			url:   "",
			alias: gofakeit.Word() + gofakeit.Word(),
			error: "field URL is a required field",
		},
		{
			name:  "Invalid URL with empty alias",
			url:   "Hello world",
			alias: "",
			error: "field URL is not a valid URL",
		},
		{
			name:  "Invalid URL with correct alias",
			url:   "Hello world",
			alias: "hey-hey",
			error: "field URL is not a valid URL",
		},
		{
			name:  "First record for duplicate url",
			url:   "https://google.com",
			alias: duplicatedAlias,
		},
		{
			name:  "Duplicate url",
			url:   "https://google.com",
			alias: duplicatedAlias,
			error: "url already exists",
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   host,
			}

			e := httpexpect.Default(t, u.String())

			// Save
			resp := e.POST("/url").WithJSON(save.Request{
				URL:   tc.url,
				Alias: tc.alias,
			}).
				WithBasicAuth("user", "userpass").
				Expect().
				Status(http.StatusOK).
				JSON().
				Object()

			if tc.error != "" {
				resp.NotContainsKey("alias")
				resp.Value("error").String().IsEqual(tc.error)

				return
			}

			alias := tc.alias

			if tc.alias != "" {
				resp.Value("alias").String().IsEqual(alias)
			} else {
				resp.Value("alias").String().NotEmpty()

				alias = resp.Value("alias").String().Raw()
			}

			// Redirect
			testRedirect(t, alias, tc.url)

			// Remove
			e.DELETE("/"+path.Join("url", alias)).
				WithBasicAuth("user", "userpass").
				Expect().
				Status(http.StatusOK)

			// redirect again
			testRedirectNotFound(t, alias)
		})

	}
}

func testRedirect(t *testing.T, alias, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}
	redirectedToUrl, err := api.GetRedirect(u.String())
	require.NoError(t, err)

	require.Equal(t, urlToRedirect, redirectedToUrl)
}

func testRedirectNotFound(t *testing.T, alias string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}
	_, err := api.GetRedirect(u.String())

	require.Equal(t, err, fmt.Errorf("%s: %w: %d", api.GetRedirectOp, api.ErrInvalidStatusCode, http.StatusOK))
}
