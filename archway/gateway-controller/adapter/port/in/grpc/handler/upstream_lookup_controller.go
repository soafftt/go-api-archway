package handler

import (
	"context"
	"gateway/controller/application/port/in"
	pb "gateway/protobuf"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UpstreamLookupController struct {
	pb.UnimplementedUpstreamLookupServiceServer
	upstreamLookupUseCase in.UpstreamLookupUseCase
}

func NewUpstreamLookupController(upstreamLookupUseCase in.UpstreamLookupUseCase) *UpstreamLookupController {
	return &UpstreamLookupController{upstreamLookupUseCase: upstreamLookupUseCase}
}

func (u *UpstreamLookupController) Lookup(_ context.Context, request *pb.UpstreamLookupRequest) (*pb.UpstreamLookupResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	var accessToken *string = nil
	if token := request.GetAccessToken(); token != "" {
		accessToken = &token
	}

	lookupResult := u.upstreamLookupUseCase.Lookup(
		request.GetVersion(),
		request.GetService(),
		request.GetDomain(),
		request.GetPath(),
		accessToken,
	)
	if lookupResult.Error != nil {
		return &pb.UpstreamLookupResponse{Error: lookupResult.Error.Error()}, nil
	}

	userKey := lookupResult.Info.UserKey

	return &pb.UpstreamLookupResponse{
		Info: &pb.UpstreamInfo{
			Host:            lookupResult.Info.Host,
			Path:            lookupResult.Info.Path,
			OriginalPath:    lookupResult.Info.OriginalPath,
			Method:          lookupResult.Info.Method,
			RequestTimeout:  lookupResult.Info.RequestTimeout,
			ResponseTimeout: lookupResult.Info.ResponseTimeout,
			CacheTimeout:    lookupResult.Info.CacheTimeout,
			UserKey:         userKey,
			RateLimitCount:  lookupResult.Info.RateLimitCount,
		},
	}, nil
}
