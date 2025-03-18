// SPDX-License-Identifier: MIT

package tracer

import (
	"fmt"
	"time"
)

var defaultOptions = []Option{
	WithCollectorHost("localhost"),
	WithCollectorPort(4317), //nolint:mnd
}

type Option func(opts *Options)

// Noop disables tracer (for tests and local development).
func Noop() Option {
	return func(o *Options) {
		o.noop = true
	}
}

func WithCollectorHost(host string) Option {
	return func(opts *Options) {
		opts.host = host
	}
}

func WithCollectorPort(port uint16) Option {
	return func(opts *Options) {
		opts.port = port
	}
}

func WithKeepaliveTime(val time.Duration) Option {
	return func(opts *Options) {
		opts.keepaliveTime = &val
	}
}

func WithKeepaliveTimeout(val time.Duration) Option {
	return func(opts *Options) {
		opts.keepaliveTimeout = &val
	}
}

func WithKeepalivePermitWithoutStream(val bool) Option {
	return func(opts *Options) {
		opts.keepalivePermitWithoutStream = &val
	}
}

type Options struct {
	keepaliveTime                *time.Duration
	keepaliveTimeout             *time.Duration
	keepalivePermitWithoutStream *bool

	host string
	port uint16

	noop bool
}

func buildOptions(opts []Option) Options {
	options := Options{}

	opts = append(defaultOptions, opts...)
	for _, opt := range opts {
		opt(&options)
	}

	return options
}

func (o Options) GetGrpcTarget() string {
	return fmt.Sprintf("%s:%d", o.host, o.port)
}

func (o Options) IsNoop() bool {
	return o.noop
}
