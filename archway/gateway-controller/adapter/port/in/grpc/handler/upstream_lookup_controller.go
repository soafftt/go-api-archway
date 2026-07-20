package handler

import (
	"context"
	"fmt"
	"gateway/controller/application/port/in"
	pb "gateway/protobuf"
	"strings"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type UpstreamLookupController struct {
	pb.UnimplementedUpstreamLookupServiceServer
	upstreamLookupUseCase in.UpstreamLookupUseCase
}

func NewUpstreamLookupController(upstreamLookupUseCase in.UpstreamLookupUseCase) *UpstreamLookupController {
	return &UpstreamLookupController{upstreamLookupUseCase: upstreamLookupUseCase}
}

func (u *UpstreamLookupController) Lookup(ctx context.Context, request *pb.UpstreamLookupRequest) (*pb.UpstreamLookupResponse, error) {
	if request == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}
	version, service, domain, lookupPath, err := parseLookupPath(request.GetPath())
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, err.Error())
	}

	var accessToken *string = nil
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		authValues := md.Get("authorization")
		if len(authValues) > 0 {
			accessToken = &authValues[0]
		}
	}

	lookupResult := u.upstreamLookupUseCase.Lookup(
		version,
		service,
		domain,
		lookupPath,
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

func parseLookupPath(path string) (version, service, domain, lookupPath string, err error) {
	segments := strings.Split(strings.Trim(path, "/"), "/")
	if len(segments) < 4 {
		return "", "", "", "", fmt.Errorf("invalid path")
	}

	serviceIndex := 0
	versionIndex := 1
	if isVersionSegment(segments[0]) {
		serviceIndex = 1
		versionIndex = 0
	}

	service = segments[serviceIndex]
	version = segments[versionIndex]
	domain = segments[2]
	lookupPath = strings.Join(segments[2:], "/")
	return version, service, domain, lookupPath, nil
}

func isVersionSegment(segment string) bool {
	if len(segment) < 2 || segment[0] != 'v' {
		return false
	}
	for _, character := range segment[1:] {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
