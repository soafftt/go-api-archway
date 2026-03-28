package main

import (
	"crypto"
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwk"

	// 필수 값.
	_ "github.com/lestrrat-go/jwx/v3/jwk/ecdsa"
)

func main() {
	// app, err := InitializeNewApp()
	// if err != nil {
	// 	panic(err)
	// }

	// app.ReverseServer.Start()

	keyBase64 := "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"
	docoded, _ := base64.StdEncoding.DecodeString(keyBase64)

	key, err := jwk.ParseKey(docoded)
	if err != nil {
		fmt.Printf("failed to parse JWK: %s\n", err)
		return
	}

	// var privKey ecdsa.PrivateKey
	var privKey crypto.PrivateKey
	if err := jwk.Export(key, &privKey); err != nil {
		fmt.Printf("failed to export private key: %s\n", err)
		return
	}

	c := privKey.(*ecdsa.PrivateKey)

	println("OK")

	// pKey, err := key.PublicKey()
	// if err != nil {
	// 	println("아?")
	// 	fmt.Printf("failed to get public key: %s\n", err)
	// 	return
	// }

	t := jwt.New(jwt.SigningMethodES256)

	t.Claims = jwt.MapClaims{
		"foo": "bar",
	}

	s, err := t.SignedString(c)
	if err != nil {
		panic(err)
	}

	println(s)

	// decodeJwt, err := jwt.Parse(
	// 	s,
	// 	func(token *jwt.Token) (interface{}, error) {
	// 		return &privKey.PublicKey, nil
	// 	},
	// )

	// if err != nil {
	// 	panic(err)
	// }

	// println(decodeJwt.Valid)
	// println(decodeJwt.Claims.(jwt.MapClaims)["foo"].(string))

	// OK
}
