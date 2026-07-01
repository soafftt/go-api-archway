package main

import (
	"context"
	"errors"
	"log"
)

func main() {
	app, err := InitializeGatewayControllerApp()
	if err != nil {
		log.Fatal(err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ready := make(chan struct{})
	startErr := make(chan error, 1)
	go func() {
		if err := app.ListenRouteOperations(ctx, ready); err != nil && !errors.Is(err, context.Canceled) {
			startErr <- err
		}
	}()
	select {
	case <-ready:
	case err := <-startErr:
		log.Fatalf("route operation listener start failed: %v", err)
	}
	app.LoadRouteCache()

	app.unixServer.Start()
	cancel()
}
