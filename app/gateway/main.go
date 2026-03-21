package main

func main() {
	app, err := InitializeNewApp()
	if err != nil {
		panic(err)
	}

	app.ReverseProxy.Start()
}
