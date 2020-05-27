# dd-trace-go-redis

DataDog tracer for go-redis/redis v7 and v8

## Motivation

[dd-trace-go](https://github.com/DataDog/dd-trace-go) does not support [go-redis/redis](https://github.com/go-redis/redis) v7 and v8.

This package mimics [dd-trace-go/contrib/go-redis/redis](https://github.com/DataDog/dd-trace-go/tree/v1/contrib/go-redis/redis), but has been modified to support go-redis' Hook API.

## Thanks

Many codes were stolen from dd-trace-go.

## Install

```
go get github.com/johejo/dd-trace-go-redis
```

## Usage

See the dd-trace-go documentation for details.
https://godoc.org/gopkg.in/DataDog/dd-trace-go.v1/contrib/go-redis/redis

And replace your import (`import "gopkg.in/DataDog/dd-trace-go.v1/contrib/go-redis/redis"`) with:

for `go-redis/redis/v7`
```
"github.com/johejo/dd-trace-go-redis/v7"
```

for `go-redis/redis/v8`
```
"github.com/johejo/dd-trace-go-redis/v8"
```

## Difference from dd-trace-go

`WithContext` and `WithTimeout` will returns `go-redis/redis`'s raw `Client`.
But Hook is already set up so Tracer will work fine.
