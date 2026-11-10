// api/handlers.go
package api

import (
	"context"
	"encoding/json"
	"errors"
	"foundry/backend/db"
	"foundry/backend/middleware"
	"foundry/backend/services/embeddings"
	"foundry/backend/services/llm"
	"foundry/backend/services/youtube"
	"foundry/backend/utils"
	"log/slog"
	"net/http"
	"os"

	"github.com/google/uuid"
)

type RememberRequest struct {
	Url     string `json:"url"`
	Title   string `json:"title"`
	RawText string `json:"raw_text"`
	Type    string `json:"type"`
}

// requireAuth verifies the Foundry session JWT. On failure it writes a 401 and
// returns nil, so callers just `if claims == nil { return }`.
func requireAuth(w http.ResponseWriter, r *http.Request) *utils.FoundryClaims {
	claims, err := middleware.Authenticate(r)
	if err != nil {
		http.Error(w, "Invalid or missing token", http.StatusUnauthorized)
		return nil
	}
	return claims
}

// GoogleAuthHandler exchanges a Google ID token for a Foundry session JWT.
// This is the only place a Google token is accepted, and the only place the
// user row is created/updated on sign-in.
//
//	POST /api/auth/google  {"id_token": "..."}
//	→ 200 {"token": "...", "email": "...", "expires_at": <unix seconds>}
func GoogleAuthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		IDToken string `json:"id_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IDToken == "" {
		http.Error(w, "Body must be JSON with an id_token", http.StatusBadRequest)
		return
	}

	email, err := middleware.VerifyGoogleIDToken(r.Context(), req.IDToken)
	if err != nil {
		slog.WarnContext(r.Context(), "google id token rejected", "err", err)
		http.Error(w, "Invalid Google ID token", http.StatusUnauthorized)
		return
	}

	user, err := db.UpsertUser(email)
	if err != nil {
		slog.ErrorContext(r.Context(), "upserting user", "err", err)
		http.Error(w, "Failed to sign in", http.StatusInternalServerError)
		return
	}

	token, expiresAt, err := utils.GenerateJWT(user.ID, user.UserEmail)
	if err != nil {
		slog.ErrorContext(r.Context(), "generating jwt", "err", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]any{
		"token":      token,
		"email":      user.UserEmail,
		"expires_at": expiresAt.Unix(),
	})
}

func VerifyJWTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"email": claims.Email,
	})
}

func NewDocumentHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	var req map[string]interface{}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var rawText string = req["raw_text"].(string)

	// For videos, prefer the description from the YouTube API. If that fails
	// (bad key, quota, private video) keep the page text the extension scraped
	// rather than failing the save.
	if req["type"] == "youtube" {
		if desc, err := videoDescription(r.Context(), req["url"].(string)); err != nil {
			slog.WarnContext(r.Context(), "youtube lookup failed, using scraped page text", "err", err)
		} else if desc != "" {
			rawText = desc
		}
	}

	textToSummarize := req["title"].(string) + " " + rawText

	// The plan can change after the JWT was issued, so read it fresh.
	user, err := db.GetUserByID(claims.UserID)
	if err != nil {
		slog.ErrorContext(r.Context(), "loading user", "err", err, "user_id", claims.UserID)
		http.Error(w, "Invalid or missing token", http.StatusUnauthorized)
		return
	}

	// Check if user can add more documents
	if err := db.CanAddDocument(user.ID, user.Plan); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Document limit reached",
			"message": "You've reached your free plan limit. Please upgrade to add more documents.",
			"plan":    user.Plan,
		})
		return
	}

	llmService, err := llm.NewLLMService(os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		slog.ErrorContext(r.Context(), "initializing llm service", "err", err)
		http.Error(w, "Failed to initialize LLM service", http.StatusInternalServerError)
		return
	}

	summary, err := llmService.GenerateText(r.Context(), textToSummarize)
	if err != nil {
		slog.ErrorContext(r.Context(), "generating summary", "err", err)
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	embeddingService, err := embeddings.NewEmbeddingService(os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		slog.ErrorContext(r.Context(), "initializing embedding service", "err", err)
		http.Error(w, "Failed to initialize embedding service", http.StatusInternalServerError)
		return
	}

	embedding, err := embeddingService.GetEmbedding(r.Context(), summary)
	if err != nil {
		slog.ErrorContext(r.Context(), "generating embedding", "err", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	db.InsertDocument(user.ID, req["title"].(string), req["url"].(string), summary, embedding)

	slog.DebugContext(r.Context(), "document saved", "summary", summary)
}

func GetDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	query := r.URL.Query().Get("query")
	if query == "" {
		http.Error(w, "No query parameter 'query' provided", http.StatusBadRequest)
		return
	}

	embeddingService, err := embeddings.NewEmbeddingService(os.Getenv("GEMINI_API_KEY"))
	if err != nil {
		slog.ErrorContext(r.Context(), "initializing embedding service", "err", err)
		http.Error(w, "Failed to initialize embedding service", http.StatusInternalServerError)
		return
	}

	embedding, err := embeddingService.GetEmbedding(r.Context(), query)
	if err != nil {
		slog.ErrorContext(r.Context(), "generating embedding", "err", err)
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	documents := db.GetDocuments(claims.UserID, embedding, query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

// UpdateDocumentHandler applies a partial update to one of the user's documents.
//
//	PATCH /api/documents/{id}  {"read": true}
//	→ 200 <updated document> | 400 bad id/field/type | 404 not found
//
// Updatable fields are allowlisted in db.documentUpdatableFields.
func UpdateDocumentHandler(w http.ResponseWriter, r *http.Request) {
	claims := requireAuth(w, r)
	if claims == nil {
		return
	}

	// Parse at the edge: a malformed ID is the client's fault (400), and
	// everything below this line can trust it's a real UUID.
	documentId, err := uuid.Parse(r.PathValue("id"))
	if err != nil {
		http.Error(w, "Invalid document id", http.StatusBadRequest)
		return
	}

	var body map[string]json.RawMessage
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
		http.Error(w, "Body must be a JSON object of {field: newValue}", http.StatusBadRequest)
		return
	}
	patch, err := db.ParseDocumentPatch(body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	doc, err := db.UpdateDocument(claims.UserID, documentId, patch)
	if errors.Is(err, db.ErrDocumentNotFound) {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "updating document", "err", err, "document_id", documentId)
		http.Error(w, "Failed to update document", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(doc)
}

func videoDescription(ctx context.Context, rawURL string) (string, error) {
	id, err := youtube.VideoID(rawURL)
	if err != nil {
		return "", err
	}
	return youtube.GetVideoDescription(ctx, id)
}
