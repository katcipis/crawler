package main

import (
	"flag"
)

func main() {
	const defaultConcurrency = 4

	var concurrency uint

	flag.UintVar(&concurrency, "concurrency", defaultConcurrency, "amount of concurrent crawlers")
	flag.Parse()
}
