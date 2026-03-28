package jwt

import (
	"encoding/base64"
	"maps"
	"strconv"
	"testing"

	goJwt "github.com/golang-jwt/jwt/v5"
)

const ecdsaJWKBase64 = "eyJrdHkiOiJFQyIsImQiOiJfVXQtTUhUeGI5RG1NSUhBRTlUNmxWdklES3BlRGFZeHN0M05iUVE2a3BRIiwiY3J2IjoiUC0yNTYiLCJraWQiOiIzWWptSk53MjRrQ1BRT0M0aFJJYUU1ZkcwQmtTOHZvblNzWjN3ZW8yWVhFIiwieCI6IllQbmctaUlWY1R1M1ppYTFkNGdaZWtpM0ZsUzNvM2J4eUwtb2RUclNpUDAiLCJ5IjoiTmlPS1hqQWJPS2tTTGNyYlVia1dnUVQ3VG5YYjN2UXhlYTdhV2pvdW5SQSJ9"

var decodeJwk, _ = base64.StdEncoding.DecodeString(ecdsaJWKBase64)

func TestNewECDASigerFromJsonKey(t *testing.T) {
	singerName := "test-singer"
	var name = singerName

	ecdsaCodec, err := NewECDASigerFromJsonKey(decodeJwk, ECDSA256Signer, name)
	if err != nil {
		t.Fatalf("Failed to create ECDASigner: %v", err)
	}

	_, ok := ecdsaCodec.(ECDSACodec)
	if !ok {
		t.Fatal("Failed to assert type to ECDSACodec")
	}
}

func TestECDSSinterSeialize(t *testing.T) {
	var singerName string = "ttt"
	var name = singerName

	ecdsaCodec, err := NewECDASigerFromJsonKey(decodeJwk, ECDSA256Signer, name)
	if err != nil {
		t.Fatalf("Failed to create ECDASigner: %v", err)
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

func BenchmarkTestNewECDASigerFromJsonKey(b *testing.B) {
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
		codec, err := NewECDASigerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, singerNameMap)
		if err != nil {
			b.Fatalf("Failed to create ECDASigner: %v", err)
		}

		_, ok := codec.(ECDSACodec)
		if !ok {
			b.Fatal("Failed to assert type to ECDSACodec")
		}
	}
}

// 매번 PrivateKey 를 갱신하기 때문에 느리고 alloc 가 높음..
func BenchmarkTestColdmNewECDASigerFromJsonKey(b *testing.B) {
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
		codec, err := NewECDASigerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, singerNameMap)
		if err != nil {
			b.Fatalf("Failed to create ECDASigner: %v", err)
		}

		_, ok := codec.(ECDSACodec)
		if !ok {
			b.Fatal("Failed to assert type to ECDSACodec")
		}
	}
}

// 한번 만든 ECDASigner 를 재사용하기 때문에 빠르고 alloc 가 낮음..
func BenchmarkTestHotNewECDASigerFromJsonKey(b *testing.B) {
	lookCount := b.N

	name := "warmup-singer"

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < lookCount; i++ {
		codec, err := NewECDASigerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, name)
		if err != nil {
			b.Fatalf("Failed to create ECDASigner: %v", err)
		}

		_, ok := codec.(ECDSACodec)
		if !ok {
			b.Fatal("Failed to assert type to ECDSACodec")
		}
	}
}

// ECDSA 의 경우, public key 를 매번 갱신하기 때문에 Serialize 할 때마다 느리고 alloc 가 높음..
func BenchmarkECDASignerSerialize(b *testing.B) {

	loopCount := b.N
	testSingerMap := make(map[int]ECDSACodec, loopCount)
	testCodecNameMap := make(map[int]string, loopCount)

	for i := 0; i < loopCount; i++ {
		var singerName string = strconv.Itoa(i)
		codec, err := NewECDASigerFromJsonKey(
			decodeJwk,
			ECDSA256Signer, singerName)

		if err != nil {
			b.Fatalf("Failed to create ECDASigner: %v", err)
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

func BenchmarkECDASignerDeserialize(b *testing.B) {
	var singerName string = ""

	codec, err := NewECDASigerFromJsonKey(decodeJwk, ECDSA256Signer, singerName)
	if err != nil {
		b.Fatalf("Failed to create ECDASigner: %v", err)
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

func TestClaim(t *testing.T) {
	type testClaim goJwt.MapClaims

	type testClaims map[string]testClaim

	ttt := make(testClaims, 100)

	for i := 0; i < 100; i++ {
		key := strconv.Itoa(i)

		ttt[key] = make(testClaim, 10)

		value := ttt[key]
		value1 := maps.Clone(value)
		value1["test"] = "test"

		value2 := ttt[key]

		println(value2)
	}
}

func BenchmarkTestMapsClone(b *testing.B) {
	loopCnt := b.N
	testMap := make(map[int]map[string]string, loopCnt)

	for i := 0; i < loopCnt; i++ {
		testMap[i] = make(map[string]string, 100)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < loopCnt; i++ {
		original, _ := testMap[i]

		t := maps.Clone(original)
		t["test"] = "test"
		t["test2"] = "test"
		t["test"] = "test"

	}

}

// func BenchmarkNewECDASigerFromJsonKey(b *testing.B) {
// 	jsonKey, err := base64.StdEncoding.DecodeString(ecdsaJWKBase64)
// 	if err != nil {
// 		b.Fatalf("Failed to decode base64 key: %v", err)
// 	}

// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		_, err := NewECDASigerFromJsonKey(string(jsonKey), ECDSA256Signer, "bench-signer")
// 		if err != nil {
// 			b.Fatalf("Failed to create ECDASigner: %v", err)
// 		}
// 	}
// }

// func BenchmarkECDASignerSerialize(b *testing.B) {
// 	jsonKey, err := base64.StdEncoding.DecodeString(ecdsaJWKBase64)
// 	if err != nil {
// 		b.Fatalf("Failed to decode base64 key: %v", err)
// 	}

// 	codec, err := NewECDASigerFromJsonKey(string(jsonKey), ECDSA256Signer, "serialize-bench-signer")
// 	if err != nil {
// 		b.Fatalf("Failed to create ECDASigner: %v", err)
// 	}

// 	b.ReportAllocs()
// 	b.ResetTimer()
// 	for i := 0; i < b.N; i++ {
// 		result := codec.Serialize(
// 			func() (map[string]any, error) {
// 				return map[string]any{"kid": "bench"}, nil
// 			},
// 			func() (goJwt.MapClaims, error) {
// 				return goJwt.MapClaims{"sub": "perf", "i": i}, nil
// 			},
// 		)
// 		if result.Err != nil {
// 			b.Fatalf("Failed to serialize JWT: %v", result.Err)
// 		}
// 	}
// }
