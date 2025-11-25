// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "RSS Parser Implementation"
//   Timestamp: "2025-11-25T13:46:20Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python RSS parsing from core.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S"
//   Quality_Check: "RSS feed parsing with gofeed library"
// }}

package monitor

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/imhuimie/let-monitor-go/internal/database"
	"github.com/mmcdole/gofeed"
	log "github.com/sirupsen/logrus"
)

// RSSParser parses RSS feeds
type RSSParser struct {
	parser *gofeed.Parser
	client *http.Client
}

// NewRSSParser creates a new RSS parser
func NewRSSParser() *RSSParser {
	return &RSSParser{
		parser: gofeed.NewParser(),
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

// ParseURL parses an RSS feed URL
func (r *RSSParser) ParseURL(url string) ([]*database.Thread, error) {
	resp, err := r.client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("获取 RSS feed 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS feed 返回状态码 %d", resp.StatusCode)
	}

	feed, err := r.parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("解析 RSS feed 失败: %w", err)
	}

	// Extract domain and category from URL
	parts := strings.Split(url, "//")
	if len(parts) < 2 {
		return nil, fmt.Errorf("无效的 RSS URL")
	}

	domainParts := strings.Split(parts[1], ".")
	domain := domainParts[0] // e.g., "lowendtalk" or "lowendspirit"

	urlParts := strings.Split(url, "/")
	category := "offers" // default
	for i, part := range urlParts {
		if part == "categories" && i+1 < len(urlParts) {
			category = urlParts[i+1]
			break
		}
	}

	// Convert feed items to threads
	var threads []*database.Thread
	for i, item := range feed.Items {
		if i >= 6 { // Only process first 6 items
			break
		}

		thread, err := r.convertItemToThread(item, domain, category)
		if err != nil {
			log.Warnf("转换 RSS item 失败: %v", err)
			continue
		}

		threads = append(threads, thread)
	}

	return threads, nil
}

// convertItemToThread converts an RSS item to a Thread
func (r *RSSParser) convertItemToThread(item *gofeed.Item, domain, category string) (*database.Thread, error) {
	if item.Link == "" {
		return nil, fmt.Errorf("RSS item 缺少链接")
	}

	var pubDate time.Time
	if item.PublishedParsed != nil {
		pubDate = *item.PublishedParsed
	} else {
		pubDate = time.Now().UTC()
	}

	description := item.Description
	if description == "" {
		description = item.Content
	}

	// Strip HTML tags from description
	description = stripHTML(description)

	creator := ""
	if item.Author != nil {
		creator = item.Author.Name
	}
	if creator == "" && len(item.Authors) > 0 {
		creator = item.Authors[0].Name
	}

	return &database.Thread{
		Domain:      domain,
		Category:    category,
		Title:       item.Title,
		Link:        item.Link,
		Description: description,
		Creator:     creator,
		PubDate:     pubDate,
		CreatedAt:   time.Now().UTC(),
		LastPage:    1,
	}, nil
}

// stripHTML removes HTML tags from text (simple implementation)
func stripHTML(html string) string {
	// Remove HTML tags
	result := html
	for {
		start := strings.Index(result, "<")
		if start == -1 {
			break
		}
		end := strings.Index(result[start:], ">")
		if end == -1 {
			break
		}
		result = result[:start] + result[start+end+1:]
	}
	return strings.TrimSpace(result)
}

// processRSSFeed processes a single RSS feed URL
func (m *ForumMonitor) processRSSFeed(url string) error {
	domain := extractDomain(url)
	category := extractCategory(url)

	log.Infof("[%s] 检查 %s %s RSS...", time.Now().Format("2006-01-02 15:04:05"), domain, category)

	threads, err := m.rssParser.ParseURL(url)
	if err != nil {
		return fmt.Errorf("解析 RSS 失败: %w", err)
	}

	for _, thread := range threads {
		m.handleThread(thread)
		m.fetchComments(thread)
		time.Sleep(500 * time.Millisecond) // Rate limiting
	}

	return nil
}

// extractDomain extracts domain from URL
func extractDomain(url string) string {
	parts := strings.Split(url, "//")
	if len(parts) < 2 {
		return ""
	}
	domainParts := strings.Split(parts[1], ".")
	return domainParts[0]
}

// extractCategory extracts category from URL
func extractCategory(url string) string {
	urlParts := strings.Split(url, "/")
	for i, part := range urlParts {
		if part == "categories" && i+1 < len(urlParts) {
			return urlParts[i+1]
		}
	}
	return "offers"
}
