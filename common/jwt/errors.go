package jwt

import "errors"

type JwtError error

var (
	ErrKeyParseError               JwtError = errors.New("failed to parse JSON key")
	ErrKeyExportError              JwtError = errors.New("failed to export JSON key")
	ErrPrivateKeyParseError        JwtError = errors.New("failed to parse private key")
	ErrCryptoPrivateKeyParseError  JwtError = errors.New("failed to parse crypto private key")
	ErrPublicKeyParseError         JwtError = errors.New("failed to parse public key")
	ErrTokenInstanceError          JwtError = errors.New("failed to create JWT instance")
	ErrECDSATokenInstanceError     JwtError = errors.New("failed to create ECDSA JWT instance")
	ErrPrivateKeyConvertError      JwtError = errors.New("failed to convert private key")
	ErrPublicKeyConvertError       JwtError = errors.New("failed to convert public key")
	ErrSerializeKeyNotFoundError   JwtError = errors.New("failed to find key for JWT serialization")
	ErrDeserializeKeyNotFoundError JwtError = errors.New("failed to find key for JWT deserialization")
	ErrSerializeError              JwtError = errors.New("failed to serialize JWT")
	ErrDeserializeError            JwtError = errors.New("failed to deserialize JWT")
	ErrVerifyError                 JwtError = errors.New("failed to verify JWT")
)
