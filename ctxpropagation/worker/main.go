package main

import (
	"context"

	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jasonsoft/learning-temporal/ctxpropagation"
	"github.com/nite-coder/blackbear/pkg/log"
	"github.com/nite-coder/blackbear/pkg/log/handler/console"
	"go.opentelemetry.io/otel/api/global"
	"go.opentelemetry.io/otel/bridge/opentracing"
	"go.opentelemetry.io/otel/exporters/trace/jaeger"
	"go.opentelemetry.io/otel/label"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/internal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"

	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

// initTracer creates a new trace provider instance and registers it as global trace provider.
func initTracer() func() {
	// Create and install Jaeger export pipeline
	flush, err := jaeger.InstallNewPipeline(
		jaeger.WithCollectorEndpoint("http://localhost:14268/api/traces"),
		jaeger.WithProcess(jaeger.Process{
			ServiceName: "temporal-worker",
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
	logger := log.New()
	clog := console.New()
	logger.AddHandler(clog, log.AllLevels...)
	log.SetLogger(logger)

	// create tracer
	fn := initTracer()
	defer fn()
	tr := global.Tracer("")
	bridgeTracer, _ := opentracing.NewTracerPair(tr)

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		HostPort: "localhost:7233",
		ContextPropagators: []workflow.ContextPropagator{
			ctxpropagation.NewContextPropagator(),
		},
		Tracer: bridgeTracer,
	})
	if err != nil {
		log.Fatalf("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "hello-world", worker.Options{
		WorkflowInterceptorChainFactories: []internal.WorkflowInterceptor{},
	})

	w.RegisterWorkflow(HelloWorldWorkflow) // it only allow to use func instead of struct or pointer
	w.RegisterActivity(HelloWorldActivity)
	w.RegisterActivity(HelloWorldActivity2)

	go func() {
		err = w.Run(nil)
		if err != nil {
			log.Fatalf("Unable to start worker", err)
		}
	}()

	stopChan := make(chan os.Signal, 1)
	signal.Notify(stopChan, syscall.SIGINT, syscall.SIGKILL, syscall.SIGHUP, syscall.SIGTERM)
	<-stopChan
	log.Info("main: shutting down worker...")

	w.Stop()
	log.Info("main: worker was stopped")
}

// HelloWorldWorkflow is a Hello World workflow definition.
func HelloWorldWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	if val := ctx.Value(ctxpropagation.PropagateKey); val != nil {
		vals := val.(ctxpropagation.Values)
		log.Infof("custom context propagated to workflow, key: %s, val: %s", vals.Key, vals.Value)
	} else {
		log.Info("no value")
	}

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, "HelloWorldActivity", name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	err = workflow.ExecuteActivity(ctx, "HelloWorldActivity2", name).Get(ctx, &result)
	if err != nil {
		logger.Error("Activity failed.", "Error", err)
		return "", err
	}

	logger.Info("HelloWorld workflow completed.", "result", result)

	return result, nil
}

func HelloWorldActivity(ctx context.Context, name string) (string, error) {
	// tr := global.Tracer("")
	// ctx, span := tr.Start(ctx, "HelloWorldActivity")
	// defer span.End()

	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	if val := ctx.Value(ctxpropagation.PropagateKey); val != nil {
		vals := val.(ctxpropagation.Values)
		log.Infof("custom context propagated to activity, key: %s, val: %s", vals.Key, vals.Value)
	} else {
		log.Info("no value")
	}

	return "Hello " + name + "!", nil
}

func HelloWorldActivity2(ctx context.Context, name string) (string, error) {
	// tr := global.Tracer("")
	// ctx, span := tr.Start(ctx, "HelloWorldActivity")
	// defer span.End()

	logger := activity.GetLogger(ctx)
	logger.Info("Activity", "name", name)

	if val := ctx.Value(ctxpropagation.PropagateKey); val != nil {
		vals := val.(ctxpropagation.Values)
		log.Infof("custom context propagated to activity, key: %s, val: %s", vals.Key, vals.Value)
	} else {
		log.Info("no value")
	}

	return "Hello " + name + "!", nil
}
