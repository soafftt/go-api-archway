package client

import (
	"context"
	"gateway/adapter/config"
	"net"

	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
)

type GrpcClient interface {
	GetClient() *grpc.ClientConn
}

type grpcClient struct {
	grpcClientConn *grpc.ClientConn
}

func NewGrpcClient(appConfig *config.AppConfig) GrpcClient {
	opts := []grpc.DialOption{
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithContextDialer(func(ctx context.Context, addr string) (net.Conn, error) {
			return net.Dial(appConfig.ClientNetworkConfig.Network, appConfig.ClientNetworkConfig.UnixSocketPath)
		}),
		grpc.WithDefaultCallOptions(grpc.MaxCallRecvMsgSize(appConfig.GrpcClient.MaxRecvMsgSize)),
		grpc.WithKeepaliveParams(keepalive.ClientParameters{
			Time:                appConfig.GrpcClient.KeepaliveTime,
			Timeout:             appConfig.GrpcClient.KeepaliveTimeout,
			PermitWithoutStream: appConfig.GrpcClient.PermitWithoutStream,
		}),
		grpc.WithConnectParams(grpc.ConnectParams{
			MinConnectTimeout: appConfig.GrpcClient.MinConnectTimeout,
			Backoff: backoff.Config{
				BaseDelay:  appConfig.GrpcClient.BackoffBaseDelay,
				Multiplier: appConfig.GrpcClient.BackoffMultiplier,
				Jitter:     appConfig.GrpcClient.BackoffJitter,
				MaxDelay:   appConfig.GrpcClient.BackoffMaxDelay,
			},
		}),
	}

	if appConfig.GrpcClient.DisableRetry {
		opts = append(opts, grpc.WithDisableRetry())
	}

	conn, err := grpc.NewClient("passthrough:///"+appConfig.ClientNetworkConfig.UnixSocketPath, opts...)

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
