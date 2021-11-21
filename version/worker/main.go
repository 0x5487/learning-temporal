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

	w := worker.New(c, "version", worker.Options{})

	w.RegisterWorkflow(VersionWorkflow)
	w.RegisterActivity(ActivityA)
	w.RegisterActivity(ActivityB)
	w.RegisterActivity(ActivityC)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker", err)
	}
}

func VersionWorkflow(ctx workflow.Context, name string) (string, error) {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	logger := log.FromContext(context.Background()).Logger()

	v := workflow.GetVersion(ctx, "", workflow.DefaultVersion, 1)
	logger.Infof("current versio: %d", v)

	var result string

	switch v {
	case workflow.DefaultVersion:
		err := workflow.ExecuteActivity(ctx, "ActivityA", v).Get(ctx, nil)
		if err != nil {
			return "", err
		}
	case 1:
		err := workflow.ExecuteActivity(ctx, "ActivityB", v).Get(ctx, nil)
		if err != nil {
			return "", err
		}
	}

	logger.Infof("Version workflow completed.  result: %s", result)

	return result, nil
}

func ActivityA(ctx context.Context, version int) error {
	log.Infof("ActivityA is calling. version: %d", version)

	return nil
}

func ActivityB(ctx context.Context, version int) error {
	log.Infof("ActivityB is calling. version: %d", version)
	return nil
}

func ActivityC(ctx context.Context, version int) error {
	log.Infof("ActivityC is calling. version: %d", version)
	return nil
}
