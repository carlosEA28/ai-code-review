package handlers

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/carlosEA28/ai-code-review/internal/dto"
	"github.com/carlosEA28/ai-code-review/internal/service"
)

type WebhookHandler struct {
	webhookService service.WebhookUseCase
	webhookSecret  string
}

func NewWebhookHandler(webhookService service.WebhookUseCase, webhookSecret string) *WebhookHandler {
	return &WebhookHandler{
		webhookService: webhookService,
		webhookSecret:  strings.TrimSpace(webhookSecret),
	}
}

func (h *WebhookHandler) HandleGithubWebhook(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "failed to read request body", http.StatusBadRequest)
		return
	}

	log.Printf("github webhook payload: %s", string(body))

	if h.webhookSecret == "" {
		http.Error(w, "webhook secret is not configured", http.StatusInternalServerError)
		return
	}

	signature := r.Header.Get("X-Hub-Signature-256")
	if !isValidGithubSignature(signature, body, h.webhookSecret) {
		http.Error(w, "invalid webhook signature", http.StatusUnauthorized)
		return
	}

	if r.Header.Get("X-GitHub-Event") != "pull_request" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var payload dto.GithubPullRequestWebhookDto
	if err := json.Unmarshal(body, &payload); err != nil {
		http.Error(w, "invalid webhook payload", http.StatusBadRequest)
		return
	}

	action := strings.TrimSpace(payload.Action)
	if action != "opened" && action != "synchronize" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.WriteHeader(http.StatusOK)
	if flusher, ok := w.(http.Flusher); ok {
		flusher.Flush()
	}

	go h.queuePullRequestReview(payload)
}

func (h *WebhookHandler) queuePullRequestReview(payload dto.GithubPullRequestWebhookDto) {
	if h.webhookService == nil {
		log.Printf("webhook service is not configured")
		return
	}

	prNumber := payload.Number
	if payload.PullRequest.Number > 0 {
		prNumber = payload.PullRequest.Number
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := h.webhookService.QueuePullRequestReview(ctx, service.QueuePullRequestReviewInput{
		RepositoryFullName: payload.Repository.FullName,
		GithubPRID:         payload.PullRequest.ID,
		Number:             prNumber,
		Title:              payload.PullRequest.Title,
		Body:               payload.PullRequest.Body,
		AuthorLogin:        payload.PullRequest.User.Login,
		BaseBranch:         payload.PullRequest.Base.Ref,
		HeadBranch:         payload.PullRequest.Head.Ref,
		HeadSHA:            payload.PullRequest.Head.SHA,
		PRURL:              payload.PullRequest.HTMLURL,
		DiffURL:            payload.PullRequest.DiffURL,
	})
	if err != nil {
		log.Printf("failed to queue pull request review: %v", err)
		return
	}

	log.Printf("queued review job for %s #%d", payload.Repository.FullName, prNumber)
}

func isValidGithubSignature(signatureHeader string, body []byte, secret string) bool {
	if !strings.HasPrefix(signatureHeader, "sha256=") {
		return false
	}

	receivedSignature := strings.TrimPrefix(signatureHeader, "sha256=")
	receivedSignatureBytes, err := hex.DecodeString(receivedSignature)
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	expectedSignatureBytes := mac.Sum(nil)

	return hmac.Equal(receivedSignatureBytes, expectedSignatureBytes)
}
