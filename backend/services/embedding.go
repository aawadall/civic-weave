package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/pgvector/pgvector-go"
)

// EmbeddingService handles skill embedding generation
type EmbeddingService struct {
	apiKey string
	model  string
	client *http.Client
}

// OpenAIEmbeddingRequest represents the request to OpenAI embeddings API
type OpenAIEmbeddingRequest struct {
	Model string   `json:"model"`
	Input []string `json:"input"`
}

// OpenAIEmbeddingResponse represents the response from OpenAI embeddings API
type OpenAIEmbeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// NewEmbeddingService creates a new embedding service
func NewEmbeddingService(apiKey, model string) *EmbeddingService {
	if model == "" {
		model = "text-embedding-3-small" // Default model
	}

	return &EmbeddingService{
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// GenerateEmbedding generates embedding for a single skill text using OpenAI API
func (s *EmbeddingService) GenerateEmbedding(skillText string) (pgvector.Vector, error) {
	if strings.TrimSpace(skillText) == "" {
		return pgvector.NewVector([]float32{}), fmt.Errorf("skill text cannot be empty")
	}

	embeddings, err := s.GenerateBatchEmbeddings([]string{skillText})
	if err != nil {
		return pgvector.NewVector([]float32{}), err
	}

	if len(embeddings) == 0 {
		return pgvector.NewVector([]float32{}), fmt.Errorf("no embedding returned for skill text")
	}

	return embeddings[0], nil
}

// GenerateBatchEmbeddings generates embeddings for multiple skill texts
func (s *EmbeddingService) GenerateBatchEmbeddings(skillTexts []string) ([]pgvector.Vector, error) {
	if len(skillTexts) == 0 {
		return nil, fmt.Errorf("skill texts cannot be empty")
	}

	// Clean and validate input texts
	cleanTexts := make([]string, 0, len(skillTexts))
	for _, text := range skillTexts {
		cleaned := strings.TrimSpace(text)
		if cleaned != "" {
			cleanTexts = append(cleanTexts, cleaned)
		}
	}

	if len(cleanTexts) == 0 {
		return nil, fmt.Errorf("no valid skill texts provided")
	}

	// Prepare request
	requestBody := OpenAIEmbeddingRequest{
		Model: s.model,
		Input: cleanTexts,
	}

	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make API request
	req, err := http.NewRequest("POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+s.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make API request: %w", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("OpenAI API returned status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var response OpenAIEmbeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if len(response.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned from OpenAI API")
	}

	// Convert to pgvector.Vector format
	embeddings := make([]pgvector.Vector, len(response.Data))
	for i, data := range response.Data {
		// Convert []float64 to []float32
		float32Embedding := make([]float32, len(data.Embedding))
		for j, val := range data.Embedding {
			float32Embedding[j] = float32(val)
		}
		embeddings[i] = pgvector.NewVector(float32Embedding)
	}

	return embeddings, nil
}

// ValidateEmbeddingDimensions checks if an embedding has the expected dimensions
func (s *EmbeddingService) ValidateEmbeddingDimensions(embedding pgvector.Vector, expectedDim int) error {
	if len(embedding.Slice()) == 0 {
		return fmt.Errorf("embedding is empty")
	}

	actualDim := len(embedding.Slice())
	if actualDim != expectedDim {
		return fmt.Errorf("embedding dimension mismatch: expected %d, got %d", expectedDim, actualDim)
	}

	return nil
}

// GetEmbeddingModel returns the current embedding model being used
func (s *EmbeddingService) GetEmbeddingModel() string {
	return s.model
}

// IsAPIKeyConfigured checks if the API key is properly configured
func (s *EmbeddingService) IsAPIKeyConfigured() bool {
	return s.apiKey != "" && len(s.apiKey) > 10
}
