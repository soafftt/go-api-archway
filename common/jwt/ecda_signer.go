package jwt

import (
	// 필수 값.

	"crypto/ecdsa"
	"errors"

	goJwt "github.com/golang-jwt/jwt/v5"
)

type ecdaKeyMap map[string]*ecdsa.PrivateKey
type ECDSACodec JwtCodec
type ECDSASinger struct {
	token *goJwt.Token
}

/*
ECDASigner 는 thread-safe 하지 않아 동시성에 오류가 있습니다.
사용시 ECDASigner 를 매번 생성해주세요.
*/
type ecdasSigner struct {
	token  *goJwt.Token
	singer Signer
}

func NewECDASigerFromJsonKey(jsonKey []byte, singer Signer, singerName string) (ECDSACodec, JwtError) {
	jwtToken, err := newJwtSigner(singer)
	if err != nil {
		return ecdasSigner{}, JwtECDATokenInstanceError
	}

	putJwtKeyAndSingerName(jsonKey, singerName, JsonKey)

	return ecdasSigner{
		token:  jwtToken,
		singer: singer,
	}, nil
}

func (e ecdasSigner) Serialize(singerName string, header HeadBuilder, claims ClaimsBuilder) JwtSerializeResult {
	ok := existsSingerName(singerName)
	if !ok {
		return JwtSerializeResult{
			Err: JwtSerializeKeyNotFoundError,
		}
	}

	cacheHeader := headerSyncPool.Get().(map[string]any)
	cacheClaims := claimsSyncPool.Get().(goJwt.MapClaims)

	defer func() {
		clear(cacheHeader)
		clear(cacheClaims)

		headerSyncPool.Put(cacheHeader)
		claimsSyncPool.Put(cacheClaims)
	}()

	e.token.Header, _ = header(cacheHeader)
	e.token.Claims, _ = claims(cacheClaims)

	jwk, ok := getJwtKey(singerName)
	if !ok {
		// TODO 에러 코드 정리.
		return JwtSerializeResult{
			Err: JwtSerializeKeyNotFoundError,
		}
	}
	priKey := convertPrivateKey[ecdsa.PrivateKey](jwk.privateKey)

	singedString, err := e.token.SignedString(priKey)
	if err != nil {
		return JwtSerializeResult{
			Err: errors.Join(JwtSerializeError, err),
		}
	}

	return JwtSerializeResult{
		SignedJwt: singedString,
		Err:       nil,
	}
}

func (e ecdasSigner) Deserialize(singerName string, token string) JwtDeserializeResult {
	jwk, ok := getJwtKey(singerName)
	if !ok {
		// TODO 에러 코드 정리.
		return JwtDeserializeResult{
			Err: JwtDeserializeError,
		}
	}

	priKey := convertPrivateKey[ecdsa.PrivateKey](jwk.privateKey)

	jwtToken, err := goJwt.Parse(
		token,
		func(token *goJwt.Token) (interface{}, error) {
			ok := existsSingerName(singerName)
			if !ok {
				return nil, JwtDeserializeKeyNotFoundError
			}

			return &priKey.PublicKey, nil
		},
	)

	if err != nil {
		if errors.Is(err, JwtDeserializeKeyNotFoundError) {
			return JwtDeserializeResult{
				Err: err,
			}
		}

		return JwtDeserializeResult{
			Err: errors.Join(JwtDeserializeError, err),
		}
	}

	return JwtDeserializeResult{
		Header: jwtToken.Header,
		Claims: jwtToken.Claims.(goJwt.MapClaims),
		Verify: jwtToken.Valid,
		Err:    nil,
	}
}
