package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/carlosEA28/ai-code-review/internal/domain"
	"github.com/golang-jwt/jwt/v5"
)

type JWTConfig struct {
	SecretKey  string
	Issuer     string
	Expiration time.Duration
}

type TokenClaims struct {
	UserID      string `json:"user_id"`
	GithubID    int    `json:"github_id"`
	GithubLogin string `json:"github_login"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	AvatarURL   string `json:"avatar_url"`
	jwt.RegisteredClaims
}

type TokenUseCase interface {
	GenerateToken(user *domain.User) (token string, expiresAt time.Time, err error)
	ParseToken(token string) (*TokenClaims, error)
}

type JWTService struct {
	secretKey  []byte
	issuer     string
	expiration time.Duration
}

func NewJWTService(cfg JWTConfig) (*JWTService, error) {
	secret := strings.TrimSpace(cfg.SecretKey)
	if secret == "" {
		return nil, errors.New("jwt secret is required")
	}

	issuer := strings.TrimSpace(cfg.Issuer)
	if issuer == "" {
		issuer = "ai-code-review"
	}

	expiration := cfg.Expiration
	if expiration <= 0 {
		expiration = 24 * time.Hour
	}

	return &JWTService{
		secretKey:  []byte(secret),
		issuer:     issuer,
		expiration: expiration,
	}, nil
}

func (s *JWTService) GenerateToken(user *domain.User) (string, time.Time, error) {
	if user == nil {
		return "", time.Time{}, errors.New("user is required")
	}

	now := time.Now()
	expiresAt := now.Add(s.expiration)

	claims := TokenClaims{
		UserID:      user.ID.String(),
		GithubID:    user.GithubID,
		GithubLogin: user.GithubLogin,
		Name:        user.Name,
		Email:       user.Email,
		AvatarURL:   user.AvatarUrl,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    s.issuer,
			Subject:   user.ID.String(),
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("sign jwt token: %w", err)
	}

	return tokenString, expiresAt, nil
}

func (s *JWTService) ParseToken(token string) (*TokenClaims, error) {
	token = strings.TrimSpace(token)
	if token == "" {
		return nil, errors.New("token is required")
	}

	parsedToken, err := jwt.ParseWithClaims(token, &TokenClaims{}, func(t *jwt.Token) (any, error) {
		if t.Method != jwt.SigningMethodHS256 {
			return nil, errors.New("unexpected token signing method")
		}

		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("parse jwt token: %w", err)
	}

	claims, ok := parsedToken.Claims.(*TokenClaims)
	if !ok || !parsedToken.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}
