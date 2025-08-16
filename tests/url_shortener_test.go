package tests

import (
	"net/http"
	"net/url"
	"testing"

	"github.com/Matvey-Makaro/url-shortener/internal/config"
	"github.com/Matvey-Makaro/url-shortener/internal/http-server/handlers/url/delete"
	"github.com/Matvey-Makaro/url-shortener/internal/http-server/handlers/url/save"
	"github.com/Matvey-Makaro/url-shortener/internal/lib/api"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/require"
)

func TestURLShortener(t *testing.T) {
	testCases := []struct {
		name  string
		url   string
		alias string
		err   string
	}{
		{
			name:  "ValidURL",
			url:   gofakeit.URL(),
			alias: gofakeit.Word() + gofakeit.Word(),
		},
	}

	cfg := config.MustLoad()
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			u := url.URL{
				Scheme: "http",
				Host:   cfg.Address,
			}

			e := httpexpect.Default(t, u.String())

			// Save
			resp := e.POST("/url").
				WithJSON(save.Request{
					URL:   tc.url,
					Alias: tc.alias,
				}).
				WithBasicAuth(cfg.User, cfg.Password).
				Expect().Status(http.StatusOK).
				JSON().Object()

			if tc.err != "" {
				resp.NotContainsKey("alias")
				resp.Value("Error").String().IsEqual(tc.err)
			}

			alias := tc.alias

			if alias != "" {
				resp.Value("alias").String().IsEqual(alias)
			}

			// Read
			testRedirect(t, cfg.Address, alias, tc.url)

			// Delete
			e.DELETE("/url").
				WithJSON(delete.Request{
					Alias: tc.alias,
				}).
				WithBasicAuth(cfg.User, cfg.Password).
				Expect().Status(http.StatusOK)
		})
	}
}

func testRedirect(t *testing.T, host, alias, urlToRedirect string) {
	u := url.URL{
		Scheme: "http",
		Host:   host,
		Path:   alias,
	}

	redirectedUrl, err := api.GetRedirect(u.String())
	require.NoError(t, err)
	require.Equal(t, urlToRedirect, redirectedUrl)
}
