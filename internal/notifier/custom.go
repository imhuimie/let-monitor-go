// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Custom Notifier Implementation"
//   Timestamp: "2025-11-25T13:45:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python custom notification from send.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S"
//   Quality_Check: "Custom webhook with message placeholder support"
// }}

package notifier

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imhuimie/let-monitor-go/internal/database"
	"github.com/imhuimie/let-monitor-go/internal/utils"
	log "github.com/sirupsen/logrus"
)

// CustomNotifier sends notifications via custom webhook
type CustomNotifier struct {
	webhookURL string
	client     *http.Client
}

// NewCustomNotifier creates a new custom notifier
func NewCustomNotifier(webhookURL string) *CustomNotifier {
	return &CustomNotifier{
		webhookURL: webhookURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a message via custom webhook
func (c *CustomNotifier) Send(message string) error {
	url := strings.ReplaceAll(c.webhookURL, "{message}", message)

	resp, err := c.client.Get(url)
	if err != nil {
		log.Warnf("发送自定义通知失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("自定义通知 API 返回非 200 状态码: %d", resp.StatusCode)
		return fmt.Errorf("自定义通知 API 错误: 状态码 %d", resp.StatusCode)
	}

	log.Info("自定义通知发送成功")
	return nil
}

// SendThread sends a thread notification
func (c *CustomNotifier) SendThread(thread *database.Thread, aiDescription string) error {
	message := utils.FormatThreadMessage(thread, aiDescription)
	return c.Send(message)
}

// SendComment sends a comment notification
func (c *CustomNotifier) SendComment(thread *database.Thread, comment *database.Comment, aiDescription string) error {
	message := utils.FormatCommentMessage(thread, comment, aiDescription)
	return c.Send(message)
}
