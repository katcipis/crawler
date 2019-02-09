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

var formatters map[string]crawler.Formatter = map[string]crawler.Formatter{
	"text":     crawler.FormatAsTextSitemap,
	"graphviz": crawler.FormatAsGraphvizSitemap,
}

func main() {
	const defaultConcurrency = 10
	const defaultFormat = "text"
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
		defaultFormat,
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
		fmt.Fprintf(os.Stderr, "\ncrawling failed:%s\n", err)
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

	if entrypoint.Scheme == "" {
		entrypoint.Scheme = "http"
		entrypoint.Host = entrypoint.Path
		entrypoint.Path = ""
	}

	formatter, err := getFormatter(format)
	if err != nil {
		return err
	}

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
	formatter, ok := formatters[name]
	if !ok {
		return nil, fmt.Errorf("unknown formatter:[%s]", name)
	}
	return formatter, nil
}

func availableFormats() []string {
	fmts := []string{}
	for f := range formatters {
		fmts = append(fmts, f)
	}
	return fmts
}
