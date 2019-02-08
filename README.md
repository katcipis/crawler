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

# Installation

If you have Go >= 1.11 installed you can run:

```
go get github.com/katcipis/crawler/cmd/crawler
```

# Basic Usage

After building the crawler just run:

```
./cmd/crawler/crawler -url <url>
```


For more options run:

```
./cmd/crawler/crawler -help
```

There is a make target that makes it easy to visualize the sitemap
as a graph. To use it just run:

```
make graph url=<entrypoint>
```

And it will generate a PNG file with the graphical representation
of the sitemap and open it using the default application in your
system to display images.


# Sitemap Formatters

There are multiple formats to represent a sitemap. The default
sitemap specification does not show the relation between links.

To make that easier to check there is two extra outputs besides
the default textual sitemap:

* linked
* graphviz

The **linked** formatter is pretty much like a sitemap with the
exception that is shows pairs of URLs showing from where a URL
has been reached.

The **graphviz** formatter will produce an graphviz file with
the full graph of the site which can be used to generate
a graphical representation of the sitemap.

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
