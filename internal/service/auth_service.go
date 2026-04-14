package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/carlosEA28/ai-code-review/internal/domain"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/github"
)

const (
	githubUserEndpoint   = "https://api.github.com/user"
	githubEmailsEndpoint = "https://api.github.com/user/emails"
)

type GithubOAuthConfig struct {
	ClientID     string
	ClientSecret string
	RedirectURL  string
	Scopes       []string
}

type AuthUseCase interface {
	GithubLoginURL(state string) string
	AuthenticateGithub(ctx context.Context, code string) (*AuthenticatedUser, error)
}

type AuthenticatedUser struct {
	User      *domain.User
	Token     string
	ExpiresAt time.Time
}

type AuthService struct {
	oauthConfig *oauth2.Config
	userService UserUseCase
	jwtService  TokenUseCase
}

func NewAuthService(userService UserUseCase, jwtService TokenUseCase, cfg GithubOAuthConfig) (*AuthService, error) {
	if userService == nil {
		return nil, errors.New("user service is required")
	}

	if jwtService == nil {
		return nil, errors.New("jwt service is required")
	}

	clientID := strings.TrimSpace(cfg.ClientID)
	clientSecret := strings.TrimSpace(cfg.ClientSecret)
	redirectURL := strings.TrimSpace(cfg.RedirectURL)

	if clientID == "" {
		return nil, errors.New("github client id is required")
	}

	if clientSecret == "" {
		return nil, errors.New("github client secret is required")
	}

	if redirectURL == "" {
		return nil, errors.New("github redirect url is required")
	}

	scopes := cfg.Scopes
	if len(scopes) == 0 {
		scopes = []string{"read:user", "user:email"}
	}

	return &AuthService{
		oauthConfig: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			RedirectURL:  redirectURL,
			Scopes:       scopes,
			Endpoint:     github.Endpoint,
		},
		userService: userService,
		jwtService:  jwtService,
	}, nil
}

func (s *AuthService) GithubLoginURL(state string) string {
	return s.oauthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
}

func (s *AuthService) AuthenticateGithub(ctx context.Context, code string) (*AuthenticatedUser, error) {
	code = strings.TrimSpace(code)
	if code == "" {
		return nil, errors.New("authorization code is required")
	}

	token, err := s.oauthConfig.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("exchange github oauth code: %w", err)
	}

	githubUser, err := s.fetchGithubUser(ctx, token)
	if err != nil {
		return nil, err
	}

	email := strings.TrimSpace(githubUser.Email)
	if email == "" {
		email, err = s.fetchPrimaryEmail(ctx, token)
		if err != nil {
			return nil, err
		}
	}

	user, err := s.userService.UpsertGithubUser(ctx, SaveGithubUserInput{
		GithubID:    githubUser.ID,
		GithubLogin: githubUser.Login,
		Name:        githubUser.Name,
		Email:       email,
		AvatarURL:   githubUser.AvatarURL,
		GithubToken: token.AccessToken,
	})
	if err != nil {
		return nil, fmt.Errorf("persist github user: %w", err)
	}

	jwtToken, expiresAt, err := s.jwtService.GenerateToken(user)
	if err != nil {
		return nil, fmt.Errorf("generate jwt token: %w", err)
	}

	return &AuthenticatedUser{
		User:      user,
		Token:     jwtToken,
		ExpiresAt: expiresAt,
	}, nil
}

type githubUserResponse struct {
	ID        int    `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	AvatarURL string `json:"avatar_url"`
}

type githubEmailResponse struct {
	Email    string `json:"email"`
	Primary  bool   `json:"primary"`
	Verified bool   `json:"verified"`
}

func (s *AuthService) fetchGithubUser(ctx context.Context, token *oauth2.Token) (*githubUserResponse, error) {
	var output githubUserResponse

	if err := s.githubGet(ctx, token, githubUserEndpoint, &output); err != nil {
		return nil, fmt.Errorf("fetch github user profile: %w", err)
	}

	output.Login = strings.TrimSpace(output.Login)
	if output.ID <= 0 || output.Login == "" {
		return nil, errors.New("invalid github user response")
	}

	return &output, nil
}

func (s *AuthService) fetchPrimaryEmail(ctx context.Context, token *oauth2.Token) (string, error) {
	var emails []githubEmailResponse

	if err := s.githubGet(ctx, token, githubEmailsEndpoint, &emails); err != nil {
		return "", fmt.Errorf("fetch github user emails: %w", err)
	}

	for _, email := range emails {
		if email.Primary && email.Verified {
			return strings.TrimSpace(email.Email), nil
		}
	}

	for _, email := range emails {
		if email.Verified {
			return strings.TrimSpace(email.Email), nil
		}
	}

	for _, email := range emails {
		if strings.TrimSpace(email.Email) != "" {
			return strings.TrimSpace(email.Email), nil
		}
	}

	return "", nil
}

func (s *AuthService) githubGet(ctx context.Context, token *oauth2.Token, endpoint string, output any) error {
	client := s.oauthConfig.Client(ctx, token)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("build request to github: %w", err)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request github api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("github api status %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	if err := json.NewDecoder(resp.Body).Decode(output); err != nil {
		return fmt.Errorf("decode github response: %w", err)
	}

	return nil
}
