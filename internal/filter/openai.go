// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "OpenAI Compatible API Filter"
//   Timestamp: "2025-11-25T14:21:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Implemented OpenAI-compatible API support for AI filtering"
//   Principle_Applied: "Aether-Engineering-SOLID-O, Interface Segregation"
//   Quality_Check: "Full OpenAI API compatibility with error handling and response parsing"
// }}

package filter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// OpenAIFilter filters content using OpenAI-compatible API
type OpenAIFilter struct {
	apiURL string
	apiKey string
	model  string
	client *http.Client
}

// NewOpenAIFilter creates a new OpenAI-compatible filter
func NewOpenAIFilter(apiURL, apiKey, model string) *OpenAIFilter {
	return &OpenAIFilter{
		apiURL: apiURL,
		apiKey: apiKey,
		model:  model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// OpenAIMessage represents a chat message in OpenAI format
type OpenAIMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// OpenAIRequest represents the request to OpenAI-compatible API
type OpenAIRequest struct {
	Model    string          `json:"model"`
	Messages []OpenAIMessage `json:"messages"`
}

// OpenAIResponse represents the response from OpenAI-compatible API
type OpenAIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// Filter sends content to OpenAI-compatible API and returns the filtered result
func (f *OpenAIFilter) Filter(content string, prompt string) (string, error) {
	req := OpenAIRequest{
		Model: f.model,
		Messages: []OpenAIMessage{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("无法序列化请求: %w", err)
	}

	httpReq, err := http.NewRequest("POST", f.apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("无法创建请求: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.apiKey))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("OpenAI API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("无法读取响应: %w", err)
	}

	// Check for HTTP errors
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("OpenAI API 返回错误状态码 %d: %s", resp.StatusCode, string(body))
	}

	var openAIResp OpenAIResponse
	if err := json.Unmarshal(body, &openAIResp); err != nil {
		return "", fmt.Errorf("无法解析响应: %w", err)
	}

	if len(openAIResp.Choices) == 0 {
		return "", fmt.Errorf("OpenAI API 返回空结果")
	}

	result := openAIResp.Choices[0].Message.Content

	// Extract content before "END" marker (consistent with Cloudflare implementation)
	if idx := strings.Index(result, "END"); idx >= 0 {
		result = result[:idx]
	}

	result = strings.TrimSpace(result)

	log.Debugf("OpenAI AI 过滤结果: %s", result)
	log.Debugf("Token 使用量 - Prompt: %d, Completion: %d, Total: %d",
		openAIResp.Usage.PromptTokens,
		openAIResp.Usage.CompletionTokens,
		openAIResp.Usage.TotalTokens)

	return result, nil
}

// IsValidResult checks if the AI result is valid (not "FALSE")
func (f *OpenAIFilter) IsValidResult(result string) bool {
	return strings.ToUpper(strings.TrimSpace(result)) != "FALSE"
}
