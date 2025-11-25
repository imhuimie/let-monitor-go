// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "WeChat Notifier Implementation"
//   Timestamp: "2025-11-25T13:44:30Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python WeChat implementation from send.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S"
//   Quality_Check: "息知 API integration implemented"
// }}

package notifier

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/imhuimie/let-monitor-go/internal/database"
	"github.com/imhuimie/let-monitor-go/internal/utils"
	log "github.com/sirupsen/logrus"
)

// WeChatNotifier sends notifications via WeChat (息知)
type WeChatNotifier struct {
	apiKey string
	client *http.Client
}

// NewWeChatNotifier creates a new WeChat notifier
func NewWeChatNotifier(apiKey string) *WeChatNotifier {
	return &WeChatNotifier{
		apiKey: apiKey,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

// Send sends a message via WeChat
func (w *WeChatNotifier) Send(message string) error {
	apiURL := fmt.Sprintf("https://xizhi.qqoq.net/%s.send", w.apiKey)

	params := url.Values{}
	params.Set("title", "库存变更通知")
	params.Set("content", message)

	resp, err := w.client.Get(apiURL + "?" + params.Encode())
	if err != nil {
		log.Warnf("发送微信消息失败: %v", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Warnf("微信 API 返回非 200 状态码: %d", resp.StatusCode)
		return fmt.Errorf("微信 API 错误: 状态码 %d", resp.StatusCode)
	}

	log.Info("微信消息发送成功")
	return nil
}

// SendThread sends a thread notification
func (w *WeChatNotifier) SendThread(thread *database.Thread, aiDescription string) error {
	message := utils.FormatThreadMessage(thread, aiDescription)
	return w.Send(message)
}

// SendComment sends a comment notification
func (w *WeChatNotifier) SendComment(thread *database.Thread, comment *database.Comment, aiDescription string) error {
	message := utils.FormatCommentMessage(thread, comment, aiDescription)
	return w.Send(message)
}
