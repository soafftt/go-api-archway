package main

import (
	"log"
)

func main() {
	app, err := InitializeGatewayControllerApp()
	if err != nil {
		log.Fatal(err)
	}

	app.LoadRouteCache()
	app.listenerServer.Start()
}
