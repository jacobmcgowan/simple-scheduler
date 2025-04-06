package auth

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	authProviderTypes "github.com/jacobmcgowan/simple-scheduler/shared/auth/auth-provider-types"
	envVars "github.com/jacobmcgowan/simple-scheduler/shared/resources/env-vars"
)

func RegisterAuth(router *gin.Engine) error {
	providerType := os.Getenv(envVars.Oauth2ProviderType)

	switch providerType {
	case string(authProviderTypes.GitHub):
		registerGitHubAuth(router)
		return nil
	default:
		return fmt.Errorf("OAuth2 provider type %s not supported", providerType)
	}
}
