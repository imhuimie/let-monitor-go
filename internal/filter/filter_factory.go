// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "AI Filter Factory Pattern"
//   Timestamp: "2025-11-25T14:22:00Z"
//   Authoring_Role: "AR"
//   Analysis_Performed: "Implemented Factory pattern for AI filter provider selection"
//   Principle_Applied: "Aether-Engineering-SOLID-O (Open/Closed Principle)"
//   Quality_Check: "Supports easy addition of new AI providers"
// }}

package filter

import (
	"fmt"

	"github.com/imhuimie/let-monitor-go/internal/config"
)

// AIFilterInterface defines the common interface for all AI filters
type AIFilterInterface interface {
	Filter(content string, prompt string) (string, error)
	IsValidResult(result string) bool
}

// NewAIFilterFromConfig creates an AI filter based on configuration
func NewAIFilterFromConfig(cfg *config.Config) (AIFilterInterface, error) {
	if !cfg.UseAIFilter {
		return nil, fmt.Errorf("AI 过滤未启用")
	}

	switch cfg.AIProvider {
	case "cloudflare":
		return NewAIFilter(cfg.CFAccountID, cfg.CFToken, cfg.Model), nil
	case "openai":
		return NewOpenAIFilter(cfg.OpenAIAPIURL, cfg.OpenAIAPIKey, cfg.OpenAIModel), nil
	default:
		return nil, fmt.Errorf("不支持的 AI 提供商: %s", cfg.AIProvider)
	}
}
