module gateway

go 1.25.0

require (
	gateway/common v0.0.0
	github.com/caarlos0/env/v11 v11.4.0
	github.com/google/wire v0.7.0
	github.com/joho/godotenv v1.5.1
)

replace gateway/common => ../../common
