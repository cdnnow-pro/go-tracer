package tracer

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

func makeGrpcExporter(ctx context.Context, options Options) (*otlptrace.Exporter, func() error, error) {
	conn, err := grpc.NewClient(options.GetGrpcTarget(), grpcDialOptions(options)...)
	if err != nil {
		return nil, nil, fmt.Errorf("trace collector connection error: %w", err)
	}
	conn.Connect()

	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	return exporter, func() error {
		if err := conn.Close(); err != nil {
			return fmt.Errorf("failed to close tracer connection: %w", err)
		}
		return nil
	}, nil
}

func grpcDialOptions(options Options) []grpc.DialOption {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	}

	var (
		useKeepalive              bool
		keepaliveClientParameters keepalive.ClientParameters
	)

	if options.keepaliveTime != nil {
		useKeepalive = true
		keepaliveClientParameters.Time = *options.keepaliveTime
	}
	if options.keepaliveTimeout != nil {
		useKeepalive = true
		keepaliveClientParameters.Timeout = *options.keepaliveTimeout
	}
	if options.keepalivePermitWithoutStream != nil {
		useKeepalive = true
		keepaliveClientParameters.PermitWithoutStream = *options.keepalivePermitWithoutStream
	}

	if useKeepalive {
		opts = append(opts, grpc.WithKeepaliveParams(keepaliveClientParameters))
	}

	return opts
}
