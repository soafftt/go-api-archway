package di

import (
	"gateway/controller/adapter/port/in/grpc/handler"
	pb "gateway/protobuf"

	"github.com/google/wire"
	"google.golang.org/grpc"
)

type GrpcServiceRegistrar interface {
	Register(grpc.ServiceRegistrar)
}

type grpcServiceRegistrarFunc func(grpc.ServiceRegistrar)

func (g grpcServiceRegistrarFunc) Register(registrar grpc.ServiceRegistrar) {
	g(registrar)
}

type GrpcServiceProvider struct {
	Registrars []GrpcServiceRegistrar
}

func NewGrpcServiceRegistrars(upstreamLookupServer pb.UpstreamLookupServiceServer) []GrpcServiceRegistrar {
	return []GrpcServiceRegistrar{
		grpcServiceRegistrarFunc(func(registrar grpc.ServiceRegistrar) {
			pb.RegisterUpstreamLookupServiceServer(registrar, upstreamLookupServer)
		}),
	}
}

var GrpcServiceProviderSet = wire.NewSet(
	handler.NewUpstreamLookupController,
	wire.Bind(new(pb.UpstreamLookupServiceServer), new(*handler.UpstreamLookupController)),
	NewGrpcServiceRegistrars,
	wire.Struct(new(GrpcServiceProvider), "*"),
)
