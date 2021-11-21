package main

import (
	"fmt"
	"time"

	"github.com/nite-coder/blackbear/pkg/log"
	"github.com/nite-coder/blackbear/pkg/log/handler/console"
	"go.temporal.io/api/enums/v1"
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

	w := worker.New(c, "loop", worker.Options{})

	w.RegisterWorkflow(myWorkflow)
	w.RegisterWorkflow(LoopChildWorkflow)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker", err)
	}
}

// myWorkflow is a Hello World workflow definition.
func myWorkflow(ctx workflow.Context) (string, error) {
	execution := workflow.GetInfo(ctx).WorkflowExecution
	childID := fmt.Sprintf("child_workflow:%v", execution.RunID)
	cwo := workflow.ChildWorkflowOptions{
		// Do not specify WorkflowID if you want Temporal server to generate a unique ID for child execution
		WorkflowID:        childID,
		ParentClosePolicy: enums.PARENT_CLOSE_POLICY_TERMINATE, // it doesn't work for NewContinueAsNewError scenario
	}
	ctx = workflow.WithChildOptions(ctx, cwo)

	logger := workflow.GetLogger(ctx)
	logger.Info("myworkflow started")

	workflow.Go(ctx, func(ctx workflow.Context) {
		defer logger.Info("first goroutine completed.")

		workflow.ExecuteChildWorkflow(ctx, LoopChildWorkflow)

	})

	workflow.Sleep(ctx, 15*time.Second)

	logger.Info("myworkflow completed.")

	return "", nil
}

func LoopChildWorkflow(ctx workflow.Context) error {
	for {
		log.Infof("=== begin LoopChildWorkflow, now: %s ===", time.Now().String())
		workflow.Sleep(ctx, 3*time.Second)
	}

	return nil
	//return workflow.NewContinueAsNewError(ctx, LoopChildWorkflow)
}
