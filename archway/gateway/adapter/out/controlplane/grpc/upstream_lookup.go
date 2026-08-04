package grpc

import (
	"context"
	"core/errs"
	"gateway/adapter/config/client"
	"gateway/application/port/out"
	pb "protobuf"

	"google.golang.org/grpc/metadata"
)

type upstreamLookup struct {
	serviceClient pb.UpstreamLookupServiceClient
}

func NewUpstreamLookup(grpc client.GrpcClient) out.UpstreamLookupGrpcPort {
	return &upstreamLookup{
		serviceClient: pb.NewUpstreamLookupServiceClient(grpc.GetClient()),
	}
}

func (u *upstreamLookup) GetUpstreamInfo(
	path string,
	accessToken *string,
) (out.UpStreamLookupPortResult, error) {
	ctx := context.Background()
	if accessToken != nil {
		md := metadata.Pairs("authorization", *accessToken)
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	result, err := u.serviceClient.Lookup(ctx, &pb.UpstreamLookupRequest{Path: path})

	var lookupResult out.UpStreamLookupPortResult
	if err != nil {
		return lookupResult, errs.ToArchwayFromError(err)
	}

	if result.Error != "" {
		return lookupResult, errs.ToArchwayError(result.Error)
	}

	return out.UpStreamLookupPortResult{
		ServiceName:     result.Info.ServiceName,
		Host:            result.Info.Host,
		Path:            result.Info.Path,
		OriginalPath:    result.Info.OriginalPath,
		Method:          result.Info.Method,
		ResponseTimeout: result.Info.ResponseTimeout,
		RequestTimeout:  result.Info.RequestTimeout,
		CacheTimeout:    result.Info.CacheTimeout,
		UserKey:         result.Info.UserKey,
		RateLimitCount:  result.Info.RateLimitCount,
	}, nil
}
