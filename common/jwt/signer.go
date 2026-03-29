package jwt

import (
	"crypto"
	"errors"
	"sync"

	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

// JwtToken 생성, 검증을 위한 Signer 타입을 재정의 합니다.
type Signer string

// Singer 타입은 JWT 토큰을 생성할 때 사용할 알고리즘을 나타냅니다.
const (
	RSA256Signer   Signer = "RSA256"
	RSA512Signer   Signer = "RSA512"
	ECDSA256Signer Signer = "ECDSA256"
	ECDSA512Signer Signer = "ECDSA512"
	HMACSigner     Signer = "HMAC"
)

// KeyType Jwt.Token 을 생성할때 키 타입을 나타냅니다.
type KeyType string

// 지원 하는 KeyType 종류
const (
	PEMKeyType  KeyType = "PEM"
	JSONKeyType KeyType = "JSON"
)

// Jwt Claims 에서 사용할 수 있는 기본 클레임을 정의 합니다.
type DefaultClaims string

const (
	// Jwt Issuer
	Issuer DefaultClaims = "iss"
	// Jwt Subject
	Subject DefaultClaims = "sub"
	// Jwt Audience
	Audience DefaultClaims = "aud"
	// Jwt Expiration
	Expiration DefaultClaims = "exp"
	// Jwt NotBefore
	NotBefore DefaultClaims = "nbf"
	// Jwt IssuedAt
	IssuedAt DefaultClaims = "iat"
)

// Jwt 의 RSA/ ECDSA 키를 저장하는 구조체와, Signer 이름을 저장하는 구조체를 정의 합니다.
type jwtKey struct {
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
}

// JwtCodec 인터페이스는 JWT 토큰을 직렬화(Serialize)하고 역직렬화(Deserialize)하는 메서드를 정의합니다.
// 재사용의 목적으로 map 으로 관리 하며 읽기/쓰기 시 동시성 문제를 방지하기 위해 sync.RWMutex 를 사용합니다.
type jwtKeyMap map[string]jwtKey

// JwtKey 객체를 생성합니다.
func newJwtKey(privateKey crypto.PrivateKey, publicKey crypto.PublicKey) jwtKey {
	return jwtKey{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

type singerSignatureMap map[string]struct{}

var (
	jwtLock          sync.RWMutex
	signatureNameMap = make(singerSignatureMap, 50)
	jwtKeysMap       = make(jwtKeyMap, 50)
)

func existsSignature(name string) bool {
	jwtLock.RLock()
	defer jwtLock.RUnlock()

	_, exists := signatureNameMap[name]

	return exists
}

func makeCryptoKeyFromJwk(jsonKey []byte, pem bool) (jwtKey, JwtError) {
	jwkPrivateKey, err := jwk.ParseKey(jsonKey, jwk.WithPEM(pem))
	if err != nil {
		return jwtKey{}, errors.Join(ErrKeyParseError, err)
	}

	var privateKey crypto.PrivateKey
	if err := jwk.Export(jwkPrivateKey, &privateKey); err != nil {
		return jwtKey{}, errors.Join(ErrKeyExportError, err)
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		// TODO: 에러 코드 정리
		return jwtKey{}, ErrCryptoPrivateKeyParseError
	}

	return newJwtKey(privateKey, signer.Public()), nil
}

func putJwtKeyAndSingerSignature(key []byte, singerName string, keyType KeyType) JwtError {
	jwtLock.RLock()
	_, existsJwtKey := jwtKeysMap[singerName]
	_, existsSingerName := signatureNameMap[singerName]
	jwtLock.RUnlock()

	if !existsJwtKey || !existsSingerName {
		var jwkKey jwtKey
		if !existsJwtKey {
			var err JwtError
			jwkKey, err = makeCryptoKeyFromJwk(key, keyType == PEMKeyType)
			if err != nil {
				return err
			}
		}

		jwtLock.Lock()
		defer jwtLock.Unlock()

		if !existsSingerName {
			if _, exists := signatureNameMap[singerName]; !exists {
				signatureNameMap[singerName] = struct{}{}
			}
		}

		if !existsJwtKey {
			if _, exists := jwtKeysMap[singerName]; !exists {
				jwtKeysMap[singerName] = jwkKey
			}
		}
	}

	return nil
}

func existsJwtKey(singerName string) bool {
	jwtLock.RLock()
	defer jwtLock.RUnlock()

	_, existsJwtKey := jwtKeysMap[singerName]
	return existsJwtKey
}

func getJwtKey(singerName string) (jwtKey, bool) {
	jwtLock.RLock()
	defer jwtLock.RUnlock()

	cryptoPrivateKey, exists := jwtKeysMap[singerName]
	if !exists {
		return jwtKey{}, false
	}

	return cryptoPrivateKey, exists
}

var (
	headerSyncPool = sync.Pool{
		New: func() any {
			return make(map[string]any, 50)
		},
	}

	claimsSyncPool = sync.Pool{
		New: func() any {
			return make(goJwt.MapClaims, 50)
		},
	}
)

func getHeaderFromSyncPool() map[string]any {
	return headerSyncPool.Get().(map[string]any)
}

func getClaimsFromSyncPool() goJwt.MapClaims {
	return claimsSyncPool.Get().(goJwt.MapClaims)
}

func deferSyncPool(header map[string]any, claims goJwt.MapClaims) {
	clear(header)
	clear(claims)

	headerSyncPool.Put(header)
	claimsSyncPool.Put(claims)
}

type HeadBuilder func(map[string]any) (map[string]any, error)
type ClaimsBuilder func(goJwt.MapClaims) (goJwt.MapClaims, error)

type JwtSerializeResult struct {
	SignedJwt string
	Err       error
}

func handleErrorToJwtSerializeResult(err error, jwtError JwtError) JwtSerializeResult {
	return JwtSerializeResult{
		Err: errors.Join(jwtError, err),
	}
}

func handleJwtErrorToJwtSerializeResult(jwtError JwtError) JwtSerializeResult {
	return JwtSerializeResult{
		Err: jwtError,
	}
}

type JwtDeserializeResult struct {
	Header map[string]any
	Claims goJwt.MapClaims
	Verify bool
	Err    JwtError
}

func handleErrorToJwtDeserializeResult(err error, jwtError JwtError) JwtDeserializeResult {
	var retError error
	if errors.Is(err, jwtError) {
		retError = err
	} else {
		retError = errors.Join(jwtError, err)
	}

	return JwtDeserializeResult{
		Err: retError,
	}
}

func handleJwtErrorToJwtDeserializeResult(jwtError JwtError) JwtDeserializeResult {
	return JwtDeserializeResult{
		Err: jwtError,
	}
}

type JwtCodec interface {
	Serialize(singerName string, header HeadBuilder, claims ClaimsBuilder) JwtSerializeResult
	Deserialize(singerName string, token string) JwtDeserializeResult
}

/*
goJwt(github.com/golang-jwt/jwt/v5) 는 thread-safe 하지 핞다.
Token 만들거나, Token Verify 등을 할때는 매번 새로운 객체를 만들어야 한다.
*/
type signer struct {
	Token *goJwt.Token
}

func newSigner(singer Signer) (*goJwt.Token, JwtError) {
	jwtToken, jwtError := singer.makeJwtToken()
	if jwtError != nil {
		return nil, jwtError
	}

	return jwtToken, nil
}

func convertPrivateKey[T any](privateKey crypto.PrivateKey) (*T, JwtError) {
	// c := privKey.(*ecdsa.PrivateKey)
	convertKey, ok := privateKey.(*T)
	if !ok {
		return nil, ErrPrivateKeyConvertError
	}
	return convertKey, nil
}

func convertPublicKey[T any](publicKey crypto.PublicKey) (*T, JwtError) {
	convertKey, ok := publicKey.(*T)
	if !ok {
		return nil, ErrPublicKeyConvertError
	}
	return convertKey, nil
}

func (s Signer) makeJwtToken() (*goJwt.Token, JwtError) {
	switch s {
	case RSA256Signer, RSA512Signer:
		method := goJwt.SigningMethodRS256
		if s == RSA512Signer {
			method = goJwt.SigningMethodRS512
		}

		return goJwt.New(method), nil

	case ECDSA256Signer, ECDSA512Signer:
		method := goJwt.SigningMethodES256
		if s == ECDSA512Signer {
			method = goJwt.SigningMethodES512
		}
		return goJwt.New(method), nil

	default:
		// Todo: 에러 코드 정의
		return nil, ErrTokenInstanceError
	}
}
