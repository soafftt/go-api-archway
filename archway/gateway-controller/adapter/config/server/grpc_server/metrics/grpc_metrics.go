package metrics

import (
	"github.com/google/wire"
	mp "github.com/grpc-ecosystem/go-grpc-middleware/providers/prometheus"
	p "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type GrpcServerMetrics interface {
	GetRegistry() *p.Registry
	GetServerMetrics() *mp.ServerMetrics
}

type grpcServerMetrics struct {
	registry *p.Registry
	metrics  *mp.ServerMetrics
}

func NewServerMetrics() GrpcServerMetrics {
	metrics := mp.NewServerMetrics(mp.WithServerHandlingTimeHistogram())
	registry := p.NewRegistry()
	// 기본 Go/Process 런타임 메트릭을 레지스트리에 함께 등록
	registry.MustRegister(collectors.NewGoCollector())
	registry.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))

	registry.MustRegister(metrics)

	return &grpcServerMetrics{
		registry: registry,
		metrics:  metrics,
	}
}

func (s *grpcServerMetrics) GetRegistry() *p.Registry {
	return s.registry
}

func (s *grpcServerMetrics) GetServerMetrics() *mp.ServerMetrics {
	return s.metrics
}

var GrpcMetricsProvider = wire.NewSet(NewServerMetrics)
