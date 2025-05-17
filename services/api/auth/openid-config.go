package auth

type OpenIdConfig struct {
	Issuer  string `json:"issuer"`
	JwksUri string `json:"jwks_uri"`
}
