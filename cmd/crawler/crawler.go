package main

import (
	"context"
	"flag"
	"fmt"
	"net/url"
	"os"
	"time"

	"github.com/katcipis/crawler/crawler"
)

func main() {
	const defaultConcurrency = 4
	const defaultFormatter = "sitemap"
	const defaultRequestTimeout = time.Minute
	const defaultTimeout = 0

	var concurrency uint
	var url string
	var format string
	var timeout time.Duration
	var reqTimeout time.Duration

	flag.UintVar(
		&concurrency,
		"concurrency",
		defaultConcurrency,
		"amount of concurrent crawlers",
	)
	flag.DurationVar(
		&timeout,
		"timeout",
		defaultTimeout,
		"timeout of the entire crawling, 0 if you want it to run until all links are reached",
	)
	flag.DurationVar(
		&reqTimeout,
		"request-timeout",
		defaultRequestTimeout,
		"timeout to be used on each request made",
	)
	flag.StringVar(
		&url,
		"url",
		"",
		"url that will be the entry point of the crawler (obligatory)",
	)
	flag.StringVar(
		&format,
		"format",
		"",
		fmt.Sprintf("format of the output, available formats: %s", availableFormats()),
	)

	flag.Parse()

	if url == "" {
		fmt.Fprint(os.Stderr, "\nurl is an obligatory parameter\n\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	err := startCrawler(url, concurrency, timeout, reqTimeout, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\ncrawling failed:%s", err)
		os.Exit(1)
	}
}

func startCrawler(
	ep string,
	concurrency uint,
	timeout time.Duration,
	reqTimeout time.Duration,
	format string,
) error {
	entrypoint, err := url.Parse(ep)
	if err != nil {
		return fmt.Errorf("error[%s] parsing entrypoint URL[%s]", err, ep)
	}

	formatter, err := getFormatter(format)

	var ctx context.Context
	var cancel context.CancelFunc

	if timeout > 0 {
		ctx, cancel = context.WithTimeout(context.Background(), timeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	res, errs := crawler.Start(ctx, *entrypoint, concurrency, reqTimeout)

	go drainErrors(errs)

	return formatter(res, os.Stdout)
}

func drainErrors(errs <-chan error) {
	for err := range errs {
		fmt.Fprintln(os.Stderr, err)
	}
}

func getFormatter(name string) (crawler.Formatter, error) {
	return crawler.FormatAsTextSitemap, nil
}

func availableFormats() []string {
	return []string{"sitemap"}
}
