// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Notifier Interface and Factory"
//   Timestamp: "2025-11-25T13:43:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python notification system from send.py"
//   Principle_Applied: "Aether-Engineering-SOLID-I (Interface Segregation), Factory Pattern"
//   Quality_Check: "Multi-channel notification support with clean interface"
// }}

package notifier

import (
	"fmt"

	"github.com/imhuimie/let-monitor-go/internal/config"
	"github.com/imhuimie/let-monitor-go/internal/database"
)

// Notifier defines the interface for sending notifications
type Notifier interface {
	Send(message string) error
	SendThread(thread *database.Thread, aiDescription string) error
	SendComment(thread *database.Thread, comment *database.Comment, aiDescription string) error
}

// NewNotifier creates a notifier based on configuration
func NewNotifier(cfg *config.Config) (Notifier, error) {
	switch cfg.NoticeType {
	case "telegram":
		return NewTelegramNotifier(cfg.TelegramBot, cfg.ChatID), nil
	case "wechat":
		return NewWeChatNotifier(cfg.WeChatKey), nil
	case "custom":
		return NewCustomNotifier(cfg.CustomURL), nil
	default:
		return nil, fmt.Errorf("不支持的通知类型: %s", cfg.NoticeType)
	}
}
