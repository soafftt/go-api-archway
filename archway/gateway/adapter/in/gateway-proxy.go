package in

import (
	"context"
	coreAdapterIn "core/adapter/in"
	"core/consts/httpheader"
	"core/utils"
	"encoding/json"
	"gateway/adapter/in/ctxkey"
	"gateway/application/port/in"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"sync"
	"time"
)

var logger = utils.GetLogger()

const proxyBufferSize = 10 * 10 * 1024

type httpResponseBufferPool struct {
	bufferPool sync.Pool
}

func newProxyBufferPool() httputil.BufferPool {
	return &httpResponseBufferPool{
		bufferPool: sync.Pool{
			New: func() any {
				return new(make([]byte, proxyBufferSize))
			},
		},
	}
}

func (p *httpResponseBufferPool) Get() []byte {
	buffer := p.bufferPool.Get().(*[]byte)
	return *buffer
}

func (p *httpResponseBufferPool) Put(bytes []byte) {
	bytes = bytes[:0]
	clear(bytes)
	p.bufferPool.Put(&bytes)
}

type GatewayProxy struct {
	HttpProxy *httputil.ReverseProxy
}

func NewGatewayProxy() *GatewayProxy {
	return &GatewayProxy{
		HttpProxy: newReversProxy(),
	}
}

func newReversProxy() *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		// rewire 설정 (설정하지 않으면 기본값 사용)
		Transport: &http.Transport{
			// 외부로 나가는 방법
			// FromEnvironment 는 기본 설정값에 따름
			// google -> 엇 외부니까 NAT 타고 나가야지?
			// localhost --> 엇? 내부네.. ㅎ
			// OS 설정에 따른다는 것이 아래의 뜻이고.
			// Cluster 내의 POD 끼리 할려면 nil 로 처리.
			// 이게 가장 안정적일듯
			Proxy: http.ProxyFromEnvironment,

			// 네트워크 timeout 재처리
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.DialTimeout(network, addr, time.Second*10)
			},
			// keep-alive 를 끌껏인가? (굳이....)
			DisableKeepAlives: false,
			// http/2 로 강제 변환 전송 할것인가
			ForceAttemptHTTP2: false,
			// Pool 에 가지고 있어도 되는 최대 커넥션 수
			MaxIdleConns:        500,
			MaxIdleConnsPerHost: 500,
			// Pool 에 얼마나 머무를 것인가.
			IdleConnTimeout: 120 * time.Second,
			// ssl handshake timeout
			TLSHandshakeTimeout: 1 * time.Second,
			// httpStatus 100 을 기다리는 시간.(upstream 이 status 100 을 구현해야 함)
			ExpectContinueTimeout: 1 * time.Second,
		},
		// rewrite
		Rewrite: func(proxy *httputil.ProxyRequest) {
			proxyRewrite(proxy)
		},
		// 응답 변경 = body 는 upstream 응답을 그대로 사용.
		ModifyResponse: func(res *http.Response) error {
			return proxyModifyResponse(res)
		},
		// 에러처리.
		ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
			target := r.Context().Value(ctxkey.LOOKUP_KEY).(in.UpstreamLookupResult)
			// 에러로깅.
			// 알수 없는 에러가 발생하는 경우.
			// 이렇기에 에러 응답을 맞춰야 우 싰음
			logger.ErrorW("ProxyError:["+target.Method+"]"+target.FullUrl(), err)

			w.WriteHeader(http.StatusInternalServerError)
			errResponse := coreAdapterIn.NewErrorResponse(err)

			// json encoding error
			if err := json.NewEncoder(w).Encode(errResponse); err != nil {
				logger.ErrorW("ErrorResponse jsonEncode:["+target.Method+"]"+target.FullUrl(), err)
			}

		},
		// bufferPool 을 syncPool 로 사용하게 하여, 메모리 최적회 (1MB) --> 줄여도 됨.
		BufferPool: newProxyBufferPool(),
	}
}

func proxyRewrite(proxy *httputil.ProxyRequest) {
	ctx := proxy.In.Context()
	lookupResult := ctx.Value(ctxkey.LOOKUP_KEY).(in.UpstreamLookupResult)

	proxy.Out.URL = &url.URL{
		Scheme:   lookupResult.Scheme(),
		Host:     lookupResult.GetDomain(),
		Path:     lookupResult.Path,
		RawQuery: proxy.In.URL.RawQuery,
	}

	xForwardedFor := proxy.In.Header.Get(httpheader.XForwardedFor)
	if xForwardedFor != "" {
		xForwardedFor = proxy.In.RemoteAddr
	} else {
		xForwardedFor = xForwardedFor + ";" + proxy.In.Host
	}

	proxy.Out.Method = lookupResult.Method
	proxy.Out.Body = proxy.In.Body

	// 요청 헤더를 그대로 복사한다.
	proxy.Out.Header = proxy.In.Header.Clone()
	proxy.Out.Header.Set(httpheader.XRequestHost, proxy.In.Host)
	proxy.Out.Header.Set(httpheader.XForwardedFor, xForwardedFor)
	// 인증체크 정보가 있으면 X-USER 로 전달한다.
	if lookupResult.UserKey != nil {
		proxy.Out.Header.Set(httpheader.XUser, lookupResult.UserKey.(string))
	}
}

func proxyModifyResponse(res *http.Response) error {
	// 에러이면 Cache 자체를 사용할 일이 없음.
	if res.StatusCode != http.StatusOK {
		return nil
	}

	lookupResult := res.Request.Context().Value(ctxkey.LOOKUP_KEY).(in.UpstreamLookupResult)
	if lookupResult.CacheTimeout > 0 {
		res.Header.Set(httpheader.CacheControl, "max-age="+strconv.FormatInt(lookupResult.CacheTimeout, 10))
	} else {
		res.Header.Set(httpheader.CacheControl, "max-age=-1")
	}

	return nil
}
