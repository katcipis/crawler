package crawler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestCrawler(t *testing.T) {
	server, url := setupServer(t, "./testdata/fakesite")
	defer server.Close()

	fmt.Println("created test server", url)
}

func setupServer(t *testing.T, dir string) (*httptest.Server, url.URL) {
	handler := http.FileServer(http.Dir(dir))
	server := httptest.NewServer(handler)
	url, err := url.Parse(server.URL)

	fatalerr(t, err, "setting up server")
	return server, *url
}

func fatalerr(t *testing.T, err error, op string) {
	t.Helper()

	if err != nil {
		t.Fatalf("error[%s] while %s", err, op)
	}
}
