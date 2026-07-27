package service

import (
	"gateway/application/port/in"
	"gateway/application/port/out"
)

type controlPlaneMetricService struct {
	metricPort out.ControlPlaneMetricPort
}

func NewControlPlaneMetricService(metricPort out.ControlPlaneMetricPort) in.ControlPlaneMetricUseCase {
	return &controlPlaneMetricService{metricPort: metricPort}
}

func (c controlPlaneMetricService) GetMetric() string {
	return c.metricPort.GetMetric()
}
