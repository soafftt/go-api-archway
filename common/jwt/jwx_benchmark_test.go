package jwt

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"

	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwt"
)

func BenchmarkJWXSign(b *testing.B) {
	privKey, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		t := jwt.New()
		t.Set(jwt.SubjectKey, "perf")
		t.Set("i", i)

		_, _ = jwt.Sign(t, jwt.WithKey(jwa.ES256(), privKey))
	}
}
