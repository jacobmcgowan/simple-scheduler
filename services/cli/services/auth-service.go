package services

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

const CliAddr = "localhost:5556"
const CliUrl = "http://localhost:5556"
const CacheFileName = ".simple-scheduler-cli-cache"

type AuthService struct {
	ApiUrl       string
	oauth2Config oauth2.Config
	verifier     *oidc.IDTokenVerifier
	server       http.Server
	loggedId     chan struct{}
}

func (svc *AuthService) Start(ctx context.Context, clientId string, clientSecret string, issuer string, wg *sync.WaitGroup) error {
	svc.loggedId = make(chan struct{})

	provider, err := oidc.NewProvider(ctx, issuer)
	if err != nil {
		return fmt.Errorf("failed to create provider: %s", err.Error())
	}
	oidcConfig := &oidc.Config{
		ClientID: clientId,
	}
	svc.verifier = provider.Verifier(oidcConfig)

	svc.oauth2Config = oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/callback", CliUrl),
		Endpoint:     provider.Endpoint(),
		Scopes: []string{
			oidc.ScopeOpenID,
			"jobs:read",
			"jobs:write",
			"runs:read",
			"runs:write",
		},
	}

	handler := http.NewServeMux()
	handler.HandleFunc("/login", svc.handleLogin)
	handler.HandleFunc("/callback", svc.handleCallback)

	svc.server = http.Server{
		Addr:    CliAddr,
		Handler: handler,
	}

	go svc.listenAndServe(wg)
	return nil
}

func (svc *AuthService) Stop() error {
	if err := svc.server.Shutdown(context.Background()); err != nil {
		return fmt.Errorf("failed to shutdown server: %s", err.Error())
	}

	return nil
}

func (svc *AuthService) Login() error {
	url := fmt.Sprintf("%s/login", CliUrl)
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default:
		if isWsl() {
			cmd = "cmd.exe"
			args = []string{"/c", "start", url}
		} else {
			cmd = "xdg-open"
			args = []string{url}
		}
	}

	if err := exec.Command(cmd, args...).Start(); err != nil {
		return fmt.Errorf("failed to open browser: %s", err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for {
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				return fmt.Errorf("login timed out")
			}
		case <-svc.loggedId:
			return nil
		}
	}
}

func (svc *AuthService) GetAccessToken() (string, error) {
	accessToken, err := svc.loadAccessToken()
	if err != nil {
		return "", fmt.Errorf("failed to load access token: %s", err.Error())
	}

	return accessToken, nil
}

func (svc *AuthService) listenAndServe(wg *sync.WaitGroup) error {
	wg.Add(1)
	defer wg.Done()

	if err := svc.server.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return fmt.Errorf("failed to start server: %s", err.Error())
	}

	return nil
}

func (svc *AuthService) handleLogin(wrtr http.ResponseWriter, req *http.Request) {
	state, err := randString()
	if err != nil {
		http.Error(wrtr, fmt.Sprintf("failed to generate state: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	setCookie(wrtr, req, "state", state)

	nonce, err := randString()
	if err != nil {
		http.Error(wrtr, fmt.Sprintf("failed to generate nonce: %s", err.Error()), http.StatusInternalServerError)
		return
	}
	setCookie(wrtr, req, "nonce", nonce)

	http.Redirect(wrtr, req, svc.oauth2Config.AuthCodeURL(state, oidc.Nonce(nonce)), http.StatusFound)
}

func (svc *AuthService) handleCallback(resWrtr http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	state, err := req.Cookie("state")
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to get state cookie: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if req.URL.Query().Get("state") != state.Value {
		http.Error(resWrtr, "state did not match", http.StatusBadRequest)
		return
	}

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
		http.Error(resWrtr, "id_token not found in token", http.StatusInternalServerError)
		return
	}
	idToken, err := svc.verifier.Verify(ctx, rawIdToken)
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to verify id_token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	nonce, err := req.Cookie("nonce")
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to get nonce cookie: %s", err.Error()), http.StatusBadRequest)
		return
	}
	if idToken.Nonce != nonce.Value {
		http.Error(resWrtr, "nonce did not match", http.StatusBadRequest)
		return
	}

	if err = svc.saveAccessToken(oauth2Token.AccessToken); err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to save access token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	resWrtr.WriteHeader(http.StatusNoContent)
	svc.loggedId <- struct{}{}
}

func (svc *AuthService) saveAccessToken(accessToken string) error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %s", err.Error())
	}

	filePath := fmt.Sprintf("%s/%s", dir, CacheFileName)
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

	filePath := fmt.Sprintf("%s/%s", dir, CacheFileName)
	data, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read the cache: %s", err.Error())
	}

	return string(data), nil
}

func AddAuthHeader(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

func randString() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random state: %s", err.Error())
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func setCookie(wrtr http.ResponseWriter, req *http.Request, name string, value string) {
	cookie := &http.Cookie{
		Name:     name,
		Value:    value,
		MaxAge:   int(time.Hour.Seconds()),
		Secure:   req.TLS != nil,
		HttpOnly: true,
	}
	http.SetCookie(wrtr, cookie)
}

func isWsl() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}
