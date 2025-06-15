package llm

import (
	"context"
	"fmt"

	"google.golang.org/genai"
)

// LLMService handles text generation using Google's Gemini API
type LLMService struct {
	client *genai.Client
}

// NewLLMService creates a new LLM service instance
func NewLLMService(apiKey string) (*LLMService, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	return &LLMService{
		client: client,
	}, nil
}

// GenerateText generates text using Gemini
func (s *LLMService) GenerateText(ctx context.Context, prompt string) (string, error) {
	prompt += " Please summarize the following text in a concise manner. This summary will be used to create an embedding for the document. The goal is to make the content indexable and searchable with fuzzy search. Ignore anything related to advertising."
	resp, err := s.client.Models.GenerateContent(ctx, "gemini-2.0-flash", genai.Text(prompt), nil)

	if err != nil {
		return "", fmt.Errorf("failed to generate text: %v", err)
	}

	return resp.Text(), nil
}
