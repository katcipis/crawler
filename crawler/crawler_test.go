package crawler_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/katcipis/crawler/crawler"
)

// TODO:
// 1 - entry point answers 500 ERROR

func TestCrawlingMultipleLinks(t *testing.T) {
	server, entrypoint := setupFileServer(t, "./testdata/fakesite")
	defer server.Close()

	want := []crawler.Result{
		result(entrypoint, "", "/info.html"),
		result(entrypoint, "", "/nesting/info.html"),
		result(entrypoint, "", "/dir"),
		result(entrypoint, "", "/wontExist.html"),
		result(entrypoint, "", "/wont/exist/page.html"),
		result(entrypoint, "", "/wont/exist2"),
		result(entrypoint, "/info.html", "/cycle.html"),
		result(entrypoint, "/info.html", "/final.html"),
		result(entrypoint, "/cycle.html", "/info.html"),
		result(entrypoint, "/cycle.html", "/final.html"),
		result(entrypoint, "/nesting/info.html", "/cycle.html"),
		result(entrypoint, "/nesting/info.html", "/final.html"),
		result(entrypoint, "/dir", "/dir/page1.html"),
		result(entrypoint, "/dir", "/dir/page2.html"),
		result(entrypoint, "/dir", "/dir/page3.txt"),
	}

	const maxConcurrency uint = 10

	for concurrency := uint(1); concurrency <= maxConcurrency; concurrency++ {
		t.Run(fmt.Sprintf("Concurrency%d", concurrency), func(t *testing.T) {
			testCrawler(t, entrypoint, concurrency, copyResults(want))
		})
	}
}

func TestCrawlingEmptySite(t *testing.T) {
	server, entrypoint := setupFileServer(t, "./testdata/emptysite")
	defer server.Close()

	const concurrency = 5

	testCrawler(t, entrypoint, concurrency, []crawler.Result{})
}

func TestCrawlerFailsToStartIfConcurrencyIsZero(t *testing.T) {
	server, entrypoint := setupFileServer(t, "./testdata/emptysite")
	defer server.Close()

	_, err := crawler.Start(entrypoint, 0, time.Minute)
	if err == nil {
		t.Fatal("expected error")
	}
}

func testCrawler(
	t *testing.T,
	entrypoint url.URL,
	concurrency uint,
	want []crawler.Result,
) {
	t.Helper()

	timeout := time.Minute
	results, errs := crawler.Start(entrypoint, concurrency, timeout)

	drainedErrs := make(chan struct{})
	errsCount := 0

	go func() {
		for range errs {
			errsCount += 1
		}
		close(drainedErrs)
	}()

	seen := map[string]bool{}
	for got := range results {
		if seen[got.String()] {
			t.Fatalf("duplicated result[%s]", got)
		}
		seen[got.String()] = true
		want = removeResult(t, want, got)
	}

	if len(want) > 0 {
		t.Fatalf("missing wanted results: %+v", want)
	}

	<-drainedErrs
	// TODO: check errsCount
}

func removeResult(t *testing.T, want []crawler.Result, got crawler.Result) []crawler.Result {
	for i, w := range want {
		if w == got {
			return append(want[:i], want[i+1:]...)
		}
	}

	t.Fatalf("unable to find crawler result:\n\n%+v\n\nin wanted results:\n\n%+v", got, want)
	return want
}

func copyResults(res []crawler.Result) []crawler.Result {
	copied := make([]crawler.Result, len(res))
	copy(copied, res)
	return copied
}

func result(entrypoint url.URL, parent string, link string) crawler.Result {
	return crawler.Result{
		Parent: url.URL{
			Scheme: entrypoint.Scheme,
			Host:   entrypoint.Host,
			Path:   parent,
		},
		Link: url.URL{
			Scheme: entrypoint.Scheme,
			Host:   entrypoint.Host,
			Path:   link,
		},
	}
}

func setupFileServer(t *testing.T, dir string) (*httptest.Server, url.URL) {
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
