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

And to run integration tests (non deterministic with dependencies
on services on the internet):

```
make check-integration
```

# Coverage

To generate the coverage report:

```
make coverage
```

To generate and view it in your browser:

```
make coverage-view
```


# Static Analysis

To perform static analysis run:

```
make static-analysis
```
