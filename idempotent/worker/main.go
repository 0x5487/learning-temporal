package main

import (
	"context"
	"time"

	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	// set up log target
	log.
		Str("app_id", "starter").
		SaveToDefault()

	clog := console.New()
	log.AddHandler(clog, log.AllLevels...)
	defer log.Flush() // flush log buffer

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalf("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "idempotent", worker.Options{})

	w.RegisterWorkflow(myWorkflow)
	w.RegisterActivity(Activity1)
	w.RegisterActivity(Activity2)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker", err)
	}
}

func myWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := workflow.GetLogger(ctx)
	logger.Info("HelloWorld workflow started", "name", name)

	var result string
	err := workflow.ExecuteActivity(ctx, "Activity1", "a1").Get(ctx, &result)
	if err != nil {
		logger.Error("Activity1 failed.", "Error", err)
		return "", err
	}

	log.Info("waiting......")
	time.Sleep(3 * time.Second)

	var result2 string
	err = workflow.ExecuteActivity(ctx, "Activity2", "a2").Get(ctx, &result2)
	if err != nil {
		logger.Error("Activity2 failed.", "Error", err)
		return "", err
	}

	logger.Info("workflow completed.", "result", result+result2)

	return result, nil
}

func Activity1(ctx context.Context, name string) (string, error) {
	log.Info("activity_1 is calling")
	return "Hello " + name + "!", nil
}

func Activity2(ctx context.Context, name string) (string, error) {
	log.Info("activity_2 is calling")
	return "Hello " + name + "!", nil
}
