package jwt

import (
	"encoding/base64"
	"strconv"
	"testing"

	goJwt "github.com/golang-jwt/jwt/v5"
)

const ecdsaJWKBase64 = "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"

var decodeJwk, _ = base64.StdEncoding.DecodeString(ecdsaJWKBase64)

func TestNewECDSASignerFromJsonKey(t *testing.T) {
	signerName := "test-signer"
	var name = signerName

	ecdsaCodec, err := NewECDSASignerFromJsonKey(decodeJwk, ECDSA256Signer, name)
	if err != nil {
		t.Fatalf("Failed to create ECDSASigner: %v", err)
	}

	if ecdsaCodec == nil {
		t.Fatal("ECDSASigner is nil")
	}
}

func TestECDSASignerSerialize(t *testing.T) {
	var singerName string = "ttt"
	var name = singerName

	ecdsaCodec, err := NewECDSASignerFromJsonKey(decodeJwk, ECDSA256Signer, name)
	if err != nil {
		t.Fatalf("Failed to create ECDSASigner: %v", err)
	}

	headerBuilder := func(mh map[string]any) (map[string]any, error) {
		mh["kid"] = "bench"
		return mh, nil
	}

	claimsBuilder := func(mc goJwt.MapClaims) (goJwt.MapClaims, error) {
		mc["sub"] = "perf"
		mc["i"] = "1"
		return mc, nil
	}

	for i := 0; i < 100; i++ {
		result := ecdsaCodec.Serialize(name, headerBuilder, claimsBuilder)

		if result.Err != nil {
			t.Fatalf("Failed to serialize JWT: %v", result.Err)
		}

		if result.SignedJwt == "" {
			t.Fatal("Signed JWT is empty")
		}
	}
}

func BenchmarkTestNewECDSASignerFromJsonKey(b *testing.B) {
	lookCount := b.N

	names := make(map[int]string, lookCount)
	for i := 0; i < lookCount; i++ {
		var singerName string = strconv.Itoa(i)
		names[i] = singerName
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		singerNameMap, _ := names[i]
		ecdsaCodec, err := NewECDSASignerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, singerNameMap)
		if err != nil {
			b.Fatalf("Failed to create ECDSASigner: %v", err)
		}

		if ecdsaCodec == nil {
			b.Fatal("ECDSASigner is nil")
		}
	}
}

// 매번 PrivateKey 를 갱신하기 때문에 느리고 alloc 가 높음..
func BenchmarkTestColdNewECDSASignerFromJsonKey(b *testing.B) {
	lookCount := b.N

	names := make(map[int]string, lookCount)
	for i := 0; i < lookCount; i++ {
		var singerName string = strconv.Itoa(i)
		names[i] = singerName
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		singerNameMap, _ := names[i]
		ecdsaCodec, err := NewECDSASignerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, singerNameMap)
		if err != nil {
			b.Fatalf("Failed to create ECDSASigner: %v", err)
		}

		if ecdsaCodec != nil {
			b.Fatal("Failed to assert type to ECDSACodec")
		}
	}
}

// 한번 만든 ECDSASigner 를 재사용하기 때문에 빠르고 alloc 가 낮음..
func BenchmarkTestHotNewECDSASignerFromJsonKey(b *testing.B) {
	lookCount := b.N

	name := "warmup-singer"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < lookCount; i++ {
		codec, err := NewECDSASignerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, name)
		if err != nil {
			b.Fatalf("Failed to create ECDSASigner: %v", err)
		}
		if codec == nil {
			b.Fatal("codec is nil")
		}
	}
}

// ECDSA 의 경우, public key 를 매번 갱신하기 때문에 Serialize 할 때마다 느리고 alloc 가 높음..
func BenchmarkECDSASignerSerialize(b *testing.B) {

	loopCount := b.N
	testSingerMap := make(map[int]ECDSACodec, loopCount)
	testCodecNameMap := make(map[int]string, loopCount)

	for i := 0; i < loopCount; i++ {
		var singerName string = strconv.Itoa(i)
		codec, err := NewECDSASignerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, singerName)

		if err != nil {
			b.Fatalf("Failed to create ECDSASigner: %v", err)
		}

		testSingerMap[i] = codec
		testCodecNameMap[i] = singerName
	}

	headerBuilder := func(mh map[string]any) (map[string]any, error) {
		mh["kid"] = "bench"
		return mh, nil
	}

	claimsBuilder := func(mc goJwt.MapClaims) (goJwt.MapClaims, error) {
		mc["sub"] = "perf"
		mc["i"] = "1"
		return mc, nil
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < loopCount; i++ {
		name, _ := testCodecNameMap[i]
		codec, _ := testSingerMap[i]
		codec.Serialize(
			name,
			headerBuilder,
			claimsBuilder,
		)
	}
}

func BenchmarkECDSASignerDeserialize(b *testing.B) {
	var singerName string = ""

	codec, err := NewECDSASignerFromJsonKey(decodeJwk, ECDSA256Signer, singerName)
	if err != nil {
		b.Fatalf("Failed to create ECDSASigner: %v", err)
	}

	result := codec.Serialize(
		singerName,
		func(mh map[string]any) (map[string]any, error) {
			mh["kid"] = "bench"
			return mh, nil
		},
		func(mc goJwt.MapClaims) (goJwt.MapClaims, error) {
			mc["sub"] = "perf"
			mc["i"] = "1"
			return mc, nil
		},
	)

	if result.Err != nil {
		b.Fatalf("Failed to serialize JWT: %v", result.Err)
	}

	signedJwt := result.SignedJwt

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		codec.Deserialize(singerName, signedJwt)
	}
}
