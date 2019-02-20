package crawler_test

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/katcipis/crawler/crawler"
)

func BenchmarkCrawler(b *testing.B) {

	const timeout = 20 * time.Second
	const reqTimeout = 5 * time.Second
	const concurrency = 4
	const site = "https://monzo.com"

	entrypoint, err := url.Parse(site)
	if err != nil {
		b.Fatal(err)
	}

	for i := 0; i < b.N; i++ {
		ctx, cancel := context.WithTimeout(context.Background(), timeout)
		defer cancel()

		results, errs := crawler.Start(ctx, *entrypoint, concurrency, reqTimeout)

		go func() {
			for range errs {
			}
		}()

		linksCount := 0

		for range results {
			linksCount += 1
		}

		if linksCount == 0 {
			b.Fatalf("expected at least one link from [%s]", site)
		}
	}
}
