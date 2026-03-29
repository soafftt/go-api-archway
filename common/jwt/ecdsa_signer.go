package jwt

import (
	"crypto/ecdsa"
	"errors"
	"fmt"

	goJwt "github.com/golang-jwt/jwt/v5"
)

type ecdsaKeyMap map[string]*ecdsa.PrivateKey
type ECDSACodec JwtCodec
type ECDSASinger struct {
	token *goJwt.Token
}

/*
ECDSASigner 는 thread-safe 하지 않아 동시성에 오류가 있습니다.
사용시 ECDSASigner 를 매번 생성해주세요.
*/
type ecdsaSigner struct {
	token  *goJwt.Token
	singer Signer
}

func NewECDSASignerFromJsonKey(jsonKey []byte, singer Signer, singerName string) (ECDSACodec, JwtError) {
	codec, err := newECDSASinger(jsonKey, singer, singerName, JSONKeyType)
	if err != nil {
		return ecdsaSigner{}, err
	}

	return codec, nil
}

func NewECDSASingerFomPem(pemKey []byte, singer Signer, singerName string) (ECDSACodec, JwtError) {
	codec, err := newECDSASinger(pemKey, singer, singerName, PEMKeyType)
	if err != nil {
		return ecdsaSigner{}, err
	}

	return codec, nil
}

func newECDSASinger(key []byte, signer Signer, singerName string, keyType KeyType) (ECDSACodec, JwtError) {
	jwtToken, err := newSigner(signer)
	if err != nil {
		return ecdsaSigner{}, ErrECDSATokenInstanceError
	}

	err = putJwtKeyAndSingerSignature(key, singerName, keyType)
	if err != nil {
		return ecdsaSigner{}, err
	}

	return ecdsaSigner{
		token:  jwtToken,
		singer: signer,
	}, nil
}

func (e ecdsaSigner) Serialize(singerName string, header HeadBuilder, claims ClaimsBuilder) JwtSerializeResult {
	exists := existsSignature(singerName)
	if !exists {
		return handleJwtErrorToJwtSerializeResult(ErrSerializeKeyNotFoundError)
	}

	cacheHeader := getHeaderFromSyncPool()
	cacheClaims := getClaimsFromSyncPool()

	defer deferSyncPool(cacheHeader, cacheClaims)

	h, err := header(cacheHeader)
	if err != nil {
		return handleJwtErrorToJwtSerializeResult(errors.Join(ErrSerializeError, err))
	}
	e.token.Header = h

	c, err := claims(cacheClaims)
	if err != nil {
		return handleJwtErrorToJwtSerializeResult(errors.Join(ErrSerializeError, err))
	}
	e.token.Claims = c

	jwk, exists := getJwtKey(singerName)
	if !exists {
		return handleJwtErrorToJwtSerializeResult(ErrSerializeKeyNotFoundError)
	}

	priKey, jwtError := convertPrivateKey[ecdsa.PrivateKey](jwk.privateKey)
	if jwtError != nil {
		return handleJwtErrorToJwtSerializeResult(jwtError)
	}

	singedString, err := e.token.SignedString(priKey)
	if err != nil {
		return handleErrorToJwtSerializeResult(err, ErrSerializeError)
	}

	return JwtSerializeResult{
		SignedJwt: singedString,
		Err:       nil,
	}
}

func (e ecdsaSigner) Deserialize(singerName string, token string) JwtDeserializeResult {
	jwk, exists := getJwtKey(singerName)
	if !exists {
		return handleJwtErrorToJwtDeserializeResult(ErrDeserializeError)
	}

	priKey, jwtError := convertPrivateKey[ecdsa.PrivateKey](jwk.privateKey)
	if jwtError != nil {
		return handleJwtErrorToJwtDeserializeResult(jwtError)
	}

	jwtToken, err := goJwt.Parse(
		token,
		func(token *goJwt.Token) (interface{}, error) {
			// 알고리즘 검증
			if _, ok := token.Method.(*goJwt.SigningMethodECDSA); !ok {
				return nil, errors.Join(ErrVerifyError, fmt.Errorf("unexpected signing method: %v", token.Header["alg"]))
			}

			ok := existsSignature(singerName)
			if !ok {
				return nil, ErrDeserializeKeyNotFoundError
			}

			return &priKey.PublicKey, nil
		},
	)

	if err != nil {
		return handleErrorToJwtDeserializeResult(err, ErrDeserializeKeyNotFoundError)
	}

	return JwtDeserializeResult{
		Header: jwtToken.Header,
		Claims: jwtToken.Claims.(goJwt.MapClaims),
		Verify: jwtToken.Valid,
		Err:    nil,
	}
}
