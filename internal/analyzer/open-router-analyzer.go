package analyzer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/BerylCAtieno/document-summarizer-api/internal/utils"

	"github.com/BerylCAtieno/document-summarizer-api/internal/models"
)

type Analyzer interface {
	Analyze(ctx context.Context, text string) (*models.LLMAnalysisResult, error)
}

type openRouterAnalyzer struct {
	apiKey string
	model  string
	logger *utils.Logger
	client *http.Client
}

type OpenRouterRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type OpenRouterResponse struct {
	Choices []Choice `json:"choices"`
	Error   *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
	} `json:"error,omitempty"`
}

type Choice struct {
	Message Message `json:"message"`
}

func NewOpenRouterAnalyzer(apiKey, model string, logger *utils.Logger) Analyzer {
	return &openRouterAnalyzer{
		apiKey: apiKey,
		model:  model,
		logger: logger,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

func (a *openRouterAnalyzer) Analyze(ctx context.Context, text string) (*models.LLMAnalysisResult, error) {
	// Truncate text if too long (keep first 4000 characters)
	if len(text) > 4000 {
		text = text[:4000] + "..."
	}

	prompt := fmt.Sprintf(`Analyze the following document and provide a structured response in JSON format only.

Document text:
%s

Respond ONLY with a valid JSON object (no markdown, no code blocks) with the following structure:
{
  "summary": "A concise 2-3 sentence summary of the document",
  "document_type": "The type of document (invoice, cv, resume, report, letter, contract, memo, email, etc.)",
  "metadata": {
    "date": "Extracted date if found (format: YYYY-MM-DD) or null",
    "sender": "Sender name if found or null",
    "recipient": "Recipient name if found or null",
    "amount": "Total amount if invoice/financial document or null",
    "currency": "Currency code if amount found or null",
    "company": "Company name if found or null"
  }
}`, text)

	reqBody := OpenRouterRequest{
		Model: a.model,
		Messages: []Message{
			{
				Role:    "user",
				Content: prompt,
			},
		},
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", "https://openrouter.ai/api/v1/chat/completions", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+a.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/yourusername/document-api")

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		a.logger.Error("OpenRouter API error", "status", resp.StatusCode, "body", string(body))
		return nil, fmt.Errorf("OpenRouter API returned status %d", resp.StatusCode)
	}

	var openRouterResp OpenRouterResponse
	if err := json.Unmarshal(body, &openRouterResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if openRouterResp.Error != nil {
		return nil, fmt.Errorf("OpenRouter API error: %s", openRouterResp.Error.Message)
	}

	if len(openRouterResp.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := openRouterResp.Choices[0].Message.Content

	// Parse the LLM response as JSON
	var result models.LLMAnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		// If parsing fails, try to extract JSON from markdown code blocks
		content = extractJSON(content)
		if err := json.Unmarshal([]byte(content), &result); err != nil {
			a.logger.Error("Failed to parse LLM response", "content", content)
			return nil, fmt.Errorf("failed to parse LLM response as JSON: %w", err)
		}
	}

	return &result, nil
}

// extractJSON attempts to extract JSON from markdown code blocks
func extractJSON(content string) string {
	// Remove markdown code blocks if present
	if len(content) > 7 && content[:3] == "```" {
		start := 0
		end := len(content)

		// Find first newline after opening ```
		for i := 3; i < len(content); i++ {
			if content[i] == '\n' {
				start = i + 1
				break
			}
		}

		// Find closing ```
		for i := len(content) - 1; i >= 0; i-- {
			if i >= 2 && content[i-2:i+1] == "```" {
				end = i - 2
				break
			}
		}

		if start < end {
			content = content[start:end]
		}
	}

	return content
}
