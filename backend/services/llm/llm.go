package llm

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/genai"
)

const summaryInstructions = " Please summarize the following text in a concise manner. This summary will be used to create an embedding for the document. The goal is to make the content indexable and searchable with fuzzy search. Ignore anything related to advertising. Do not make the summary generic, try and highlight the most important and unique parts of the text. Include names of people, places, and companies."

// LLMService handles text generation using Google's Gemini API
type LLMService struct {
	client *genai.Client
	model  string
}

// NewLLMService creates a new LLM service instance. The model is read here,
// not in a package-level var: package vars initialize before main() loads
// .env, so `go run` would see an empty value.
func NewLLMService(apiKey string) (*LLMService, error) {
	model := os.Getenv("GEMINI_LLM_MODEL")
	if model == "" {
		return nil, errors.New("GEMINI_LLM_MODEL is not set")
	}

	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("creating Gemini client: %w", err)
	}

	return &LLMService{client: client, model: model}, nil
}

// GenerateText summarizes prompt for embedding.
func (s *LLMService) GenerateText(ctx context.Context, prompt string) (string, error) {
	resp, err := s.client.Models.GenerateContent(ctx, s.model, genai.Text(prompt+summaryInstructions), nil)
	if err != nil {
		return "", fmt.Errorf("generating text with model %s: %w", s.model, err)
	}

	text := resp.Text()
	slog.DebugContext(ctx, "summary generated", "model", s.model, "chars", len(text))
	return text, nil
}
