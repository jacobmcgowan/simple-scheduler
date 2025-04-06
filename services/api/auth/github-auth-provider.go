package auth

import (
	"os"

	"github.com/gin-gonic/gin"
	envVars "github.com/jacobmcgowan/simple-scheduler/shared/resources/env-vars"
	"github.com/zalando/gin-oauth2/github"
)

func registerGitHubAuth(router *gin.Engine) {
	credsFile := "./oauth2-local-creds.json"
	redirectUrl := os.Getenv(envVars.Oauth2RedirectUrl)
	secret := []byte(os.Getenv(envVars.Oauth2Secret))
	sessionName := "simple-scheduler-session"
	scopes := []string{
		"repo",
	}

	github.Setup(redirectUrl, credsFile, scopes, secret)
	router.Use(github.Session(sessionName))
	router.GET("/api/login", github.LoginHandler)
}
