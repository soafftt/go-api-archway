package gjwt

import (
	"encoding/base64"
	"testing"
	"time"
)

const rsaJWKBase64 = "eyJkIjoiS3V3SjFjbWRzc1pEMjkxTDVxdk5wMVZQTzMySHg4Q0hEekY1dzhyd1BYVjJvYWtVU3ZUckdhZnFzM0M1UkJkUUtZemZ1WGpfU0EwdGxjOEpObmhSQWhVd25pdG00QWg5dVZtRHlLZGtVZ1FCS0ZNVVQyWkRTdTJaT2daVEQ4OW1GcnNZX1JXSEJRMU1BcHlwbDd5ZDAyY3ZzV2dtbG9yeEVZTHNESGs0NGJ1cEkzMk9FZG05eWJjVmFlX2owTFktb3J1Q3IxNEVmdm1XWGFreXFOd204bXNkRkRYZkQ1TURpXy1MQURHNFVFa18wR29jeldlbEZQYUFCbHNhQ1BNRVUycnNEVEwzWnVUM1FCcmNCNzZaUVFEM1BFRlZCQS1FQTNtQnE4RlBrRVFqTF9IY3pUY1FYU2pfbGtsQ1pWd2ZiX21lQzNzaW5LU2w3TG5WVHo0b1F3IiwiZHAiOiJ0TXBUXzlJcVZDb2I2U1k4YUJrU3IyTWwtVXlPQk1TY29wSGRuWExsQ2ZDckVpdXo1YWdLUDVwTzhmNHktbW1nbHE0OEVpanZjY1cyUERHb3lvRzVnN280blIzM3ByOGFmcDlrQWx1QmZmWWpVZXA2WjBFdkVlaEo1RmtmUlA2UVFxbTV3M2ZVcGRPU19wb2RvTDE5bUkzTlVHRTltenpLVm9lamREel92YWMiLCJkcSI6IlpiSXJVdW5HZ2FEQW5rQXhLa0tObUpKdWp3LTJUc0Q5Sk0xUmxIdmZ1XzVxajdRS1dsdEJ4OTBxOWJHOWpXeDJnSHJCeWpPZVdWcXMxZEdUMVhOblVscUpaNWNHTjh4dXpXQmhBLXdjMUhwazhuNXpBNUg3RTlMbzdqa1puOXpaX1IyTEMwNUljS1ppTXBVaHBFUktsR0ZSUGZNb3RRU2cwV002b2laenFRayIsImUiOiJBUUFCIiwia3R5IjoiUlNBIiwibiI6IngxcU4wU0NfT3E1cVh6YjZCWDNuNk1JLWVxem5NZnFwcFVkcTBhWkpMUTB6U0RZSER0OGl4OUp1dGszQTVaOUpiZXJIT3JIdjlvQnNubmZRUnJZRHNLODdyN09hU0dVeHpyRm1ZLTZqRzg4dWZ3VWpBOEdfVl9HaVNYeTQ2VFpZRlpNb2J4UHVjdFJwNmhQeFBUNXQ5a19td2FIYnZ1Vm5zNmYyNnZOVmVrbko2YjhpLWlrbmFxR255VEhNSFNHUmtqX0FuVXlxbGF6cEFZdDZhSGZCQ2lWVllBUERzUUM3ZzYwQ01nNkNPX2hHVDFqRWRZY1c3VDZ6T1FKejE5cURIY0JkODE3dVVZQTEzR2tVaUoycEVJQUM5OTZWOWdKMFoyZGFNa0VVS0wxUlZTZk1mNm9uNURUMkswR0pibEpKVXludER1MUJDMTdBY1RxQ2FvbkdhUSIsInAiOiIzWHFaU0s5WUg0Y2FxajQ1NG9BMlA1X3Z3RTliUFFWRUNkbV9RRnd5T3d5a2tpMEtLbTBpS1NDdWVWbjdvUV92WXdrNWt2S09ZRWJJejMyaU12VGRXMEp0TlJFd0RfVTExQ1dXUGtyOE0tUGpPeDhoUjlTVVlQS0p0d0FGcUNRTWl4dnFMMGJua2I5VnptdDlKUlkxMDBfVk1RSlNJTHZEb0EzbjhzdHhXbGMiLCJxIjoiNW0wZkZfRTJ6SHRkUjBZT3Y5alh4ampycUFzWDRpOHhBNGJXRURFdFI0aUpmVFZlNXM2bU5idU9LenZBLVp3Z2l6ZGE3SGhPQUc4VU9EVEQ2WXcwQmRLVFhkdUQ3LUV2Mkh1ZERRblFOZ1N2a213SFBYZjRIdjdOajRIQVNHdmFzOWYtOGdjTEFEekc1ejlOZ29OVjI4Z2M0ZWt3UVlzQ1NKeTYxYUhPN1Q4IiwicWkiOiJneWVJbWpCMW5KRFE2ZWpmaGJLRGtReFJ1MjlKRC1RemJaSWdZVlkzOHYtN3pyMTl5TDRBSTVCOUMzSzF2VzJibVkwR1VENXRlNUdsQ2lDbHZ6NmpIVzBrRlNIUmY5V3VnSkVSVDQxQ081X25GZ1FMaUxnUXB4VGV3MjEyRVlfajB1aGJIMkhMRnp2blFNcUtNTlcyYmFUMlhFYy10dEFDaE1pb0Q1T0pXalEifQ=="

var rsaJWK, _ = base64.StdEncoding.DecodeString(rsaJWKBase64)

func TestRegisterRSAKey(t *testing.T) {
	if err := RegisterKey("rsa-test", rsaJWK, JSONKey); err != nil {
		t.Fatalf("RegisterKey failed: %v", err)
	}
	if !HasKey("rsa-test") {
		t.Fatal("key not found after registration")
	}
}

func TestRSACodecSerialize(t *testing.T) {
	const name = "rsa-serialize"
	c, err := NewCodec(name, rsaJWK, JSONKey, RS256)
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

func TestRSACodecParse(t *testing.T) {
	const name = "rsa-roundtrip"
	c, err := NewCodec(name, rsaJWK, JSONKey, RS256)
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

func BenchmarkRSASerialize(b *testing.B) {
	const name = "rsa-bench-serialize"
	c, err := NewCodec(name, rsaJWK, JSONKey, RS256)
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

func BenchmarkRSADeserialize(b *testing.B) {
	const name = "rsa-bench-deserialize"
	c, err := NewCodec(name, rsaJWK, JSONKey, RS256)
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
