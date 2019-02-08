package crawler_test

import (
	"context"
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
	entrypoint, teardown := setupFileServer(t, "./testdata/fakesite")
	defer teardown()

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
		result(entrypoint, "/dir/page1.html", ""),
	}

	const maxConcurrency = 10
	const wantCrawlingErrs = 3
	const timeout = time.Minute

	for concurrency := uint(1); concurrency <= maxConcurrency; concurrency++ {
		t.Run(fmt.Sprintf("Concurrency%d", concurrency), func(t *testing.T) {
			testCrawler(
				t,
				context.Background(),
				entrypoint,
				concurrency,
				timeout,
				copyResults(want),
				wantCrawlingErrs,
			)
		})
	}
}

func TestCrawlingUnreachableSite(t *testing.T) {
	const concurrency = 5
	const wantCrawlingErrs = 1
	const timeout = time.Minute

	testCrawler(
		t,
		context.Background(),
		url.URL{
			Scheme: "http",
			Host:   "unreachable.com.io.it",
		},
		concurrency,
		timeout,
		[]crawler.Result{},
		wantCrawlingErrs,
	)
}

func TestCrawlingEmptySite(t *testing.T) {
	entrypoint, teardown := setupFileServer(t, "./testdata/emptysite")
	defer teardown()

	const concurrency = 5
	const wantCrawlingErrs = 0
	const timeout = time.Minute

	testCrawler(
		t,
		context.Background(),
		entrypoint,
		concurrency,
		timeout,
		[]crawler.Result{},
		wantCrawlingErrs,
	)
}

func TestCrawlingRespectsPerRequestTimeout(t *testing.T) {
	entrypoint, teardown := setupHangingServer(t, time.Hour)
	defer teardown()

	const concurrency = 5
	const wantCrawlingErrs = 1
	const timeout = time.Millisecond

	testCrawler(
		t,
		context.Background(),
		entrypoint,
		concurrency,
		timeout,
		[]crawler.Result{},
		wantCrawlingErrs,
	)
}

func TestCrawlingRespectsCancellation(t *testing.T) {
	entrypoint, teardown := setupHangingServer(t, time.Hour)
	defer teardown()

	const concurrency = 5
	const wantCrawlingErrs = 1
	const timeout = time.Hour

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	defer cancel()

	testCrawler(
		t,
		ctx,
		entrypoint,
		concurrency,
		timeout,
		[]crawler.Result{},
		wantCrawlingErrs,
	)
}

func TestCrawlerFailsToStartIfConcurrencyIsZero(t *testing.T) {
	entrypoint, teardown := setupFileServer(t, "./testdata/emptysite")
	defer teardown()

	res, errs := crawler.Start(context.Background(), entrypoint, 0, time.Minute)

	err := <-errs
	if err == nil {
		t.Fatal("expected error")
	}

	_, okRes := <-res
	_, okErrs := <-errs

	if okRes {
		t.Error("expected results channel to be closed")
	}

	if okErrs {
		t.Error("expected errs channel to be closed")
	}

}

func testCrawler(
	t *testing.T,
	ctx context.Context,
	entrypoint url.URL,
	concurrency uint,
	timeout time.Duration,
	want []crawler.Result,
	wantErrs uint,
) {
	t.Helper()

	results, errs := crawler.Start(ctx, entrypoint, concurrency, timeout)

	drainedErrs := make(chan struct{})
	errsCount := uint(0)

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

	<-drainedErrs

	if len(want) > 0 {
		t.Errorf("missing wanted results: %+v", want)
	}

	if wantErrs != errsCount {
		t.Fatalf("expected [%d] errors but got [%d] instead", wantErrs, errsCount)
	}
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

func newServer(t *testing.T, h http.Handler) (url.URL, func()) {
	server := httptest.NewServer(h)
	url, err := url.Parse(server.URL)
	fatalerr(t, err, "setting up server")
	return *url, server.Close
}

func setupHangingServer(t *testing.T, hangtime time.Duration) (url.URL, func()) {
	ctx, cancel := context.WithTimeout(context.Background(), hangtime)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-ctx.Done()
		w.WriteHeader(http.StatusOK)
	})
	entrypoint, teardown := newServer(t, handler)
	return entrypoint, func() {
		// WHY: hanging server Close will block while all pending requests are answered
		//      This way we wait for hangtime or for the test to cancel/close the server
		cancel()
		teardown()
	}
}

func setupFileServer(t *testing.T, dir string) (url.URL, func()) {
	handler := http.FileServer(http.Dir(dir))
	return newServer(t, handler)
}

func fatalerr(t *testing.T, err error, op string) {
	t.Helper()

	if err != nil {
		t.Fatalf("error[%s] while %s", err, op)
	}
}
