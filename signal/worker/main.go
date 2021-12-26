package main

import (
	"github.com/jasonsoft/learning-temporal/signal"
	"github.com/nite-coder/blackbear/pkg/log"
	"github.com/nite-coder/blackbear/pkg/log/handler/console"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	logger := log.New()
	clog := console.New()
	logger.AddHandler(clog, log.AllLevels...)
	log.SetLogger(logger)

	// The client and worker are heavyweight objects that should be created once per process.
	c, err := client.NewClient(client.Options{
		Namespace: "default",
	})
	if err != nil {
		panic(err)
	}
	defer c.Close()

	w := worker.New(c, "signal", worker.Options{})

	w.RegisterWorkflow(signal.WithdrawWorkflow)
	w.RegisterActivity(signal.CreateWithdrawOrderActivity)
	w.RegisterActivity(signal.ApproveActivity)
	w.RegisterActivity(signal.RejectActivity)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker", err)
	}
}
