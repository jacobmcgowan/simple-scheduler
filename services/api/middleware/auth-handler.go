package middleware

import (
	"crypto/rsa"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"net/http"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/jacobmcgowan/simple-scheduler/services/api/auth"
)

func AuthHandler(cache *auth.AuthCache, reqScopes []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := ctx.GetHeader("Authorization")
		if authHeader == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is missing"})
			ctx.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization format"})
			ctx.Abort()
			return
		}

		rawToken := authHeader[7:]
		if rawToken == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Access token is missing"})
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(rawToken, func(token *jwt.Token) (any, error) {
			return validateTokenSign(token, cache)
		}, jwt.WithValidMethods([]string{jwt.SigningMethodRS256.Alg()}))
		if err != nil {
			log.Printf("Error parsing access token: %s", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			ctx.Abort()
			return
		}
		if !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid access token"})
			ctx.Abort()
			return
		}

		issuer, err := token.Claims.GetIssuer()
		if err != nil {
			log.Printf("Error getting issuer from token: %s", err.Error())
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid issuer"})
			ctx.Abort()
			return
		}
		if issuer != cache.Issuer {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid issuer"})
			ctx.Abort()
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			log.Printf("Error getting claims from token")
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid claims"})
			ctx.Abort()
			return
		}

		scopes := strings.Split(claims["scope"].(string), " ")
		missingScopes := []string{}
		for _, reqScope := range reqScopes {
			if !slices.Contains(scopes, reqScope) {
				missingScopes = append(missingScopes, reqScope)
			}
		}

		if len(missingScopes) > 0 {
			ctx.JSON(http.StatusForbidden, gin.H{"error": "Missing required scopes", "missing_scopes": missingScopes})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func validateTokenSign(token *jwt.Token, cache *auth.AuthCache) (any, error) {
	kid, ok := token.Header["kid"].(string)
	if !ok {
		return nil, fmt.Errorf("kid header is missing")
	}

	key, ok := cache.GetKey(kid)
	if !ok {
		return nil, fmt.Errorf("JWK not found for access token")
	}

	pubKey, err := jwkToRsaPublicKey(key)
	if err != nil {
		return nil, fmt.Errorf("failed to convert JWK to RSA public key: %s", err.Error())
	}

	return pubKey, nil
}

func jwkToRsaPublicKey(jwk auth.Jwk) (*rsa.PublicKey, error) {
	if jwk.Kty != "RSA" {
		return nil, fmt.Errorf("unsupported key type: %s", jwk.Kty)
	}

	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %s", err.Error())
	}

	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %s", err.Error())
	}

	return &rsa.PublicKey{
		N: big.NewInt(0).SetBytes(nBytes),
		E: int(big.NewInt(0).SetBytes(eBytes).Int64()),
	}, nil
}
