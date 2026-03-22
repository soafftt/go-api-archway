package di

import (
	"gateway/server/middleware"

	"github.com/google/wire"
)

type MiddlewareContainers struct {
	Middlewares []middleware.Middleware
}

/**
 * MiddlewareContainers는 여러 미들웨어를 담는 컨테이너입니다. Chain 함수에 이 컨테이너의 Middlewares 필드를 전달하여 여러 미들웨어를 한 번에 등록할 수 있습니다.
 */
func NewMiddlewareContainers(
	upstreamCheckMiddleware middleware.UpstreamCheckMiddleware,
	authCheckMiddleware middleware.AuthMiddleware,
) *MiddlewareContainers {
	middlewares := make([]middleware.Middleware, 0)
	middlewares = append(middlewares, upstreamCheckMiddleware) // Chain 순서는 upstreamCheckMiddleware 가 가장 먼저 실행되어야 한다.
	middlewares = append(middlewares, authCheckMiddleware)

	return &MiddlewareContainers{
		Middlewares: middlewares,
	}
}

var MiddlewareContainerSet = wire.NewSet(
	middleware.NewUpstreamCheckMiddleware,
	middleware.NewAuthMiddleware,
	NewMiddlewareContainers,
)
