module gateway

go 1.26.1

require (
	github.com/caarlos0/env/v11 v11.4.1
	github.com/google/wire v0.7.0
	github.com/joho/godotenv v1.5.1
)

replace gateway/core => ../../core
