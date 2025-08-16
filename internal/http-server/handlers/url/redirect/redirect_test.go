package redirect_test

import (
	"net/http/httptest"
	"testing"

	"github.com/Matvey-Makaro/url-shortener/internal/http-server/handlers/url/redirect"
	"github.com/Matvey-Makaro/url-shortener/internal/lib/api"
	"github.com/Matvey-Makaro/url-shortener/internal/lib/logger/handlers/slogdiscard"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/require"
)

func TestRedirectHandler(t *testing.T) {
	tests := []struct {
		name      string
		alias     string
		url       string
		respError string
		mockError error
	}{
		{
			name:  "Success",
			alias: "test",
			url:   "https://www.google.com/",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			urlGetter := redirect.NewMockURLGetter(t)

			if tc.respError == "" || tc.mockError != nil {
				urlGetter.On("GetURL", tc.alias).Return(tc.url, tc.mockError).Once()
			}

			r := chi.NewRouter()
			r.Get("/{alias}", redirect.New(slogdiscard.NewDiscardLogger(), urlGetter))

			ts := httptest.NewServer(r)
			defer ts.Close()

			redirectedToURL, err := api.GetRedirect(ts.URL + "/" + tc.alias)
			require.NoError(t, err)
			require.Equal(t, tc.url, redirectedToURL)
		})
	}
}
