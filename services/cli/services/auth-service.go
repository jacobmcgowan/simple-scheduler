package services

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	authProviderTypes "github.com/jacobmcgowan/simple-scheduler/shared/auth/auth-provider-types"
	"golang.org/x/oauth2"
)

const CliUrl = "http://localhost:5556"
const CacheFileName = ".simple-scheduler-cli-cache"

type AuthService struct {
	ApiUrl       string
	oidcConfig   *oidc.Config
	oidcVerifier *oidc.IDTokenVerifier
	oauth2Config oauth2.Config
}

func (svc *AuthService) Init(ctx context.Context, clientId string, clientSecret string, providerType authProviderTypes.AuthProviderType) error {
	providerUrl, err := getProviderUrl(providerType)
	if err != nil {
		return fmt.Errorf("failed to get provider url: %s", err.Error())
	}

	provider, err := oidc.NewProvider(ctx, providerUrl)
	if err != nil {
		return err
	}

	svc.oidcConfig = &oidc.Config{
		ClientID: clientId,
	}
	svc.oidcVerifier = provider.Verifier(svc.oidcConfig)

	svc.oauth2Config = oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/callback", CliUrl),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	http.HandleFunc(fmt.Sprintf("%s/login", CliUrl), svc.handleLogin)
	http.HandleFunc(fmt.Sprintf("%s/callback", CliUrl), svc.handleCallback)
	return http.ListenAndServe(CliUrl, nil)
}

func (svc *AuthService) Login() error {
	resp, err := http.Get(fmt.Sprintf("%s/login", CliUrl))
	if err != nil {
		return fmt.Errorf("failed to login: %s", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to login: %s", resp.Status)
	}

	return nil
}

func (svc *AuthService) handleLogin(resWrtr http.ResponseWriter, req *http.Request) {
	http.Redirect(resWrtr, req, svc.oauth2Config.AuthCodeURL("state"), http.StatusFound)
}

func (svc *AuthService) handleCallback(resWrtr http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	code := req.URL.Query().Get("code")
	if code == "" {
		http.Error(resWrtr, "authorization code is required", http.StatusBadRequest)
		return
	}

	oauth2Token, err := svc.oauth2Config.Exchange(ctx, code)
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to exchange token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	rawIdToken, ok := oauth2Token.Extra("id_token").(string)
	if !ok {
		http.Error(resWrtr, "id_token is missing from oauth2 token", http.StatusInternalServerError)
		return
	}

	idToken, err := svc.oidcVerifier.Verify(ctx, rawIdToken)
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to verify id token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	nonce, err := req.Cookie("nonce")
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("nonce cookie is missing: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	if nonce.Value != idToken.Nonce {
		http.Error(resWrtr, "nonce does not match", http.StatusBadRequest)
		return
	}

	svc.saveAccessToken(oauth2Token.AccessToken)
	resWrtr.WriteHeader(http.StatusNoContent)
}

func (svc *AuthService) saveAccessToken(accessToken string) error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %s", err.Error())
	}

	filePath := fmt.Sprintf("%s/%s", CacheFileName, dir)
	err = os.WriteFile(filePath, []byte(accessToken), 0644)
	if err != nil {
		return fmt.Errorf("failed to cache the access token: %s", err.Error())
	}

	return nil
}

func (svc *AuthService) loadAccessToken() (string, error) {
	dir, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("failed to get user home directory: %s", err.Error())
	}

	filePath := fmt.Sprintf("%s/%s", CacheFileName, dir)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read the cache: %s", err.Error())
	}

	return string(data), nil
}

func getProviderUrl(providerType authProviderTypes.AuthProviderType) (string, error) {
	switch providerType {
	case authProviderTypes.GitHub:
		return "https://github.com/login/device/code", nil
	default:
		return "", fmt.Errorf("unsupported provider type: %s", providerType)
	}
}
