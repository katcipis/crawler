package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func main() {
	const defaultConcurrency = 4
	const defaultFormatter = "sitemap"
	const defaultTimeout = time.Minute

	var concurrency uint
	var url string
	var format string
	var timeout time.Duration

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

	err := startCrawler(url, concurrency, timeout, format)
	if err != nil {
		fmt.Fprintf(os.Stderr, "\ncrawling failed:%s", err)
		os.Exit(1)
	}
}

func startCrawler(url string, concurrency uint, timeout time.Duration, format string) error {
	return nil
}

func availableFormats() []string {
	return []string{"sitemap"}
}
