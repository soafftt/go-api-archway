package client

import (
	"context"
	"gateway/adapter/config/appconfig"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GrpcClientConfig interface {
	GetGrpcClientProperties() appconfig.GrpcClientProperties
}

type GrpcClient interface {
	GetClient() *grpc.ClientConn
}

type grpcClient struct {
	grpcClientConn *grpc.ClientConn
}

func NewGrpcClient(
	networkConfig appconfig.ClientNetworkConfig,
	grpcConfig GrpcClientConfig,
) GrpcClient {
	networkProperties := networkConfig.GetClientNetworkProperties()
	grpcProperties := grpcConfig.GetGrpcClientProperties()

	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.Dial(networkProperties.Network, networkProperties.UnixSocketPath)
		}),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(grpcProperties.MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                grpcProperties.KeepaliveTime,
			Timeout:             grpcProperties.KeepaliveTimeout,
			PermitWithoutStream: grpcProperties.PermitWithoutStream,
		}),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: grpcProperties.MinConnectTimeout,
			Backoff: backoff.Config{
				BaseDelay:  grpcProperties.BackoffBaseDelay,
				Multiplier: grpcProperties.BackoffMultiplier,
				Jitter:     grpcProperties.BackoffJitter,
				MaxDelay:   grpcProperties.BackoffMaxDelay,
			},
		}),
	}

	if grpcProperties.DisableRetry {
		opts = append(opts, grpc.WithDisableRetry())
	}

	conn, err := grpc.NewClient("passthrough:///"+networkProperties.UnixSocketPath, opts...)
	if err != nil {
		panic(err)
	}

	conn.Connect()
	return &grpcClient{
		grpcClientConn: conn,
	}
}

func (u *grpcClient) GetClient() *grpc.ClientConn {
	return u.grpcClientConn
}
