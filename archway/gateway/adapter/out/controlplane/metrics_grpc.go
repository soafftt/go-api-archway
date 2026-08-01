package controlplane

import (
	"context"
	"core/utils"
	"gateway/adapter/config/client"
	"gateway/application/port/out"
	pb "protobuf/matrics"
)

var logger = utils.GetLogger()

type GrpcMetricOutPort out.ControlPlaneMetricPort

type MetricsLookup struct {
	serviceClient pb.MetricsServiceClient
}

func NewMetricsLookup(grpc client.GrpcClient) GrpcMetricOutPort {
	return &MetricsLookup{
		serviceClient: pb.NewMetricsServiceClient(grpc.GetClient()),
	}
}

func (c MetricsLookup) GetMetric() string {
	metric, err := c.serviceClient.GetMetrics(context.Background(), &pb.GetMetricsRequest{})
	if err != nil {
		// logging
		logger.ErrorW("getMetric control-plane error", err)
		return ""
	}

	return metric.Data
}
