package main

import (
	"context"

	"github.com/jasonsoft/log/v2"
	"github.com/jasonsoft/log/v2/handlers/console"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/worker"
)

func main() {
	// set up log target
	log.
		Str("app_id", "worker").
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

	w := worker.New(c, "idempotent2", worker.Options{})

	w.RegisterActivity(Activity2)

	err = w.Run(worker.InterruptCh())
	if err != nil {
		log.Fatalf("Unable to start worker2", err)
	}
}

func Activity2(ctx context.Context, name string) (string, error) {
	log.Info("activity_2 is calling")
	return "Hello " + name + "!", nil
}
