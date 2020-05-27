// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// This package was created by imitating https://github.com/DataDog/dd-trace-go/tree/v1/contrib/go-redis/redis.

// Package redis provides tracing functions for tracing the go-redis/redis package (https://github.com/go-redis/redis).
// This package supports go-redis/redis v7.
package redis

import (
	"context"
	"math"
	"net"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// Client is used to trace requests to a redis server.
type Client struct {
	*redis.Client
	*params
}

var _ redis.Cmdable = (*Client)(nil)

// params holds the tracer and a set of parameters which are recorded with every trace.
type params struct {
	host   string
	port   string
	db     string
	config *clientConfig
}

// NewClient returns a new Client that is traced with the default tracer under
// the service name "redis".
func NewClient(opt *redis.Options, opts ...ClientOption) *Client {
	return WrapClient(redis.NewClient(opt), opts...)
}

// WrapClient wraps a given redis.Client with a tracer under the given service name.
func WrapClient(c *redis.Client, opts ...ClientOption) *Client {
	cfg := new(clientConfig)
	defaults(cfg)
	for _, fn := range opts {
		fn(cfg)
	}
	opt := c.Options()
	host, port, err := net.SplitHostPort(opt.Addr)
	if err != nil {
		host = opt.Addr
		port = "6379"
	}
	params := &params{
		host:   host,
		port:   port,
		db:     strconv.Itoa(opt.DB),
		config: cfg,
	}
	tc := &Client{Client: c, params: params}
	tc.Client.AddHook(&hook{tc: tc})
	return tc
}

type hook struct {
	tc *Client
}

var _ redis.Hook = (*hook)(nil)

func (h *hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	raw := cmd.String()
	parts := strings.Split(raw, " ")
	length := len(parts) - 1
	p := h.tc.params
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeRedis),
		tracer.ServiceName(p.config.serviceName),
		tracer.ResourceName(parts[0]),
		tracer.Tag(ext.TargetHost, p.host),
		tracer.Tag(ext.TargetPort, p.port),
		tracer.Tag("out.db", p.db),
		tracer.Tag("redis.raw_command", raw),
		tracer.Tag("redis.args_length", strconv.Itoa(length)),
	}
	if !math.IsNaN(p.config.analyticsRate) {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, p.config.analyticsRate))
	}
	_, ctxWithSpan := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
	return ctxWithSpan, nil
}

func (h *hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span, _ := tracer.SpanFromContext(ctx)
	var finishOpts []ddtrace.FinishOption
	err := cmd.Err()
	if err != redis.Nil {
		finishOpts = append(finishOpts, tracer.WithError(err))
	}
	span.Finish(finishOpts...)
	return err
}

func (h *hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	raw := commandsToString(cmds)
	parts := strings.Split(raw, " ")
	length := len(parts) - 1
	p := h.tc.params
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeRedis),
		tracer.ServiceName(p.config.serviceName),
		tracer.ResourceName(parts[0]),
		tracer.Tag(ext.TargetHost, p.host),
		tracer.Tag(ext.TargetPort, p.port),
		tracer.Tag("out.db", p.db),
		tracer.Tag("redis.raw_command", raw),
		tracer.Tag("redis.args_length", strconv.Itoa(length)),
	}
	if !math.IsNaN(p.config.analyticsRate) {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, p.config.analyticsRate))
	}
	_, ctxWithSpan := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
	return ctxWithSpan, nil
}

func (h *hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
	span, _ := tracer.SpanFromContext(ctx)
	span.SetTag(ext.ResourceName, commandsToString(cmds))
	span.SetTag("redis.pipeline_length", strconv.Itoa(len(cmds)))
	span.Finish()
	return nil
}

// commandsToString returns a string representation of a slice of redis Commands, separated by newlines.
func commandsToString(cmds []redis.Cmder) string {
	var b strings.Builder
	for _, cmd := range cmds {
		b.WriteString(cmd.String())
		b.WriteString("\n")
	}
	return b.String()
}
