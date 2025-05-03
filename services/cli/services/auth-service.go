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

	authProviderTypes "github.com/jacobmcgowan/simple-scheduler/shared/auth/auth-provider-types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/endpoints"
)

const CliAddr = "localhost:5556"
const CliUrl = "http://localhost:5556"
const CacheFileName = ".simple-scheduler-cli-cache"

type AuthService struct {
	ApiUrl       string
	oauth2Config oauth2.Config
	verifier     string
	server       http.Server
	state        string
	loggedId     chan struct{}
}

func (svc *AuthService) Start(ctx context.Context, clientId string, clientSecret string, providerType authProviderTypes.AuthProviderType, wg *sync.WaitGroup) error {
	svc.loggedId = make(chan struct{})

	providerEndpoint, err := getProviderEndpoints(providerType)
	if err != nil {
		return fmt.Errorf("failed to get provider url: %s", err.Error())
	}

	svc.oauth2Config = oauth2.Config{
		ClientID:     clientId,
		ClientSecret: clientSecret,
		RedirectURL:  fmt.Sprintf("%s/callback", CliUrl),
		Endpoint:     providerEndpoint,
		Scopes: []string{
			"jobs:read",
			"jobs:write",
			"runs:read",
			"runs:write",
		},
	}
	svc.verifier = oauth2.GenerateVerifier()
	svc.state, err = generateState()
	if err != nil {
		return fmt.Errorf("failed to initialize state: %s", err.Error())
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

func (svc *AuthService) handleLogin(resWrtr http.ResponseWriter, req *http.Request) {
	authCodeUrl := svc.oauth2Config.AuthCodeURL(svc.state, oauth2.AccessTypeOffline, oauth2.VerifierOption(svc.verifier))
	http.Redirect(resWrtr, req, authCodeUrl, http.StatusFound)
}

func (svc *AuthService) handleCallback(resWrtr http.ResponseWriter, req *http.Request) {
	ctx := context.Background()
	code := req.URL.Query().Get("code")
	state := req.URL.Query().Get("state")

	if code == "" {
		http.Error(resWrtr, "authorization code is required", http.StatusBadRequest)
		return
	}

	if state != svc.state {
		http.Error(resWrtr, "invalid state", http.StatusBadRequest)
		return
	}

	oauth2Token, err := svc.oauth2Config.Exchange(ctx, code, oauth2.VerifierOption(svc.verifier))
	if err != nil {
		http.Error(resWrtr, fmt.Sprintf("failed to exchange token: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	svc.saveAccessToken(oauth2Token.AccessToken)
	resWrtr.WriteHeader(http.StatusNoContent)

	svc.loggedId <- struct{}{}
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

func AddAuthHeader(req *http.Request, accessToken string) {
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", accessToken))
}

func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random state: %s", err.Error())
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func getProviderEndpoints(providerType authProviderTypes.AuthProviderType) (oauth2.Endpoint, error) {
	switch providerType {
	case authProviderTypes.GitHub:
		return endpoints.GitHub, nil
	default:
		return oauth2.Endpoint{}, fmt.Errorf("unsupported provider type: %s", providerType)
	}
}

func isWsl() bool {
	releaseData, err := exec.Command("uname", "-r").Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(releaseData)), "microsoft")
}
