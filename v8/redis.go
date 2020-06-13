// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016-2020 Datadog, Inc.

// This package was created by imitating https://github.com/DataDog/dd-trace-go/tree/v1/contrib/go-redis/redis.

// Package redis provides tracing functions for tracing the go-redis/redis package (https://github.com/go-redis/redis).
// This package supports go-redis/redis v8.
package redis

import (
	"context"
	"math"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v8"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/ext"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

// NewClient returns a new Client that is traced with the default tracer under
// the service name "redis".
func NewClient(opt *redis.Options, opts ...ClientOption) *redis.Client {
	return WrapClient(redis.NewClient(opt), opts...)
}

// WrapClient wraps a given redis.Client with a tracer under the given service name.
func WrapClient(c *redis.Client, opts ...ClientOption) *redis.Client {
	_opts := []ClientOption{withRedisOptions(c.Options())}
	_opts = append(_opts, opts...)
	c.AddHook(NewHook(_opts...))
	return c
}

type Hook struct {
	cfg *clientConfig
}

func NewHook(opts ...ClientOption) *Hook {
	cfg := new(clientConfig)
	defaults(cfg)
	for _, opt := range opts {
		opt(cfg)
	}
	return &Hook{
		cfg: cfg,
	}
}

var _ redis.Hook = (*Hook)(nil)

func (h *Hook) BeforeProcess(ctx context.Context, cmd redis.Cmder) (context.Context, error) {
	raw := cmd.String()
	parts := strings.Split(raw, " ")
	length := len(parts) - 1
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeRedis),
		tracer.ServiceName(h.cfg.serviceName),
		tracer.ResourceName(parts[0]),
		tracer.Tag(ext.TargetHost, h.cfg.host),
		tracer.Tag(ext.TargetPort, h.cfg.port),
		tracer.Tag("out.db", h.cfg.db),
		tracer.Tag("redis.raw_command", raw),
		tracer.Tag("redis.args_length", strconv.Itoa(length)),
	}
	if !math.IsNaN(h.cfg.analyticsRate) {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, h.cfg.analyticsRate))
	}
	_, ctxWithSpan := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
	return ctxWithSpan, nil
}

func (h *Hook) AfterProcess(ctx context.Context, cmd redis.Cmder) error {
	span, _ := tracer.SpanFromContext(ctx)
	var finishOpts []ddtrace.FinishOption
	err := cmd.Err()
	if err != redis.Nil {
		finishOpts = append(finishOpts, tracer.WithError(err))
	}
	span.Finish(finishOpts...)
	return err
}

func (h *Hook) BeforeProcessPipeline(ctx context.Context, cmds []redis.Cmder) (context.Context, error) {
	raw := commandsToString(cmds)
	parts := strings.Split(raw, " ")
	length := len(parts) - 1
	opts := []ddtrace.StartSpanOption{
		tracer.SpanType(ext.SpanTypeRedis),
		tracer.ServiceName(h.cfg.serviceName),
		tracer.ResourceName(parts[0]),
		tracer.Tag(ext.TargetHost, h.cfg.host),
		tracer.Tag(ext.TargetPort, h.cfg.port),
		tracer.Tag("out.db", h.cfg.db),
		tracer.Tag("redis.raw_command", raw),
		tracer.Tag("redis.args_length", strconv.Itoa(length)),
	}
	if !math.IsNaN(h.cfg.analyticsRate) {
		opts = append(opts, tracer.Tag(ext.EventSampleRate, h.cfg.analyticsRate))
	}
	_, ctxWithSpan := tracer.StartSpanFromContext(ctx, "redis.command", opts...)
	return ctxWithSpan, nil
}

func (h *Hook) AfterProcessPipeline(ctx context.Context, cmds []redis.Cmder) error {
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
