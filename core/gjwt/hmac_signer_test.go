package gjwt

import (
	"testing"
	"time"
)

var hmacJWK = []byte(`{"kty":"oct","k":"c3VwZXItc2VjcmV0LWtleS1mb3ItaHMyNTY","alg":"HS256"}`)

func TestHMACCodecRoundTrip(t *testing.T) {
	const name = "hmac-roundtrip"
	if err := RegisterKey(name, hmacJWK, JSONKey, HS256.String()); err != nil {
		t.Fatalf("RegisterKey: %v", err)
	}
	codec, err := NewCodec(name)
	if err != nil {
		t.Fatalf("NewCodec: %v", err)
	}

	now := time.Now()
	token, err := codec.Serialize(nil, func(claims map[string]any) {
		claims[string(Subject)] = "user-1"
		claims[string(IssuedAt)] = now.Unix()
		claims[string(Expiration)] = now.Add(time.Hour).Unix()
	})
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	result := codec.Parse(token)
	if result.Err != nil {
		t.Fatalf("Parse: %v", result.Err)
	}
	if !result.Valid {
		t.Fatal("token should be valid")
	}
}

func BenchmarkHMACSerialize(b *testing.B) {
	const name = "hmac-bench-serialize"
	if err := RegisterKey(name, hmacJWK, JSONKey, HS256.String()); err != nil {
		b.Fatalf("RegisterKey: %v", err)
	}
	codec, err := NewCodec(name)
	if err != nil {
		b.Fatalf("NewCodec: %v", err)
	}
	now := time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = codec.Serialize(nil, func(claims map[string]any) {
			claims[string(Subject)] = "bench"
			claims[string(IssuedAt)] = now.Unix()
			claims[string(Expiration)] = now.Add(time.Hour).Unix()
		})
	}
}

func BenchmarkHMACDeserialize(b *testing.B) {
	const name = "hmac-bench-deserialize"
	if err := RegisterKey(name, hmacJWK, JSONKey, HS256.String()); err != nil {
		b.Fatalf("RegisterKey: %v", err)
	}
	codec, err := NewCodec(name)
	if err != nil {
		b.Fatalf("NewCodec: %v", err)
	}
	now := time.Now()
	token, err := codec.Serialize(nil, func(claims map[string]any) {
		claims[string(Subject)] = "bench"
		claims[string(IssuedAt)] = now.Unix()
		claims[string(Expiration)] = now.Add(time.Hour).Unix()
	})
	if err != nil {
		b.Fatalf("Serialize: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = codec.Parse(token)
	}
}
