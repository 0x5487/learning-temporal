package main

import (
	"context"
	"time"

	"github.com/nite-coder/blackbear/pkg/log"
	"github.com/nite-coder/blackbear/pkg/log/handler/console"
	"go.temporal.io/sdk/client"
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

	w := worker.New(c, "idempotent1", worker.Options{})

	w.RegisterWorkflow(myWorkflow)
	w.RegisterActivity(Activity1)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker1", err)
	}
}

func myWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		TaskQueue:              "idempotent1",
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	log.Str("workflow runID", workflow.GetInfo(ctx).WorkflowExecution.RunID).Infof("HelloWorld workflow started. name: %s", name)

	var result string
	err := workflow.ExecuteActivity(ctx, "Activity1", "a1").Get(ctx, &result)
	if err != nil {
		log.Err(err).Error("Activity1 failed.")
		return "", err
	}

	log.Info("waiting......")
	time.Sleep(3 * time.Second)

	ao = workflow.ActivityOptions{
		TaskQueue:              "idempotent2",
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	var result2 Message
	err = workflow.ExecuteActivity(ctx, "Activity2", "a2").Get(ctx, &result2)
	if err != nil {
		log.Err(err).Error("Activity2 failed.")
		return "", err
	}

	log.Infof("workflow completed. result: %s", result+result2.Content)
	return result + result2.Content, nil
}

type Message struct {
	ID        string
	Content   string
	CreatedAt time.Time
}

func Activity1(ctx context.Context, name string) (string, error) {
	log.Info("activity_1 is calling")
	return "Hello " + name, nil
}
