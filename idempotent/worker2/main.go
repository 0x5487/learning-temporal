package main

import (
	"context"
	"time"

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

type Message struct {
	ID        string
	Content   string
	CreatedAt time.Time
}

func Activity2(ctx context.Context, name string) (*Message, error) {
	log.Info("activity_2 is calling")
	
	result := Message{
		ID:        "abc",
		Content:   "content1111",
		CreatedAt: time.Now(),
	}
	return &result, nil
}
