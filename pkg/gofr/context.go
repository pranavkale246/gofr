package gofr

import (
	"context"

	"gofr.dev/pkg/gofr/container"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

type Context struct {
	context.Context

	// Request needs to be public because handlers need to access request details. Else, we need to provide all
	// functionalities of the Request as a method on context. This is not needed because Request here is an interface
	// So, internals are not exposed anyway.
	Request

	// Same logic as above.
	*container.Container

	// responder is private as Handlers do not need to worry about how to respond. But it is still an abstraction over
	// normal response writer as we want to keep the context independent of http. Will help us in writing CMD application
	// or grpc servers etc using the same handler signature.
	responder Responder
}

/*
Trace returns an open telemetry span. We have to always close the span after corresponding work is done. Usages:

	span := c.Trace("Some Work")
	// Do some work here.
	defer span.End()

If an entire function has to traced as span, we can use a simpler format:

	defer c.Trace("ExampleHandler").End()

We can write this at the start of function and because of how defer works, trace will start at that line
but End will be called after function ends.

Developer Note: If you chain methods in a defer statement, everything except the last function will be evaluated at call time.
*/
func (c *Context) Trace(name string) trace.Span {
	tr := otel.GetTracerProvider().Tracer("gofr-context")
	_, span := tr.Start(c.Context, name)
	// c.Context = context // TODO - wanted to change the internal context to the newly returned context. But it crashes.
	return span
}

func (c *Context) Bind(i interface{}) error {
	return c.Request.Bind(i)
}

// func (c *Context) reset(w Responder, r Request) {
//	c.Request = r
//	c.responder = w
//	c.Context = nil
//	// c.Logger = nil // For now, all loggers are same. So, no need to set nil.
// }

func newContext(w Responder, r Request, newContainer *container.Container) *Context {
	return &Context{
		Context:   r.Context(),
		Request:   r,
		responder: w,
		Container: newContainer,
	}
}
