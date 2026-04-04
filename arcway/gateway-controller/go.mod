module gateway/controller

go 1.26.1

require (
	github.com/caarlos0/env/v11 v11.4.0
	github.com/joho/godotenv v1.5.1
	github.com/valkey-io/valkey-go v1.0.72
)

require golang.org/x/sys v0.42.0 // indirect

require (
	gateway/common v0.0.0
	github.com/google/wire v0.7.0
)

replace gateway/common => ../../common
