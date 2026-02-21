package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"strings"
	"syscall"
)

func main() {

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			target, _ := url.Parse("https://www.naver.com")

			// 1. URL과 Host 헤더를 한 번에 표준에 맞게 수정
			pr.SetURL(target)
		},
		ModifyResponse: func(res *http.Response) error {
			if res.StatusCode == http.StatusMovedPermanently {
				// 타겟 서버가 보낸 내부 주소를 외부 도메인 주소로 교체
				location := res.Header.Get("Location")
				newLocation := strings.Replace(location, "http://10.0.1.10:8080", "https://naver.com", 1)
				res.Header.Set("Location", newLocation)
			}
			return nil
		},
	}

	go func() {
		func() {
			if r := recover(); r != nil {
				log.Printf("panic occurred: %v", r)
			}
		}()

		if err := http.ListenAndServe(":8083", proxy); err != nil {
			log.Fatalf("서버 실행 실패: %v", err)
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	<-ctx.Done()
	log.Println("서버 종료")
}
