package main

type testStruct struct {
	value string
}

func main() {
	app, err := InitializeNewApp()
	if err != nil {
		panic(err)
	}

	app.ReverseServer.Start()
}
