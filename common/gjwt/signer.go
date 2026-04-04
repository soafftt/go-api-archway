package gjwt

import (
	"errors"
	"sync"

	goJwt "github.com/golang-jwt/jwt/v5"
)

type (
	// HeaderBuilder 는 JWT 헤더 필드를 직접 채우는 함수 타입이다.
	HeaderBuilder func(map[string]any)
	// ClaimsBuilder 는 JWT 클레임 필드를 직접 채우는 함수 타입이다.
	ClaimsBuilder func(map[string]any)
)

// Codec 은 고정된 키와 알고리즘으로 JWT 서명 및 파싱을 담당한다.
// 동시성에 안전하며 요청 간 재사용을 전제로 설계되었다.
type Codec interface {
	// Serialize 는 JWT를 생성하고 서명한다. header, claims 빌더는 nil 가능.
	Serialize(header HeaderBuilder, claims ClaimsBuilder) (string, error)
	// Parse 는 JWT 문자열을 검증하고 파싱한다.
	Parse(tokenString string) ParseResult
}

type codec struct {
	keyStoreName string
	method       goJwt.SigningMethod
	keyEntry     keyEntry
}

// NewCodec 은 키 데이터를 등록하고 Codec 을 생성한다.
// 동일한 keyStoreName 이 이미 등록되어 있으면 keyData 는 무시되며 기존 키를 재사용한다.
func NewCodec(keyStoreName string) (Codec, error) {
	keyEntry, ok := GetKey(keyStoreName)
	if !ok {
		return nil, ErrKeyNotFound
	}

	method, err := signingMethod(keyEntry.Algorithm)
	if err != nil {
		return nil, err
	}
	return &codec{keyStoreName: keyStoreName, method: method, keyEntry: keyEntry}, nil
}

func signingMethod(alg Algorithm) (goJwt.SigningMethod, error) {
	switch alg {
	case ES256:
		return goJwt.SigningMethodES256, nil
	case ES512:
		return goJwt.SigningMethodES512, nil
	case RS256:
		return goJwt.SigningMethodRS256, nil
	case RS512:
		return goJwt.SigningMethodRS512, nil
	case HS256:
		return goJwt.SigningMethodHS256, nil
	case HS512:
		return goJwt.SigningMethodHS512, nil
	default:
		return nil, ErrAlgNotFound
	}
}

// headerPool, claimsPool 은 서명 핫패스의 map 할당 비용을 줄이기 위한 풀이다.
var (
	headerPool = sync.Pool{New: func() any { m := make(map[string]any, 4); return &m }}
	claimsPool = sync.Pool{New: func() any { m := make(map[string]any, 8); return &m }}
)

func (c *codec) Serialize(header HeaderBuilder, claims ClaimsBuilder) (string, error) {
	hp := headerPool.Get().(*map[string]any)
	cp := claimsPool.Get().(*map[string]any)
	defer func() {
		clear(*hp)
		clear(*cp)
		headerPool.Put(hp)
		claimsPool.Put(cp)
	}()

	if header != nil {
		header(*hp)
	}
	if claims != nil {
		claims(*cp)
	}

	token := goJwt.NewWithClaims(c.method, goJwt.MapClaims(*cp))
	for k, v := range *hp {
		token.Header[k] = v
	}

	signed, err := token.SignedString(c.keyEntry.PrivateKey)
	if err != nil {
		return "", errors.Join(ErrSerialize, err)
	}
	return signed, nil
}

func (c *codec) Parse(tokenString string) ParseResult {
	return Parse(c.keyStoreName, tokenString)
}
