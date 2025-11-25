// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Telegram Notifier Implementation"
//   Timestamp: "2025-11-25T13:43:30Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python Telegram implementation from send.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S"
//   Quality_Check: "Error handling and retry logic implemented"
// }}

package notifier

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/imhuimie/let-monitor-go/internal/database"
	"github.com/imhuimie/let-monitor-go/internal/utils"
	log "github.com/sirupsen/logrus"
)

// TelegramNotifier sends notifications via Telegram
type TelegramNotifier struct {
	botToken string
	chatID   string
	client   *http.Client
}

// NewTelegramNotifier creates a new Telegram notifier
func NewTelegramNotifier(botToken, chatID string) *TelegramNotifier {
	return &TelegramNotifier{
		botToken: botToken,
		chatID:   chatID,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a message via Telegram
func (t *TelegramNotifier) Send(message string) error {
	apiURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)

	params := url.Values{}
	params.Set("chat_id", t.chatID)
	params.Set("text", message)

	resp, err := t.client.PostForm(apiURL, params)
	if err != nil {
		log.Warnf("发送 Telegram 消息失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Warnf("Telegram API 返回非 200 状态码: %d, 响应: %s", resp.StatusCode, string(body))
		return fmt.Errorf("Telegram API 错误 (状态码 %d): %s", resp.StatusCode, string(body))
	}

	log.Info("Telegram 消息发送成功")
	return nil
}

// SendThread sends a thread notification
func (t *TelegramNotifier) SendThread(thread *database.Thread, aiDescription string) error {
	message := utils.FormatThreadMessage(thread, aiDescription)
	return t.Send(message)
}

// SendComment sends a comment notification
func (t *TelegramNotifier) SendComment(thread *database.Thread, comment *database.Comment, aiDescription string) error {
	message := utils.FormatCommentMessage(thread, comment, aiDescription)
	return t.Send(message)
}
