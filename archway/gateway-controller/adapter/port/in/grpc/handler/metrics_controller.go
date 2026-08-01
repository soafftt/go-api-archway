package handler

import (
	"bytes"
	"context"
	"core/utils"
	"gateway/controller/adapter/config/server/grpc_server/metrics"
	pb "protobuf/matrics"
	"sync"

	"github.com/prometheus/common/expfmt"
)

var logger = utils.GetLogger()
var bufferPool = sync.Pool{
	New: func() any {
		// 10kb
		return bytes.NewBuffer(make([]byte, 0, (10 * 1024)))
	},
}

type MetricsController struct {
	pb.UnimplementedMetricsServiceServer
	grpcServerMetrics metrics.GrpcServerMetrics
}

func NewMetricsController(grpcServerMetrics metrics.GrpcServerMetrics) *MetricsController {
	return &MetricsController{
		grpcServerMetrics: grpcServerMetrics,
	}
}

// ctx 와 request 사용하지 않음.
func (m *MetricsController) GetMetrics(
	ctx context.Context,
	request *pb.GetMetricsRequest,
) (*pb.GetMetricsResponse, error) {
	metricsFamily, err := m.grpcServerMetrics.GetRegistry().Gather()
	if err != nil {
		return nil, err
	}

	buffer := bufferPool.Get().(*bytes.Buffer)
	buffer.Reset()
	defer func() {
		buffer.Reset()
		bufferPool.Put(buffer)
	}()

	enc := expfmt.NewEncoder(buffer, expfmt.NewFormat(expfmt.TypeTextPlain))
	for index := range metricsFamily {
		if err := enc.Encode(metricsFamily[index]); err != nil {
			// 로깅
			logger.ErrorW("grpc metrics encoding fail", err)
			return nil, err
		}
	}

	return &pb.GetMetricsResponse{Data: utils.ToStringFromBytes(buffer.Bytes())}, nil
}
