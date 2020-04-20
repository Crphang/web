# Introduction

Gets R packages metadata in a sqlite file. Get started with:

```sh
go get ./... # Install dependecies
go run scrape.go
```

# API

API serves on search endpoint.

```sh
go run main.go

curl -X GET \
  'http://localhost:8080/search?name=abc' \
  -H 'cache-control: no-cache' \
  -H 'postman-token: d72b3cf6-bd1d-21dd-9212-2eb2dac34882'

```

# Enhancement

1. Faster scraping and processing and seeding.
1. Better database abstraction that does not have direct dependency with sqlite.
1. Better server abstraction with router
1. Better data transfer format, possibly JSON.
