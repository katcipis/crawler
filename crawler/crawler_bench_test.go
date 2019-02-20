package crawler_test

import (
	"context"
	"fmt"
	"net/url"
	"testing"
	"time"

	"github.com/katcipis/crawler/crawler"
)

func BenchmarkCrawler(b *testing.B) {

	const timeout = time.Hour
	const reqTimeout = 10 * time.Second
	const maxConcurrency = 30
	const startConcurrency = 1
	const step = 2
	const site = "https://monzo.com"

	entrypoint, err := url.Parse(site)
	if err != nil {
		b.Fatal(err)
	}

	for concurrency := uint(startConcurrency); concurrency <= maxConcurrency; concurrency += step {

		b.Run(fmt.Sprintf("Concurrency%d", concurrency), func(b *testing.B) {

			for i := 0; i < b.N; i++ {
				ctx, cancel := context.WithTimeout(
					context.Background(),
					timeout,
				)
				defer cancel()

				results, errs := crawler.Start(
					ctx,
					*entrypoint,
					concurrency,
					reqTimeout,
				)

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
		})
	}
}
