module gateway

go 1.25.0

require (
	gateway/common v0.0.0
	github.com/google/wire v0.7.0
)

require (
	github.com/caarlos0/env/v11 v11.4.0 // indirect
	github.com/golang-jwt/jwt/v5 v5.3.1 // indirect
	github.com/joho/godotenv v1.5.1 // indirect
)

replace gateway/common => ../../common
