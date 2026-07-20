module gateway

go 1.26.1

require (
	github.com/caarlos0/env/v11 v11.4.1
	github.com/google/wire v0.7.0
	github.com/joho/godotenv v1.5.1
	golang.org/x/time v0.15.0
)

require gateway/protobuf v0.0.0

require (
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	golang.org/x/text v0.36.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260715232425-e75dac1f907d // indirect
	google.golang.org/grpc v1.82.1 // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

replace (
	gateway/core => ../../core
	gateway/protobuf => ../../protobuf
)
