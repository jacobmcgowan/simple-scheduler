package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

type AuthCache struct {
	Issuer     string
	keys       map[string]Jwk
	cachedLock sync.Mutex `default:"sync.Mutex{}"`
	cached     bool
}

func (cache *AuthCache) GetKey(kid string) (Jwk, bool) {
	cache.cachedLock.Lock()
	defer cache.cachedLock.Unlock()

	if !cache.cached {
		err := cache.loadKeys()
		if err != nil {
			return Jwk{}, false
		}

		cache.cached = true
	}

	cert, ok := cache.keys[kid]
	return cert, ok
}

func (cache *AuthCache) loadKeys() error {
	config, err := cache.getOpenIdConfig()
	if err != nil {
		return fmt.Errorf("failed to get OpenID configuration: %s", err.Error())
	}

	jwks, err := cache.getJwks(config)
	if err != nil {
		return fmt.Errorf("failed to get JWK certs: %s", err.Error())
	}

	cache.keys = make(map[string]Jwk)
	for _, cert := range jwks.Keys {
		cache.keys[cert.Kid] = cert
	}

	return nil
}

func (cache *AuthCache) getOpenIdConfig() (OpenIdConfig, error) {
	config := OpenIdConfig{}
	url := fmt.Sprintf("%s/.well-known/openid-configuration", cache.Issuer)
	resp, err := http.Get(url)
	if err != nil {
		return config, fmt.Errorf("failed to get OpenID configuration: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return config, fmt.Errorf("failed to get OpenID configuration: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return config, fmt.Errorf("failed to read OpenID configuration: %s", err.Error())
	}

	err = json.Unmarshal(body, &config)
	if err != nil {
		return config, fmt.Errorf("failed to parse OpenID configuration: %s", err.Error())
	}

	return config, nil
}

func (cache *AuthCache) getJwks(config OpenIdConfig) (Jwks, error) {
	certs := Jwks{}
	resp, err := http.Get(config.JwksUri)
	if err != nil {
		return certs, fmt.Errorf("failed to get certs: %s", err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return certs, fmt.Errorf("failed to get certs: %s", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return certs, fmt.Errorf("failed to read certs: %s", err.Error())
	}

	err = json.Unmarshal(body, &certs)
	if err != nil {
		return certs, fmt.Errorf("failed to parse certs: %s", err.Error())
	}

	return certs, nil
}
