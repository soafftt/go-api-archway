package controlplane

import (
	"context"
	"gateway/adapter/config/client"
	"gateway/application/port/out"
	pb "protobuf/matrics"
)

type MetricsLookup struct {
	serviceClient pb.MetricsServiceClient
}

func NewMetricsLookup(grpc client.GrpcClient) out.ControlPlaneMetricPort {
	return &MetricsLookup{
		serviceClient: pb.NewMetricsServiceClient(grpc.GetClient()),
	}
}

func (c MetricsLookup) GetMetric() string {
	metric, err := c.serviceClient.GetMetrics(context.Background(), &pb.GetMetricsRequest{})
	if err != nil {
		// logging
		return ""
	}

	return metric.Data
}
