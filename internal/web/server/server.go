package server

import (
	"net/http"

	"github.com/carlosEA28/ai-code-review/internal/service"
	"github.com/carlosEA28/ai-code-review/internal/web/handlers"
	"github.com/carlosEA28/ai-code-review/internal/web/middleware"
	"github.com/go-chi/chi/v5"
)

type Server struct {
	router         *chi.Mux
	server         *http.Server
	port           string
	authUseCase    service.AuthUseCase
	tokenUseCase   service.TokenUseCase
	webhookUseCase service.WebhookUseCase
	webhookSecret  string
}

func NewServer(
	port string,
	authUseCase service.AuthUseCase,
	tokenUseCase service.TokenUseCase,
	webhookUseCase service.WebhookUseCase,
	webhookSecret string,
) *Server {
	return &Server{
		router:         chi.NewRouter(),
		port:           port,
		authUseCase:    authUseCase,
		tokenUseCase:   tokenUseCase,
		webhookUseCase: webhookUseCase,
		webhookSecret:  webhookSecret,
	}
}

func (s *Server) ConfigureRoutes() {
	authHandler := handlers.NewAuthHandler(s.authUseCase)
	webhookHandler := handlers.NewWebhookHandler(s.webhookUseCase, s.webhookSecret)

	s.router.Get("/auth/github/login", authHandler.GithubLogin)
	s.router.Get("/auth/github/callback", authHandler.GithubCallback)
	s.router.Post("/webhooks/github", webhookHandler.HandleGithubWebhook)

	s.router.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(s.tokenUseCase))
		r.Get("/auth/me", authHandler.Me)
	})
}

func (s *Server) Start() error {
	s.ConfigureRoutes()

	s.server = &http.Server{
		Addr:    ":" + s.port,
		Handler: s.router,
	}
	return s.server.ListenAndServe()
}
