package main

import (
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"

	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
	"go.temporal.io/sdk/workflow"
)

func main() {
	defer log.Flush()
	defer func() {
		if r := recover(); r != nil {
			// unknown error
			err, ok := r.(error)
			if !ok {
				err = fmt.Errorf("unknown error: %v", r)
			}
			trace := make([]byte, 4096)
			runtime.Stack(trace, true)
			log.Str("stack_trace", string(trace)).Err(err).Panic("unknown error")
		}
	}()

	log.
		Str("app_id", "worker").
		Str("env", "dev").
		SaveToDefault()

	clog := console.New()
	log.AddHandler(clog, log.AllLevels...)

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{})
	if err != nil {
		log.Fatalf("Unable to create client", err)
	}
	defer c.Close()

	w := worker.New(c, "cron", worker.Options{})

	w.RegisterWorkflow(CronWorkflow)
	w.RegisterActivity(CronActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker", err)
	}
}

// CronWorkflow is a cron workflow definition.
func CronWorkflow(ctx workflow.Context, name string) error {
	ao := workflow.ActivityOptions{
		ScheduleToStartTimeout: time.Minute,
		StartToCloseTimeout:    time.Minute,
	}
	ctx = workflow.WithActivityOptions(ctx, ao)

	err := workflow.ExecuteActivity(ctx, "CronActivity").Get(ctx, nil)
	if err != nil {
		log.Err(err).Error("CronActivity failed.")
		return err
	}

	if workflow.HasLastCompletionResult(ctx) {
		log.Info("HasLastCompletionResult")
	} else {
		log.Info("no result from last task")
	}

	log.Info("CronWorkflow completed.")

	return nil
}

func CronActivity(ctx context.Context) error {
	log.Infof("Begin CronActivity at %s", time.Now().String())
	return nil
}
