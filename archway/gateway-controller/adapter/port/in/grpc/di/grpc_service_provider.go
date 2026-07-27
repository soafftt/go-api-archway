package di

import (
	"gateway/controller/adapter/port/in/grpc/handler"
	pb "protobuf"
	pbMetrics "protobuf/matrics"

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

func NewGrpcServiceRegistrars(
	upstreamLookupServer pb.UpstreamLookupServiceServer,
	metricsServer pbMetrics.MetricsServiceServer,
) []GrpcServiceRegistrar {
	return []GrpcServiceRegistrar{
		grpcServiceRegistrarFunc(func(registrar grpc.ServiceRegistrar) {
			pb.RegisterUpstreamLookupServiceServer(registrar, upstreamLookupServer)
		}),
		grpcServiceRegistrarFunc(func(registrar grpc.ServiceRegistrar) {
			pbMetrics.RegisterMetricsServiceServer(registrar, metricsServer)
		}),
	}
}

var GrpcServiceProviderSet = wire.NewSet(
	handler.NewUpstreamLookupController,
	wire.Bind(new(pb.UpstreamLookupServiceServer), new(*handler.UpstreamLookupController)),
	handler.NewMetricsController,
	wire.Bind(new(pbMetrics.MetricsServiceServer), new(*handler.MetricsController)),
	NewGrpcServiceRegistrars,
	wire.Struct(new(GrpcServiceProvider), "*"),
)
