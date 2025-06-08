package embeddings

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// EmbeddingService handles text embeddings using Google's Gemini API
type EmbeddingService struct {
	client *genai.Client
}

// NewEmbeddingService creates a new embedding service instance
func NewEmbeddingService(apiKey string) (*EmbeddingService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &EmbeddingService{
		client: client,
	}, nil
}

// GetEmbedding generates embeddings for the given text
func (s *EmbeddingService) GetEmbedding(ctx context.Context, text string) ([]float32, error) {
	contents := []*genai.Content{
		genai.NewContentFromText(text, genai.RoleUser),
	}

	result, err := s.client.Models.EmbedContent(ctx,
		"gemini-embedding-exp-03-07",
		contents,
		nil,
	)
	fmt.Println("Result:", result)
	if err != nil {
		return nil, fmt.Errorf("failed to generate embedding: %v", err)
	}

	if len(result.Embeddings) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	fmt.Println("Length of embeddings:", len(result.Embeddings[0].Values))

	return result.Embeddings[0].Values, nil
}

func (s *EmbeddingService) GetEmbeddingArray(ctx context.Context, text string) ([]float32, error) {
	embeddings, err := s.GetEmbedding(ctx, text)
	if err != nil {
		return nil, err
	}

	result := make([]float32, len(embeddings))
	for i, v := range embeddings {
		result[i] = float32(v)
	}

	return result, nil
}
