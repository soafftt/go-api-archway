package jwt

import (
	"crypto"
	"errors"
	"sync"

	goJwt "github.com/golang-jwt/jwt/v5"
	"github.com/lestrrat-go/jwx/v3/jwk"
)

type JwtError error

var (
	JwtKeyParseError               JwtError = errors.New("failed to parse JSON key")
	JwtKeyExportError              JwtError = errors.New("failed to export JSON key")
	JwtTokenInstanceError          JwtError = errors.New("failed to create JWT instance")
	JwtECDATokenInstanceError      JwtError = errors.New("failed to create ECDSA JWT instance")
	JwtPrivateKeyConvertError      JwtError = errors.New("failed to convert private key")
	JwtSerializeKeyNotFoundError   JwtError = errors.New("failed to find key for JWT serialization")
	JwtDeserializeKeyNotFoundError JwtError = errors.New("failed to find key for JWT deserialization")
	JwtSerializeError              JwtError = errors.New("failed to serialize JWT")
	JwtDeserializeError            JwtError = errors.New("failed to deserialize JWT")
	JwtVerifyError                 JwtError = errors.New("failed to verify JWT")
)

type Signer string

const (
	RSA256Signer   Signer = "RSA256"
	RSA512Signer   Signer = "RSA512"
	ECDSA256Signer Signer = "ECDSA256"
	ECDSA512Signer Signer = "ECDSA512"
	HMACSigner     Signer = "HMAC"
)

type KeyType string

const (
	PemKey  KeyType = "PEM"
	JsonKey KeyType = "JSON"
)

type DefaultClaims string

const (
	// Jwt Isuer
	Issuer DefaultClaims = "iss"
	// Jwt SubJect
	Subject DefaultClaims = "sub"
	// jwt Audience
	Audience DefaultClaims = "aud"
	// jwt Expiration
	Expiration DefaultClaims = "exp"
	// jwt NotBefore
	NotBefore DefaultClaims = "nbf"
	// jwt IssuedAt
	IssuedAt DefaultClaims = "iat"
)

type JwkKey struct {
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
}

type jwtKeyMap map[string]JwkKey

func NewJwtKey(privateKey, publicKey crypto.PrivateKey) JwkKey {
	return JwkKey{
		privateKey: privateKey,
		publicKey:  publicKey,
	}
}

var (
	jwtKeyMapLock    sync.RWMutex
	jwtPrivateKeyMap = make(jwtKeyMap, 50)
)

func makeCryptoPrivateKeyFromJwk(jsonKey []byte, pem bool) (JwkKey, error) {
	jwkPrivateKey, err := jwk.ParseKey(jsonKey, jwk.WithPEM(pem))
	if err != nil {
		return JwkKey{}, errors.Join(JwtKeyParseError, err)
	}

	var privateKey crypto.PrivateKey
	if err := jwk.Export(jwkPrivateKey, &privateKey); err != nil {
		return JwkKey{}, errors.Join(JwtKeyExportError, err)
	}

	signer, ok := privateKey.(crypto.Signer)
	if !ok {
		// TODO: 에러 코드 정리
		return JwkKey{}, JwtPrivateKeyConvertError
	}

	return NewJwtKey(privateKey, signer.Public()), nil
}

func putJwtKeyAndSingerName(key []byte, singerName string, keyType KeyType) JwtError {
	_, existsJwtKey := jwtPrivateKeyMap[singerName]
	_, existsSingerName := singerNameMap[singerName]

	if !existsJwtKey || !existsSingerName {
		var jwkKey JwkKey
		if !existsJwtKey {
			var err error
			jwkKey, err = makeCryptoPrivateKeyFromJwk(key, keyType == PemKey)
			if err != nil {
				return err
			}
		}

		jwtKeyMapLock.Lock()
		defer jwtKeyMapLock.Unlock()

		if !existsSingerName {
			singerNameMap[singerName] = ecdsaSingerName{}
		}

		if !existsJwtKey {
			jwtPrivateKeyMap[singerName] = jwkKey
		}
	}

	return nil
}

func getJwtKey(singerName string) (JwkKey, bool) {
	jwtKeyMapLock.RLock()
	defer jwtKeyMapLock.RUnlock()

	cryptoPrivateKey, ok := jwtPrivateKeyMap[singerName]
	if !ok {
		return JwkKey{}, false
	}

	return cryptoPrivateKey, ok
}

type ecdsaSingerName struct{}
type ecdsaSingerNameMap map[string]ecdsaSingerName

var (
	ecdsaSingerNameMapLock sync.RWMutex
	singerNameMap          = make(ecdsaSingerNameMap, 50)
)

func existsSingerName(name string) bool {
	ecdsaSingerNameMapLock.RLock()
	defer ecdsaSingerNameMapLock.RUnlock()

	_, ok := singerNameMap[name]

	return ok
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

type HeadBuilder func(map[string]any) (map[string]any, error)
type ClaimsBuilder func(goJwt.MapClaims) (goJwt.MapClaims, error)

type JwtSerializeResult struct {
	SignedJwt string
	Err       error
}

type JwtDeserializeResult struct {
	Header map[string]any
	Claims goJwt.MapClaims
	Verify bool
	Err    JwtError
}

type JwtCodec interface {
	Serialize(singerName string, header HeadBuilder, claims ClaimsBuilder) JwtSerializeResult
	Deserialize(singerName string, token string) JwtDeserializeResult
}

/*
goJwt(github.com/golang-jwt/jwt/v5) 는 thread-safe 하지 핞다.
Token 만들거나, Token Verify 등을 할때는 매번 새로운 객체를 만들어야 한다.
*/
type jwtSigner struct {
	Token *goJwt.Token
}

func newJwtSigner(singer Signer) (*goJwt.Token, JwtError) {
	jwtToken, jwtError := singer.makeJwtToken()
	if jwtError != nil {
		return nil, jwtError
	}

	return jwtToken, nil
}

func convertPrivateKey[T any](privateKey crypto.PrivateKey) *T {
	// c := privKey.(*ecdsa.PrivateKey)
	return privateKey.(*T)
}

func convertPublicKey[T any](publicKey crypto.PublicKey) (interface{}, JwtError) {
	convertKey, ok := publicKey.(*T)
	if !ok {
		return nil, JwtKeyExportError
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
		return nil, JwtPrivateKeyConvertError
	}
}
