// api/handlers.go
package api

import (
	"encoding/json"
	"fmt"
	"foundry/backend/db"
	"foundry/backend/middleware"
	"foundry/backend/services/embeddings"
	"foundry/backend/services/llm"
	"foundry/backend/services/youtube"
	"foundry/backend/utils"
	"net/http"
	"os"
	"strings"
)

type RememberRequest struct {
	Url     string `json:"url"`
	Title   string `json:"title"`
	RawText string `json:"raw_text"`
	Type    string `json:"type"`
}

func SignInHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := middleware.VerifyGoogleToken(token)

	if err != nil {
		fmt.Println("Error verifying token:", err)
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	// Check if email exists in user info
	email, ok := user["email"].(string)
	if !ok {
		fmt.Println("Error: email not found in user info")
		http.Error(w, "Invalid user info", http.StatusInternalServerError)
		return
	}

	// Generate JWT token
	jwt, err := utils.GenerateJWT(email)
	if err != nil {
		fmt.Println("Error generating JWT:", err)
		http.Error(w, "Failed to generate token", http.StatusInternalServerError)
		return
	}

	// Return JWT token and email
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"token": jwt,
		"email": email,
	})
}

func VerifyJWTHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	fmt.Println("Auth header:", authHeader)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := utils.VerifyJWT(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"email": user["email"].(string),
	})
}

func RememberHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	fmt.Println("Auth header:", authHeader)
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := utils.VerifyJWT(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var rawText string = req["raw_text"].(string)

	if req["type"] == "youtube" {
		rawText = youtube.GetVideoData(strings.Split(req["url"].(string), "v=")[1])
	}

	textToSummarize := req["title"].(string) + " " + rawText

	// userId := user["user_id"].(int)
	// email := user["email"].(string)

	// Check if user can add more documents
	if err := db.CanAddDocument(user["user_id"].(int), user["plan"].(string)); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error":   "Document limit reached",
			"message": "You've reached your free plan limit. Please upgrade to add more documents.",
			"plan":    user["plan"],
		})
		return
	}

	llmService, err := llm.NewLLMService(os.Getenv("GOOGLE_API_KEY"))
	if err != nil {
		fmt.Println("Error initializing LLM service:", err)
		http.Error(w, "Failed to initialize LLM service", http.StatusInternalServerError)
		return
	}

	summary, err := llmService.GenerateText(r.Context(), textToSummarize)
	if err != nil {
		fmt.Println("Error generating summary:", err)
		http.Error(w, "Failed to generate summary", http.StatusInternalServerError)
		return
	}

	embeddingService, err := embeddings.NewEmbeddingService(os.Getenv("GOOGLE_API_KEY"))
	if err != nil {
		fmt.Println("Error initializing embedding service:", err)
		http.Error(w, "Failed to initialize embedding service", http.StatusInternalServerError)
		return
	}

	embedding, err := embeddingService.GetEmbeddingArray(r.Context(), summary)
	if err != nil {
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	db.InsertDocument(user["user_id"].(int), req["title"].(string), req["url"].(string), summary, embedding)

	fmt.Println("Summary:", summary)
}

func GetDocumentsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := utils.VerifyJWT(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	query := req["query"].(string)

	embeddingService, err := embeddings.NewEmbeddingService(os.Getenv("GOOGLE_API_KEY"))
	if err != nil {
		fmt.Println("Error initializing embedding service:", err)
		http.Error(w, "Failed to initialize embedding service", http.StatusInternalServerError)
		return
	}

	embedding, err := embeddingService.GetEmbeddingArray(r.Context(), query)
	if err != nil {
		http.Error(w, "Failed to generate embedding", http.StatusInternalServerError)
		return
	}

	documents := db.GetDocuments(user["user_id"].(int), embedding, query)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(documents)
}

func MarkDocumentAsReadHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")
	user, err := utils.VerifyJWT(token)
	if err != nil {
		http.Error(w, "Invalid token", http.StatusUnauthorized)
		return
	}

	var req map[string]interface{}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	documentId, ok := req["document_id"].(float64)
	if !ok {
		http.Error(w, "Missing or invalid document_id", http.StatusBadRequest)
		return
	}

	err = db.MarkDocumentAsRead(user["user_id"].(int), int(documentId))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Document marked as read successfully",
	})
}
