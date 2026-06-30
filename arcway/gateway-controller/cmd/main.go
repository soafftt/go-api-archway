package main

import "log"

func main() {
	app, err := InitializeGatewayControllerApp()
	if err != nil {
		log.Fatal(err)
	}
	app.unixServer.Start()
}
