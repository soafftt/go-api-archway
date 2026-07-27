module gateway/controller

go 1.26.1

require (
	github.com/caarlos0/env/v11 v11.4.1
	github.com/joho/godotenv v1.5.1
	github.com/valkey-io/valkey-go v1.0.72
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/golang/protobuf v1.5.4 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus v1.1.0 // indirect
	github.com/grpc-ecosystem/go-grpc-middleware/v2 v2.3.3 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.4 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/prometheus/client_golang v1.24.1 // indirect
	github.com/prometheus/client_model v0.6.2 // indirect
	github.com/prometheus/common v0.70.1 // indirect
	github.com/prometheus/procfs v0.21.1 // indirect
	golang.org/x/net v0.57.0 // indirect
	golang.org/x/sys v0.47.0 // indirect
	golang.org/x/text v0.40.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20260724162435-b2f20204f0df // indirect
	google.golang.org/protobuf v1.36.11 // indirect
)

require (
	gateway/protobuf v0.0.0
	github.com/google/wire v0.7.0
	google.golang.org/grpc v1.82.1
)

replace (
	gateway/core => ../../core
	gateway/protobuf => ../../protobuf
)
