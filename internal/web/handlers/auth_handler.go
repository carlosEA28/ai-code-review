package handlers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/carlosEA28/ai-code-review/internal/dto"
	"github.com/carlosEA28/ai-code-review/internal/service"
	"github.com/carlosEA28/ai-code-review/internal/web/middleware"
)

const oauthStateCookieName = "github_oauth_state"

type AuthHandler struct {
	authService service.AuthUseCase
}

func NewAuthHandler(authService service.AuthUseCase) *AuthHandler {
	return &AuthHandler{authService: authService}
}

func (h *AuthHandler) GithubLogin(w http.ResponseWriter, r *http.Request) {
	state, err := generateState()
	if err != nil {
		http.Error(w, "failed to generate oauth state", http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    state,
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(10 * time.Minute),
	})

	http.Redirect(w, r, h.authService.GithubLoginURL(state), http.StatusTemporaryRedirect)
}

func (h *AuthHandler) GithubCallback(w http.ResponseWriter, r *http.Request) {
	state := r.URL.Query().Get("state")
	if state == "" {
		http.Error(w, "state is required", http.StatusBadRequest)
		return
	}

	if err := validateOAuthState(r, state); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	code := r.URL.Query().Get("code")
	if code == "" {
		http.Error(w, "code is required", http.StatusBadRequest)
		return
	}

	user, err := h.authService.AuthenticateGithub(r.Context(), code)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     oauthStateCookieName,
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   r.TLS != nil,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})

	response := dto.GithubAuthResponseDto{
		ID:          user.User.ID.String(),
		GithubID:    user.User.GithubID,
		GithubLogin: user.User.GithubLogin,
		Name:        user.User.Name,
		Email:       user.User.Email,
		AvatarURL:   user.User.AvatarUrl,
		Token:       user.Token,
		ExpiresAt:   user.ExpiresAt.Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.TokenClaimsFromContext(r.Context())
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	response := dto.MeResponseDto{
		ID:          claims.UserID,
		GithubID:    claims.GithubID,
		GithubLogin: claims.GithubLogin,
		Name:        claims.Name,
		Email:       claims.Email,
		AvatarURL:   claims.AvatarURL,
	}

	if claims.IssuedAt != nil {
		response.IssuedAt = claims.IssuedAt.Time
	}

	if claims.ExpiresAt != nil {
		response.ExpiresAt = claims.ExpiresAt.Time
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

func validateOAuthState(r *http.Request, state string) error {
	cookie, err := r.Cookie(oauthStateCookieName)
	if err != nil {
		if errors.Is(err, http.ErrNoCookie) {
			return errors.New("oauth state cookie not found")
		}
		return errors.New("failed to validate oauth state")
	}

	if cookie.Value == "" || cookie.Value != state {
		return errors.New("invalid oauth state")
	}

	return nil
}

func generateState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(b), nil
}
