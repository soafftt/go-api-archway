package main

import (
	"unsafe"

	// 필수 값.
	_ "github.com/lestrrat-go/jwx/v3/jwk/ecdsa"
)

type testStruct struct {
	value string
}

func main() {
	// app, err := InitializeNewApp()
	// if err != nil {
	// 	panic(err)
	// }

	// app.ReverseServer.Start()
	println(unsafe.Sizeof(testStruct{}))

}
