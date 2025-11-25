// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Message Formatting Utilities"
//   Timestamp: "2025-11-25T13:44:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python message formatting from msgparse.py"
//   Principle_Applied: "Aether-Engineering-DRY"
//   Quality_Check: "Message format matches Python version"
// }}

package utils

import (
	"fmt"
	"strings"

	"github.com/imhuimie/let-monitor-go/internal/database"
)

// FormatThreadMessage formats a thread into a notification message
func FormatThreadMessage(thread *database.Thread, aiDescription string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s 新促销\n", strings.ToUpper(thread.Domain)))
	sb.WriteString(fmt.Sprintf("标题：%s\n", thread.Title))
	sb.WriteString(fmt.Sprintf("作者：%s\n", thread.Creator))
	sb.WriteString(fmt.Sprintf("时间：%s\n\n", thread.PubDate.Format("2006/01/02 15:04")))

	if aiDescription != "" {
		// Truncate AI description to 200 characters
		if len(aiDescription) > 200 {
			sb.WriteString(fmt.Sprintf("%s...\n\n", aiDescription[:200]))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n\n", aiDescription))
		}
	}

	sb.WriteString(thread.Link)

	return sb.String()
}

// FormatCommentMessage formats a comment into a notification message
func FormatCommentMessage(thread *database.Thread, comment *database.Comment, aiDescription string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("%s 新评论\n", strings.ToUpper(thread.Domain)))
	sb.WriteString(fmt.Sprintf("作者：%s\n", comment.Author))
	sb.WriteString(fmt.Sprintf("时间：%s\n\n", comment.CreatedAt.Format("2006/01/02 15:04")))

	// Truncate message to 200 characters
	if len(comment.Message) > 200 {
		sb.WriteString(fmt.Sprintf("%s...\n\n", comment.Message[:200]))
	} else {
		sb.WriteString(fmt.Sprintf("%s\n\n", comment.Message))
	}

	if aiDescription != "" {
		// Truncate AI description to 200 characters
		if len(aiDescription) > 200 {
			sb.WriteString(fmt.Sprintf("%s...\n\n", aiDescription[:200]))
		} else {
			sb.WriteString(fmt.Sprintf("%s\n\n", aiDescription))
		}
	}

	sb.WriteString(comment.URL)

	return sb.String()
}
