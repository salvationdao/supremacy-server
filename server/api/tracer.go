package api

import (
	"context"
	"fmt"
	"net/http"

	"github.com/ninja-syndicate/hub"
	"gopkg.in/DataDog/dd-trace-go.v1/ddtrace/tracer"
)

type HubTracer struct {
}

type TracerContext string

func (tc TracerContext) String() string {
	return fmt.Sprintf("contextkey_%s", string(tc))
}

func (ht *HubTracer) OnConnect(ctx context.Context, r *http.Request) context.Context {
	return ctx
}
func (ht *HubTracer) OnEventStart(ctx context.Context, operation string, commandName string, transactionID string) context.Context {
	span := tracer.StartSpan("hub_handler", tracer.ResourceName(commandName))
	return context.WithValue(ctx, TracerContext("span"), span)

}
func (ht *HubTracer) OnEventStop(ctx context.Context, l hub.Logger) {
	span := ctx.Value(TracerContext("span")).(tracer.Span)
	span.Finish()
}
