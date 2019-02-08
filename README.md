# crawler

This is a simple single domain crawler.
Given a URL it will crawl all pages within the domain name of the URL
and send a [sitemap](https://www.sitemaps.org/protocol.html)
in text format to the crawler stdout.


# Dependencies

Install:

* [Make](https://www.gnu.org/software/make/)
* [Docker](https://www.docker.com/)

And you should be able to run any of the commands documented here.

If you want to build and run tests directly in your host you
must install [Go](https://golang.org/) >= 1.11.


# Usage

After building the crawler just run:

```
./cmd/crawler -url <url>
```


For more options run:

```
./cmd/crawler -help
```

# Build

Run:

```
make
```

And you should be able to run the crawler:

```
./cmd/crawler
```

# Testing

To run the tests:

```
make check
```

# Coverage

To generate and view the coverage report:

```
make coverage
```

# Static Analysis

To perform static analysis run:

```
make static-analysis
```
