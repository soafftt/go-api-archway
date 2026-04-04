package gjwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"sync"

	goJwa "github.com/lestrrat-go/jwx/v3/jwa"
	goJwk "github.com/lestrrat-go/jwx/v3/jwk"
)

// KeyType 은 원시 키 바이트의 인코딩 형식을 나타낸다.
type KeyType string

const (
	PEMKey  KeyType = "PEM"
	JSONKey KeyType = "JSON"
)

type keyEntry struct {
	Algorithm  Algorithm
	PrivateKey crypto.PrivateKey
	PublicKey  crypto.PublicKey
}

var (
	keyMu    sync.RWMutex
	keyStore = make(map[string]keyEntry)
)

// RegisterKey 는 키를 파싱하여 주어진 이름으로 저장한다.
// 이미 등록된 이름이면 파싱을 건너뛰므로 멱등성(idempotent)을 보장한다.
func RegisterKeyByString(name, data string, keyType KeyType, algorithm string) error {
	decodedKey, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return ErrKeyParse
	}

	if err := RegisterKey(name, decodedKey, keyType, algorithm); err != nil {
		return err
	}

	return nil
}

func RegisterKey(name string, data []byte, keyType KeyType, algorithm string) error {
	keyMu.RLock()
	_, exists := keyStore[name]

	keyMu.RUnlock()
	if exists {
		return nil
	}

	jwkKey, err := goJwk.ParseKey(data, goJwk.WithPEM(keyType == PEMKey))
	if err != nil {
		return errors.Join(ErrKeyParse, err)
	}

	entry, err := exportKeyEntry(algorithm, jwkKey)
	if err != nil {
		return err
	}

	keyMu.Lock()
	defer keyMu.Unlock()

	if _, exists := keyStore[name]; !exists {
		keyStore[name] = entry
	}
	return nil
}

// exportKeyEntry 는 jwk.Key 를 구체 타입으로 변환하여 keyEntry 를 반환한다.
func exportKeyEntry(algorithm string, jwkKey goJwk.Key) (keyEntry, error) {
	alg := GetAlgorithm(algorithm)
	if alg == algorithm_UNKNOWN {
		return keyEntry{}, ErrAlgNotFound
	}

	switch jwkKey.KeyType() {
	case goJwa.EC():
		var pk ecdsa.PrivateKey
		if err := goJwk.Export(jwkKey, &pk); err != nil {
			return keyEntry{}, errors.Join(ErrKeyExport, err)
		}
		return keyEntry{PrivateKey: &pk, PublicKey: &pk.PublicKey}, nil

	case goJwa.RSA():
		var pk rsa.PrivateKey
		if err := goJwk.Export(jwkKey, &pk); err != nil {
			return keyEntry{}, errors.Join(ErrKeyExport, err)
		}
		return keyEntry{PrivateKey: &pk, PublicKey: &pk.PublicKey}, nil

	case goJwa.OctetSeq():
		// HMAC: 대칭키는 raw bytes 로 저장
		var raw []byte
		if err := goJwk.Export(jwkKey, &raw); err != nil {
			return keyEntry{}, errors.Join(ErrKeyExport, err)
		}
		return keyEntry{PrivateKey: raw, PublicKey: raw}, nil

	default:
		return keyEntry{}, ErrPrivateKey
	}
}

// HasKey 는 주어진 이름으로 키가 등록되어 있는지 여부를 반환한다.
func HasKey(name string) bool {
	keyMu.RLock()
	defer keyMu.RUnlock()
	_, ok := keyStore[name]
	return ok
}

func getKey(name string) (keyEntry, bool) {
	keyMu.RLock()
	defer keyMu.RUnlock()
	entry, ok := keyStore[name]
	return entry, ok
}
