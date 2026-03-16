package main

import (
	"context"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os/signal"
	"syscall"
)

func main() {

	proxy := &httputil.ReverseProxy{
		Rewrite: func(pr *httputil.ProxyRequest) {
			target, _ := url.Parse("https://www.naver.com")
			// 여기서, target 을 확인하고.
			// context write 함.
			// session check 가 필요하면 추가. 하고.
			// header 에 write 해줌.

			pr.Out.WithContext(context.WithValue(pr.Out.Context(), "A", "B"))

			// 1. URL과 Host 헤더를 한 번에 표준에 맞게 수정
			pr.SetURL(target)
		},
		ModifyResponse: func(res *http.Response) error {
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
