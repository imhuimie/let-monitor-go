// {{RIPER-5-Enhanced:
//   Action: "Added"
//   Task_ID: "Core Monitor Implementation"
//   Timestamp: "2025-11-25T13:46:00Z"
//   Authoring_Role: "LD"
//   Analysis_Performed: "Analyzed Python core monitoring from core.py"
//   Principle_Applied: "Aether-Engineering-SOLID-S, High Cohesion"
//   Quality_Check: "Thread-safe monitoring with graceful shutdown"
// }}

package monitor

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/imhuimie/let-monitor-go/internal/config"
	"github.com/imhuimie/let-monitor-go/internal/database"
	"github.com/imhuimie/let-monitor-go/internal/filter"
	"github.com/imhuimie/let-monitor-go/internal/notifier"
	log "github.com/sirupsen/logrus"
)

// ForumMonitor is the main monitoring engine
type ForumMonitor struct {
	config    *config.Manager
	db        database.Database
	notifier  notifier.Notifier
	scraper   *Scraper
	rssParser *RSSParser

	// Filters
	keywordFilter *filter.KeywordFilter
	aiFilter      filter.AIFilterInterface

	// Control
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup
	mu     sync.RWMutex
}

// NewForumMonitor creates a new forum monitor
func NewForumMonitor(cfgMgr *config.Manager, db database.Database) (*ForumMonitor, error) {
	cfg := cfgMgr.Get()
	if cfg == nil {
		return nil, fmt.Errorf("配置未加载")
	}

	// Create notifier
	ntf, err := notifier.NewNotifier(cfg)
	if err != nil {
		return nil, fmt.Errorf("创建通知器失败: %w", err)
	}

	// Create filters
	var keywordFilter *filter.KeywordFilter
	if cfg.UseKeywordsFilter {
		keywordFilter = filter.NewKeywordFilter(cfg.KeywordsRule)
	}

	var aiFilter filter.AIFilterInterface
	if cfg.UseAIFilter {
		var err error
		aiFilter, err = filter.NewAIFilterFromConfig(cfg)
		if err != nil {
			log.Warnf("创建 AI 过滤器失败: %v，AI 过滤将被禁用", err)
			aiFilter = nil
		}
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &ForumMonitor{
		config:        cfgMgr,
		db:            db,
		notifier:      ntf,
		scraper:       NewScraper(),
		rssParser:     NewRSSParser(),
		keywordFilter: keywordFilter,
		aiFilter:      aiFilter,
		ctx:           ctx,
		cancel:        cancel,
	}, nil
}

// Start starts the monitoring loop
func (m *ForumMonitor) Start() {
	log.Info("开始监控...")

	m.wg.Add(1)
	go m.monitorLoop()
}

// Stop stops the monitoring loop gracefully
func (m *ForumMonitor) Stop() {
	log.Info("停止监控...")
	m.cancel()
	m.wg.Wait()
	log.Info("监控已停止")
}

// Reload reloads configuration and recreates components
func (m *ForumMonitor) Reload() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.config.Reload(); err != nil {
		return fmt.Errorf("重新加载配置失败: %w", err)
	}

	cfg := m.config.Get()

	// Recreate notifier
	ntf, err := notifier.NewNotifier(cfg)
	if err != nil {
		return fmt.Errorf("重新创建通知器失败: %w", err)
	}
	m.notifier = ntf

	// Recreate filters
	if cfg.UseKeywordsFilter {
		m.keywordFilter = filter.NewKeywordFilter(cfg.KeywordsRule)
	} else {
		m.keywordFilter = nil
	}

	if cfg.UseAIFilter {
		aiFilter, err := filter.NewAIFilterFromConfig(cfg)
		if err != nil {
			log.Warnf("创建 AI 过滤器失败: %v，AI 过滤将被禁用", err)
			m.aiFilter = nil
		} else {
			m.aiFilter = aiFilter
		}
	} else {
		m.aiFilter = nil
	}

	log.Info("配置重新加载成功")
	return nil
}

// monitorLoop is the main monitoring loop
func (m *ForumMonitor) monitorLoop() {
	defer m.wg.Done()

	cfg := m.config.Get()
	ticker := time.NewTicker(time.Duration(cfg.Frequency) * time.Second)
	defer ticker.Stop()

	// Run immediately on start
	m.runCheck()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.runCheck()
		}
	}
}

// runCheck performs one complete check cycle
func (m *ForumMonitor) runCheck() {
	cfg := m.config.Get()

	log.Infof("[%s] 开始检查...", time.Now().Format("2006-01-02 15:04:05"))

	// Check extra URLs first
	if len(cfg.ExtraURLs) > 0 {
		m.checkExtraURLs(cfg.ExtraURLs)
	}

	// Check RSS feeds if not only_extra mode
	if !cfg.OnlyExtra && len(cfg.URLs) > 0 {
		m.checkRSSFeeds(cfg.URLs)
	}

	freq := cfg.Frequency
	log.Infof("[%s] 检查完成，休眠 %d 秒...", time.Now().Format("2006-01-02 15:04:05"), freq)
}

// checkRSSFeeds checks all RSS feeds
func (m *ForumMonitor) checkRSSFeeds(urls []string) {
	for _, url := range urls {
		if err := m.processRSSFeed(url); err != nil {
			log.Warnf("处理 RSS feed 失败 %s: %v", url, err)
		}
		time.Sleep(1 * time.Second) // Rate limiting
	}
}

// checkExtraURLs checks extra URLs directly
func (m *ForumMonitor) checkExtraURLs(urls []string) {
	for _, url := range urls {
		// Check if thread already exists
		thread, err := m.db.FindThread(url)
		if err != nil {
			log.Warnf("查询线程失败 %s: %v", url, err)
			continue
		}

		if thread != nil {
			// Thread exists, fetch comments
			m.fetchComments(thread)
		} else {
			// Thread doesn't exist, fetch and process
			if err := m.fetchThreadPage(url); err != nil {
				log.Warnf("抓取线程页面失败 %s: %v", url, err)
			}
		}

		time.Sleep(2 * time.Second) // Rate limiting
	}
}

// IsRunning returns whether the monitor is currently running
func (m *ForumMonitor) IsRunning() bool {
	select {
	case <-m.ctx.Done():
		return false
	default:
		return true
	}
}
