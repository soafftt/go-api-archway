package gjwt

const (
	algorithm_ES256   = "ES256"
	algorithm_ES512   = "ES512"
	algorithm_RS256   = "RS256"
	algorithm_RS512   = "RS512"
	algorithm_HS256   = "HS256"
	algorithm_HS512   = "HS512"
	algorithm_UNKNOWN = "UNKNOWN"
)

type Algorithm string

const (
	ES256   Algorithm = algorithm_ES256
	ES512   Algorithm = algorithm_ES512
	RS256   Algorithm = algorithm_RS256
	RS512   Algorithm = algorithm_RS512
	HS256   Algorithm = algorithm_HS256
	HS512   Algorithm = algorithm_HS512
	UNKNOWN Algorithm = algorithm_UNKNOWN
)

func (al Algorithm) String() string {
	switch al {
	case ES256:
		return algorithm_ES256
	case ES512:
		return algorithm_ES512
	case RS256:
		return algorithm_RS256
	case RS512:
		return algorithm_RS512
	case HS256:
		return algorithm_HS256
	case HS512:
		return algorithm_HS512
	default:
		return algorithm_UNKNOWN
	}
}

func GetAlgorithm(alg string) Algorithm {
	switch alg {
	case algorithm_ES256:
		return ES256
	case algorithm_ES512:
		return ES512
	case algorithm_RS256:
		return RS256
	case algorithm_RS512:
		return RS512
	case algorithm_HS256:
		return HS256
	case algorithm_HS512:
		return HS512
	default:
		return UNKNOWN
	}
}

// DefaultClaims 는 RFC 7519 표준 JWT 클레임 키를 정의한다.
type DefaultClaims string

const (
	Issuer     DefaultClaims = "iss"
	Subject    DefaultClaims = "sub"
	Audience   DefaultClaims = "aud"
	Expiration DefaultClaims = "exp"
	NotBefore  DefaultClaims = "nbf"
	IssuedAt   DefaultClaims = "iat"
)
