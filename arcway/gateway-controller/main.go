package main

func main() {
	app, err := InitializeApp()
	if err != nil {
		panic(err)
	}

	app.Server.StartUnixSocketServer()
}
