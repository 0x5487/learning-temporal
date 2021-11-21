package main

import (
	"context"
	"errors"
	"time"

	"github.com/nite-coder/blackbear/pkg/log"
	"github.com/nite-coder/blackbear/pkg/log/handler/console"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	logger := log.New()
	clog := console.New()
	logger.AddHandler(clog, log.AllLevels...)
	log.SetLogger(logger)

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
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        time.Second * 100, // 100 * InitialInterval
			MaximumAttempts:        3,                 // run 3 times maximum
			NonRetryableErrorTypes: []string{"no_retry"},
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

	log.Info("============new retry option==============")
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

	if name == "noretry" {
		log.Info("no retry error")
		err := errors.New("oooops....")
		return "", temporal.NewNonRetryableApplicationError("can't retry", "no_retry", err)
	}

	return "", errors.New("throw error happened")
}

func ThrowPanicActivity(ctx context.Context, name string) (string, error) {
	log.Info("ThrowPanicActivity is calling")
	err := errors.New("panic oops")
	panic(err)
}
