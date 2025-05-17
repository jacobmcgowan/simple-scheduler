package auth

type Jwks struct {
	Keys []Jwk `json:"keys"`
}
