package jwt

import (
	"errors"
	"time"

	goJwt "github.com/golang-jwt/jwt/v5"
)

const leewayDuration = time.Second * 30

var parsers = map[Algorithm]*goJwt.Parser{
	ES256: goJwt.NewParser(
		goJwt.WithValidMethods([]string{"ES256"}),
		goJwt.WithExpirationRequired(),
		goJwt.WithIssuedAt(),
		goJwt.WithLeeway(leewayDuration),
	),
	ES512: goJwt.NewParser(
		goJwt.WithValidMethods([]string{"ES512"}),
		goJwt.WithExpirationRequired(),
		goJwt.WithIssuedAt(),
		goJwt.WithLeeway(leewayDuration),
	),
	RS256: goJwt.NewParser(
		goJwt.WithValidMethods([]string{"RS256"}),
		goJwt.WithExpirationRequired(),
		goJwt.WithIssuedAt(),
		goJwt.WithLeeway(leewayDuration),
	),
	RS512: goJwt.NewParser(
		goJwt.WithValidMethods([]string{"RS512"}),
		goJwt.WithExpirationRequired(),
		goJwt.WithIssuedAt(),
		goJwt.WithLeeway(leewayDuration),
	),
	HS256: goJwt.NewParser(
		goJwt.WithValidMethods([]string{"HS256"}),
		goJwt.WithExpirationRequired(),
		goJwt.WithIssuedAt(),
		goJwt.WithLeeway(leewayDuration),
	),
	HS512: goJwt.NewParser(
		goJwt.WithValidMethods([]string{"HS512"}),
		goJwt.WithExpirationRequired(),
		goJwt.WithIssuedAt(),
		goJwt.WithLeeway(leewayDuration),
	),
}

// ParseResult 는 JWT 파싱 작업의 결과를 담는 구조체다.
type ParseResult struct {
	Header map[string]any
	Claims map[string]any
	Valid  bool
	Err    error
}

// Parse 는 keyStoreName 에 등록된 공개키와 alg 에 대응하는 파서로 tokenString 을 검증하고 파싱한다.
func Parse(keyStoreName string, alg Algorithm, tokenString string) ParseResult {
	entry, ok := getKey(keyStoreName)
	if !ok {
		return ParseResult{Err: ErrKeyNotFound}
	}

	p, ok := parsers[alg]
	if !ok {
		return ParseResult{Err: ErrAlgNotFound}
	}

	token, err := p.ParseWithClaims(
		tokenString,
		goJwt.MapClaims{},
		func(_ *goJwt.Token) (any, error) { return entry.PublicKey, nil },
	)
	if err != nil {
		return ParseResult{Err: errors.Join(ErrDeserialize, err)}
	}

	claims, _ := token.Claims.(goJwt.MapClaims)
	return ParseResult{
		Header: token.Header,
		Claims: map[string]any(claims),
		Valid:  token.Valid,
	}
}
