package gjwt

import (
	"encoding/base64"
	"testing"
	"time"
)

const ecdsaJWKBase64 = "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"

var ecdsaJWK, _ = base64.StdEncoding.DecodeString(ecdsaJWKBase64)

func TestRegisterECDSAKey(t *testing.T) {
	if err := RegisterKey("ecdsa-test", ecdsaJWK, JSONKey); err != nil {
		t.Fatalf("RegisterKey failed: %v", err)
	}
	if !HasKey("ecdsa-test") {
		t.Fatal("key not found after registration")
	}
}

func TestECDSACodecSerialize(t *testing.T) {
	const name = "ecdsa-serialize"
	c, err := NewCodec(name, ecdsaJWK, JSONKey, ES256)
	if err != nil {
		t.Fatalf("NewCodec: %v", err)
	}

	now := time.Now()
	token, err := c.Serialize(
		func(h map[string]any) { h["kid"] = name },
		func(cl map[string]any) {
			cl[string(Subject)] = "user-1"
			cl[string(IssuedAt)] = now.Unix()
			cl[string(Expiration)] = now.Add(time.Hour).Unix()
		},
	)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}
	if token == "" {
		t.Fatal("signed token is empty")
	}
}

func TestECDSACodecParse(t *testing.T) {
	const name = "ecdsa-roundtrip"
	c, err := NewCodec(name, ecdsaJWK, JSONKey, ES256)
	if err != nil {
		t.Fatalf("NewCodec: %v", err)
	}

	now := time.Now()
	token, err := c.Serialize(
		nil,
		func(cl map[string]any) {
			cl[string(Subject)] = "user-1"
			cl[string(IssuedAt)] = now.Unix()
			cl[string(Expiration)] = now.Add(time.Hour).Unix()
		},
	)
	if err != nil {
		t.Fatalf("Serialize: %v", err)
	}

	result := c.Parse(token)
	if result.Err != nil {
		t.Fatalf("Parse: %v", result.Err)
	}
	if !result.Valid {
		t.Fatal("token should be valid")
	}
	if result.Claims[string(Subject)] != "user-1" {
		t.Fatalf("unexpected subject: %v", result.Claims[string(Subject)])
	}
}

func BenchmarkECDSARegisterKey(b *testing.B) {
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Idempotent: measures the fast-path (key already cached).
		_ = RegisterKey("ecdsa-bench-hot", ecdsaJWK, JSONKey)
	}
}

func BenchmarkECDSASerialize(b *testing.B) {
	const name = "ecdsa-bench-serialize"
	c, err := NewCodec(name, ecdsaJWK, JSONKey, ES256)
	if err != nil {
		b.Fatalf("NewCodec: %v", err)
	}
	now := time.Now()

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = c.Serialize(
			func(h map[string]any) { h["kid"] = name },
			func(cl map[string]any) {
				cl[string(Subject)] = "bench"
				cl[string(IssuedAt)] = now.Unix()
				cl[string(Expiration)] = now.Add(time.Hour).Unix()
			},
		)
	}
}

func BenchmarkECDSADeserialize(b *testing.B) {
	const name = "ecdsa-bench-deserialize"
	c, err := NewCodec(name, ecdsaJWK, JSONKey, ES256)
	if err != nil {
		b.Fatalf("NewCodec: %v", err)
	}
	now := time.Now()
	token, err := c.Serialize(
		nil,
		func(cl map[string]any) {
			cl[string(Subject)] = "bench"
			cl[string(IssuedAt)] = now.Unix()
			cl[string(Expiration)] = now.Add(time.Hour).Unix()
		},
	)
	if err != nil {
		b.Fatalf("Serialize: %v", err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = c.Parse(token)
	}
}
