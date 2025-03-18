// SPDX-License-Identifier: MIT

package tracer

import (
	"context"
	"errors"
	"fmt"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	grpccodes "google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Span interface {
	Tag(key string, value any)

	// IsValid returns if the SpanContext is valid. A valid span context has a valid TraceID and SpanID.
	IsValid() bool

	// IsSampled returns if the sampling bit is set in the SpanContext's TraceFlags.
	IsSampled() bool

	// SpanId returns the SpanID from the SpanContext as string.
	SpanId() string

	// TraceId returns the TraceID from the SpanContext as string.
	TraceId() string

	// SetStatus sets the status of the Span in the form of a code and a
	// description, provided the status hasn't already been set to a higher
	// value before (OK > Error > Unset). The description is only included in a
	// status when the code is for an error.
	SetStatus(code codes.Code, description string)

	AddEvent(name string, opts ...trace.EventOption)

	// RecordError will record err as an exception span event for this span. An
	// additional call to SetStatus is required if the Status of the Span should
	// be set to Error, as this method does not change the Span status. If this
	// span is not being recorded or err is nil then this method does nothing.
	RecordError(err error)

	// End completes the Span. The Span is considered complete and ready
	// to be delivered through the rest of the telemetry pipeline after
	// this method is called. Therefore, updates to the Span are not allowed
	// after this method has been called.
	//
	// Before the Span completion, End handles the specified errors. Sets the status
	// with codes.Error if any error is not nil. Except the context.Canceled or
	// gRPC status grpccodes.Canceled: in such case the "canceled" Event will be added.
	//
	// Arguments are pointers in order to allow at the beginning of an operation make
	// defer call with empty error that will be changed later:
	//
	//  func foo(ctx context.Context) (err error) {
	//  	span:= spanFromContext(ctx)
	//  	defer span.End(&err)
	//
	//  	// ...
	//  }
	End(errs ...*error)
}

type span struct {
	s trace.Span
}

var _ Span = (*span)(nil)

func (s *span) Tag(key string, value any) {
	switch v := value.(type) {
	case int:
		s.s.SetAttributes(attribute.Int(key, v))
	case string:
		s.s.SetAttributes(attribute.String(key, v))
	case float64:
		s.s.SetAttributes(attribute.Float64(key, v))
	case int64:
		s.s.SetAttributes(attribute.Int64(key, v))
	case bool:
		s.s.SetAttributes(attribute.Bool(key, v))
	case []string:
		s.s.SetAttributes(attribute.StringSlice(key, v))
	case []int:
		s.s.SetAttributes(attribute.IntSlice(key, v))
	case fmt.Stringer:
		s.s.SetAttributes(attribute.Stringer(key, v))
	}
}

func (s *span) IsValid() bool {
	return s.s.SpanContext().IsValid()
}

func (s *span) IsSampled() bool {
	return s.s.SpanContext().IsSampled()
}

func (s *span) SpanId() string {
	return s.s.SpanContext().SpanID().String()
}

func (s *span) TraceId() string {
	return s.s.SpanContext().TraceID().String()
}

func (s *span) SetStatus(code codes.Code, description string) {
	s.s.SetStatus(code, description)
}

func (s *span) AddEvent(name string, opts ...trace.EventOption) {
	s.s.AddEvent(name, opts...)
}

func (s *span) RecordError(err error) {
	s.s.RecordError(err)
}

func (s *span) End(errs ...*error) {
	for _, err := range errs {
		if err != nil && (*err) != nil {
			s.handleError(*err)
			break
		}
	}
	s.s.End()
}

func (s *span) handleError(err error) {
	if errors.Is(err, context.Canceled) || status.Code(err) == grpccodes.Canceled {
		s.s.AddEvent("canceled", trace.WithTimestamp(time.Now()))
	} else {
		s.s.SetStatus(codes.Error, err.Error())
	}
}

func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, *span) {
	span := new(span)
	if tracer == nil {
		ctx, span.s = noop.NewTracerProvider().Tracer("noop").Start(ctx, name, opts...)
	} else {
		ctx, span.s = tracer.Start(ctx, name, opts...)
	}

	return ctx, span
}

func SpanFromContext(ctx context.Context) *span {
	span := new(span)
	span.s = trace.SpanFromContext(ctx)

	return span
}
