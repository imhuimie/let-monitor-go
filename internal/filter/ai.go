// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "AI Filter Implementation"
//   Timestamp: "2025-11-25T13:45:20Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python AI filtering from core.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S"
//   Quality_Check: "Cloudflare Workers AI integration with error handling"
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

// AIFilter filters content using Cloudflare Workers AI
type AIFilter struct {
	accountID string
	token     string
	model     string
	client    *http.Client
}

// NewAIFilter creates a new AI filter
func NewAIFilter(accountID, token, model string) *AIFilter {
	return &AIFilter{
		accountID: accountID,
		token:     token,
		model:     model,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// Message represents a chat message
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// AIRequest represents the request to Cloudflare AI API
type AIRequest struct {
	Messages []Message `json:"messages"`
}

// AIResponse represents the response from Cloudflare AI API
type AIResponse struct {
	Result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	} `json:"result"`
	Success bool     `json:"success"`
	Errors  []string `json:"errors"`
}

// Filter sends content to AI and returns the filtered result
func (f *AIFilter) Filter(content string, prompt string) (string, error) {
	apiURL := fmt.Sprintf("https://api.cloudflare.com/client/v4/accounts/%s/ai/run/%s",
		f.accountID, f.model)

	req := AIRequest{
		Messages: []Message{
			{Role: "system", Content: prompt},
			{Role: "user", Content: content},
		},
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return "", fmt.Errorf("无法序列化请求: %w", err)
	}

	httpReq, err := http.NewRequest("POST", apiURL, bytes.NewBuffer(jsonData))
	if err != nil {
		return "", fmt.Errorf("无法创建请求: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", f.token))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return "", fmt.Errorf("AI API 请求失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("无法读取响应: %w", err)
	}

	var aiResp AIResponse
	if err := json.Unmarshal(body, &aiResp); err != nil {
		return "", fmt.Errorf("无法解析响应: %w", err)
	}

	if !aiResp.Success || len(aiResp.Errors) > 0 {
		return "", fmt.Errorf("AI API 返回错误: %v", aiResp.Errors)
	}

	if len(aiResp.Result.Choices) == 0 {
		return "", fmt.Errorf("AI API 返回空结果")
	}

	result := aiResp.Result.Choices[0].Message.Content

	// Extract content before "END" marker
	if idx := strings.Index(result, "END"); idx >= 0 {
		result = result[:idx]
	}

	result = strings.TrimSpace(result)

	log.Debugf("AI 过滤结果: %s", result)
	return result, nil
}

// IsValidResult checks if the AI result is valid (not "FALSE")
func (f *AIFilter) IsValidResult(result string) bool {
	return strings.ToUpper(strings.TrimSpace(result)) != "FALSE"
}
