package jwt

type Algorithm string

const (
	ES256 Algorithm = "ES256"
	ES512 Algorithm = "ES512"
	RS256 Algorithm = "RS256"
	RS512 Algorithm = "RS512"
	HS256 Algorithm = "HS256"
	HS512 Algorithm = "HS512"
)

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
