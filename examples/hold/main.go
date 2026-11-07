package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"time"

	power "keepalive"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	acquireCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	hold, err := power.New().Acquire(acquireCtx, power.Options{
		KeepDisplay: false,
		Reason:      "serving paired devices",
	})
	cancel()
	if err != nil {
		log.Printf("cannot prevent sleep: %v; %s", err, power.Hint(err))
		return
	}
	defer func() {
		if err := hold.Release(); err != nil {
			log.Printf("release: %v", err)
		}
	}()
	log.Printf("sleep protection active: %s; Ctrl+C to release", hold.Describe())
	var lost <-chan error
	if watcher, ok := hold.(power.Watcher); ok {
		lost = watcher.Lost()
	}
	select {
	case <-ctx.Done():
	case err := <-lost:
		log.Printf("sleep protection lost: %v", err)
	}
}
