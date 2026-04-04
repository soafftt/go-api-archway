module gateway

go 1.26.1

require (
	gateway/common v0.0.0
	github.com/caarlos0/env/v11 v11.4.0
	github.com/google/wire v0.7.0
	github.com/joho/godotenv v1.5.1
)

require (
	github.com/decred/dcrd/dcrec/secp256k1/v4 v4.4.0 // indirect
	github.com/lestrrat-go/blackmagic v1.0.4 // indirect
	github.com/lestrrat-go/httpcc v1.0.1 // indirect
	github.com/lestrrat-go/httprc/v3 v3.0.2 // indirect
	github.com/lestrrat-go/option/v2 v2.0.0 // indirect
	github.com/segmentio/asm v1.2.1 // indirect
	golang.org/x/sys v0.39.0 // indirect
)

require (
	github.com/goccy/go-json v0.10.6 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/lestrrat-go/jwx/v3 v3.0.13
)

replace gateway/common => ../../common
