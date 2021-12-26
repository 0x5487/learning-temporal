package main

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jasonsoft/learning-temporal/signal"
	"github.com/nite-coder/blackbear/pkg/log"
	"github.com/nite-coder/blackbear/pkg/log/handler/console"
	"go.temporal.io/sdk/client"
)

func main() {
	logger := log.New()
	clog := console.New()
	logger.AddHandler(clog, log.AllLevels...)
	log.SetLogger(logger)

	// The client is a heavyweight object that should be created once per process.
	c, err := client.NewClient(client.Options{
		Namespace: "default",
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	err = Start(c)
	if err != nil {
		panic(err)
	}
}

func Start(c client.Client) error {

	workflowID := "withdraw-" + uuid.NewString()
	workflowOptions := client.StartWorkflowOptions{
		ID:        workflowID,
		TaskQueue: "signal",
	}

	createWithdrawOrderRequest := signal.CreateWithdrawOrderRequest{
		Amount: 100,
	}

	we, err := c.ExecuteWorkflow(context.Background(), workflowOptions, "WithdrawWorkflow", createWithdrawOrderRequest)
	if err != nil {
		return err
	}

	log.Infof("Started workflow, WORKER_FLOWER_ID: %s, RUN_ID: %s", we.GetID(), we.GetRunID())

	time.Sleep(3 * time.Second)

	signalVal := signal.MySignal{State: 3}
	err = c.SignalWorkflow(context.Background(), workflowID, we.GetRunID(), "withdraw_signal", signalVal)
	if err != nil {
		return err
	}

	return nil
}
