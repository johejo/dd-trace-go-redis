// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

package redis

import (
	"math"
	"net"
	"strconv"

	"github.com/go-redis/redis/v8"
)

const (
	defaultHost = "127.0.0.1"
	defaultPort = "6379"
	defaultDB   = "0"
)

type clientConfig struct {
	serviceName   string
	analyticsRate float64
	host          string
	port          string
	db            string
}

// ClientOption represents an option that can be used to create or wrap a client.
type ClientOption func(*clientConfig)

func defaults(cfg *clientConfig) {
	cfg.serviceName = "redis.client"
	// cfg.analyticsRate = globalconfig.AnalyticsRate()
	cfg.analyticsRate = math.NaN()
	cfg.host = defaultHost
	cfg.port = defaultPort
	cfg.db = defaultDB
}

// WithServiceName sets the given service name for the client.
func WithServiceName(name string) ClientOption {
	return func(cfg *clientConfig) {
		cfg.serviceName = name
	}
}

// WithAnalytics enables Trace Analytics for all started spans.
func WithAnalytics(on bool) ClientOption {
	return func(cfg *clientConfig) {
		if on {
			cfg.analyticsRate = 1.0
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithAnalyticsRate sets the sampling rate for Trace Analytics events
// correlated to started spans.
func WithAnalyticsRate(rate float64) ClientOption {
	return func(cfg *clientConfig) {
		if rate >= 0.0 && rate <= 1.0 {
			cfg.analyticsRate = rate
		} else {
			cfg.analyticsRate = math.NaN()
		}
	}
}

// WithHost sets the host for the client.
func WithHost(host string) ClientOption {
	return func(cfg *clientConfig) {
		cfg.host = host
	}
}

// WithPort sets the port for the client.
func WithPort(port string) ClientOption {
	return func(cfg *clientConfig) {
		cfg.port = port
	}
}

// WithDB sets the db for the client.
func WithDB(db string) ClientOption {
	return func(cfg *clientConfig) {
		cfg.db = db
	}
}

// WithRedisOptions sets the redis.Option for the client.
func WithRedisOptions(opts *redis.Options) ClientOption {
	return func(cfg *clientConfig) {
		host, port, err := net.SplitHostPort(opts.Addr)
		if err != nil {
			host = defaultHost
			port = defaultPort
		}
		cfg.host = host
		cfg.port = port
		cfg.db = strconv.Itoa(opts.DB)
	}
}
