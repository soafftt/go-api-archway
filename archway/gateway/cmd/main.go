package main

func main() {
	// TODO: cron 을 이용하여, unixSocket 이 떠 있는지 주기적으로 확인해야 함. unixSocket 이 없다면, gateway 는 장애임.
	app, err := InitializeApp()
	if err != nil {
		panic(err)
	}

	app.proxyServer.Start()
}
