package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type GroqClient struct {
	apiKey string
	model  string
	client *http.Client
}

func NewGroqClient(apiKey string) *GroqClient {
	return &GroqClient{
		apiKey: apiKey,
		model:  "llama-3.3-70b-versatile", // El modelo con mejor balance razonamiento/velocidad
		client: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *GroqClient) Chat(ctx context.Context, systemPrompt string, userPrompt string) (string, error) {
	url := "https://api.groq.com/openai/v1/chat/completions"

	// Payload siguiendo el estándar de OpenAI que usa Groq
	payload := map[string]interface{}{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": userPrompt},
		},
		"response_format": map[string]string{"type": "json_object"}, // Forzamos salida JSON
		"temperature":     0.1, // Baja temperatura para mayor consistencia en extracción de datos
	}

	jsonData, _ := json.Marshal(payload)
	req, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonData))
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error llamando a Groq: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("error de API Groq: status %d", resp.StatusCode)
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", err
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no se recibió respuesta de Groq")
	}

	return result.Choices[0].Message.Content, nil
}