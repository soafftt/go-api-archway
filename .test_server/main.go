package main

import "net/http"

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	if err := http.ListenAndServe(":8084", nil); err != nil {
		panic(err)
	}
}
