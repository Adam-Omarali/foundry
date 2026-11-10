package embeddings

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/genai"
)

// EmbeddingService handles text embeddings using Google's Gemini API
type EmbeddingService struct {
	client *genai.Client
	model  string
}

// NewEmbeddingService creates a new embedding service instance. The model is
// read here, not in a package-level var: package vars initialize before main()
// loads .env, so `go run` would see an empty value.
func NewEmbeddingService(apiKey string) (*EmbeddingService, error) {
	model := os.Getenv("GEMINI_EMBEDDING_MODEL")
	if model == "" {
		return nil, errors.New("GEMINI_EMBEDDING_MODEL is not set")
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("creating Gemini client: %w", err)
	}

	return &EmbeddingService{client: client, model: model}, nil
}

// GetEmbedding returns the embedding vector for text.
func (s *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{genai.NewContentFromText(text, genai.RoleUser)}

	result, err := s.client.Models.EmbedContent(ctx, s.model, contents, nil)
	if err != nil {
		return nil, fmt.Errorf("embedding with model %s: %w", s.model, err)
	}
	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("embedding with model %s: no embeddings returned", s.model)
	}

	values := result.Embeddings[0].Values
	slog.DebugContext(ctx, "embedding generated", "model", s.model, "dims", len(values))
	return values, nil
}
