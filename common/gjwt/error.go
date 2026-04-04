package gjwt

type JwtError string

func (e JwtError) Error() string {
	return string(e)
}

var (
	ErrKeyParse    JwtError = "failed to parse key"
	ErrKeyExport   JwtError = "failed to export key"
	ErrKeyNotFound JwtError = "key not found in key store"
	ErrPrivateKey  JwtError = "invalid or unsupported private key"
	ErrPublicKey   JwtError = "invalid or unsupported public key"
	ErrAlgNotFound JwtError = "unsupported signing algorithm"
	ErrSerialize   JwtError = "failed to sign JWT"
	ErrDeserialize JwtError = "failed to parse JWT"
)
