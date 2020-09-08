package main

import (
	"context"

	"github.com/jasonsoft/learning-temporal/ctxpropagation"
	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/workflow"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "temporal-client",
			Tags: []label.KeyValue{
				label.String("exporter", "jaeger"),
				label.Float64("float", 312.23),
			},
		}),
		jaeger.WithSDK(&sdktrace.Config{DefaultSampler: sdktrace.AlwaysSample()}),
	)
	if err != nil {
		panic(err)
	}

	return func() {
		flush()
	}
}

func main() {
	defer log.Flush()
	// set up log target
	log.
		Str("app_id", "starter").
		SaveToDefault()

	clog := console.New()
	log.AddHandler(clog, log.AllLevels...)

	ctx := context.Background()

	// create tracer
	fn := initTracer()
	defer fn()
	tr := global.Tracer("")
	bridgeTracer, _ := opentracing.NewTracerPair(tr)

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		ContextPropagators: []workflow.ContextPropagator{ctxpropagation.NewContextPropagator()},
		Tracer:             bridgeTracer,
	})
	if err != nil {
		log.Fatalf("Unable to create client", err)
	}
	defer c.Close()

	requestID := "abcd"

	ctx = context.WithValue(ctx, ctxpropagation.PropagateKey, &ctxpropagation.Values{Key: "request_id", Value: requestID})
	ctx = log.Str("request_id", requestID).WithContext(ctx)

	workflowOptions := client.StartWorkflowOptions{
		ID:        "hello_world_workflowID",
		TaskQueue: "hello-world",
	}

	//ctx, span := tr.Start(ctx, "workflow")
	we, err := c.ExecuteWorkflow(ctx, workflowOptions, "HelloWorldWorkflow", "Temporal")
	if err != nil {
		log.Fatalf("Unable to execute workflow", err)
	}
	//span.End()

	log.Infof("Started workflow", "WorkflowID", we.GetID(), "RunID", we.GetRunID())

	// Synchronously wait for the workflow completion.
	var result string
	err = we.Get(context.Background(), &result)
	if err != nil {
		log.Fatalf("Unable get workflow result", err)
	}
	log.Infof("Workflow result:", result)
}
