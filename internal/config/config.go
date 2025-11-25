// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Config Module Implementation"
//   Timestamp: "2025-11-25T13:40:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python config loading from core.py and web.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S, DRY"
//   Quality_Check: "Configuration validation and hot-reload support implemented"
// }}

package config

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"

	log "github.com/sirupsen/logrus"
)

// Config represents the application configuration
type Config struct {
	URLs          []string `json:"urls"`
	ExtraURLs     []string `json:"extra_urls"`
	OnlyExtra     bool     `json:"only_extra"`
	Frequency     int      `json:"frequency"`      // in seconds
	CommentFilter string   `json:"comment_filter"` // "by_role" or "by_author"

	// Keyword filter
	UseKeywordsFilter bool   `json:"use_keywords_filter"`
	KeywordsRule      string `json:"keywords_rule"`

	// AI filter
	UseAIFilter   bool   `json:"use_ai_filter"`
	CFAccountID   string `json:"cf_account_id"`
	CFToken       string `json:"cf_token"`
	Model         string `json:"model"`
	ThreadPrompt  string `json:"thread_prompt"`
	CommentPrompt string `json:"comment_prompt"`

	// Notification
	NoticeType  string `json:"notice_type"` // "telegram", "wechat", "custom"
	TelegramBot string `json:"telegrambot"`
	ChatID      string `json:"chat_id"`
	WeChatKey   string `json:"wechat_key"`
	CustomURL   string `json:"custom_url"`

	mu sync.RWMutex
}

// ConfigWrapper wraps the config with a "config" key
type ConfigWrapper struct {
	Config *Config `json:"config"`
}

// Manager handles configuration loading and reloading
type Manager struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

// NewManager creates a new configuration manager
func NewManager(configPath string) *Manager {
	return &Manager{
		configPath: configPath,
	}
}

// Load loads configuration from file
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Check if config file exists, if not copy from example
	if _, err := os.Stat(m.configPath); os.IsNotExist(err) {
		log.Infof("配置文件不存在，从 config.example.json 复制")
		if err := copyFile("config.example.json", m.configPath); err != nil {
			return fmt.Errorf("无法创建配置文件: %w", err)
		}
	}

	// Read config file
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		return fmt.Errorf("无法读取配置文件: %w", err)
	}

	// Parse JSON
	var wrapper ConfigWrapper
	if err := json.Unmarshal(data, &wrapper); err != nil {
		return fmt.Errorf("无法解析配置文件: %w", err)
	}

	if wrapper.Config == nil {
		return fmt.Errorf("配置文件格式错误: 缺少 'config' 键")
	}

	m.config = wrapper.Config

	// Set default values
	if m.config.Frequency == 0 {
		m.config.Frequency = 300
	}
	if m.config.CommentFilter == "" {
		m.config.CommentFilter = "by_role"
	}
	if m.config.NoticeType == "" {
		m.config.NoticeType = "telegram"
	}

	log.Info("配置文件加载成功")
	return nil
}

// Save saves configuration to file
func (m *Manager) Save(cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	wrapper := ConfigWrapper{Config: cfg}
	data, err := json.MarshalIndent(wrapper, "", "    ")
	if err != nil {
		return fmt.Errorf("无法序列化配置: %w", err)
	}

	if err := os.WriteFile(m.configPath, data, 0644); err != nil {
		return fmt.Errorf("无法保存配置文件: %w", err)
	}

	m.config = cfg
	log.Info("配置文件保存成功")
	return nil
}

// Get returns a copy of the current configuration
func (m *Manager) Get() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return nil
	}

	// Return a copy to prevent external modifications
	configCopy := *m.config
	return &configCopy
}

// Reload reloads the configuration from file
func (m *Manager) Reload() error {
	log.Info("重新加载配置...")
	return m.Load()
}

// Validate validates the configuration
func (cfg *Config) Validate() error {
	if cfg.Frequency < 10 {
		return fmt.Errorf("频率必须至少为 10 秒")
	}

	if cfg.CommentFilter != "by_role" && cfg.CommentFilter != "by_author" {
		return fmt.Errorf("comment_filter 必须是 'by_role' 或 'by_author'")
	}

	if cfg.NoticeType != "telegram" && cfg.NoticeType != "wechat" && cfg.NoticeType != "custom" {
		return fmt.Errorf("notice_type 必须是 'telegram', 'wechat' 或 'custom'")
	}

	// Validate notification settings
	switch cfg.NoticeType {
	case "telegram":
		if cfg.TelegramBot == "" || cfg.ChatID == "" {
			return fmt.Errorf("Telegram 配置不完整: 需要 telegrambot 和 chat_id")
		}
	case "wechat":
		if cfg.WeChatKey == "" {
			return fmt.Errorf("微信配置不完整: 需要 wechat_key")
		}
	case "custom":
		if cfg.CustomURL == "" {
			return fmt.Errorf("自定义通知配置不完整: 需要 custom_url")
		}
	}

	// Validate AI settings if enabled
	if cfg.UseAIFilter {
		if cfg.CFAccountID == "" || cfg.CFToken == "" {
			return fmt.Errorf("AI 过滤配置不完整: 需要 cf_account_id 和 cf_token")
		}
		if cfg.Model == "" {
			return fmt.Errorf("AI 过滤配置不完整: 需要 model")
		}
	}

	return nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	return os.WriteFile(dst, data, 0644)
}

// GetEnv retrieves environment variable or returns default value
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
