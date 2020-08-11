package main

import (
	"context"
	"errors"
	"time"

	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	defer log.Flush()
	// set up log target
	log.
		Str("app_id", "worker1").
		SaveToDefault()

	clog := console.New()
	log.AddHandler(clog, log.AllLevels...)

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalf("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "error-handling", worker.Options{})

	w.RegisterWorkflow(myWorkflow)
	w.RegisterActivity(ThrowErrorActivity)
	w.RegisterActivity(ThrowPanicActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker", err)
	}
}

// myWorkflow is a Hello World workflow definition.
func myWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:    time.Second,
			BackoffCoefficient: 2.0,
			MaximumInterval:    time.Minute,
			MaximumAttempts:    3, // retry 3 times
		},
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("myWorkflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, "ThrowErrorActivity", name).Get(ctx, &result)
	if err != nil {
		log.Err(err).Error("ThrowErrorActivity failed.")
		//return "", err
	}

	err = workflow.ExecuteActivity(ctx, "ThrowPanicActivity", name).Get(ctx, &result)
	if err != nil {
		logger.Error("ThrowPanicActivity failed.", "Error", err)
		return "", err
	}

	logger.Info("myWorkflow completed.", "result", result)

	return result, nil
}

func ThrowErrorActivity(ctx context.Context, name string) (string, error) {
	log.Info("ThrowErrorActivity is calling")
	return "", errors.New("throw error happened")
}

func ThrowPanicActivity(ctx context.Context, name string) (string, error) {
	log.Info("ThrowPanicActivity is calling")
	err := errors.New("panic oops")
	panic(err)
}
