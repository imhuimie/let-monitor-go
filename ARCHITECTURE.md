# Let-Monitor-Go 项目架构文档

## 项目概述

let-monitor-go 是 let-monitor Python 项目的 Go 语言重构版本。该项目用于监控 LowEndTalk/LowEndSpirit 论坛的新帖子和评论，通过 AI 进行内容分析、翻译和筛选，并将结果推送到 Telegram 等通知渠道。

## 核心功能

1. **RSS 监控**: 定期抓取论坛 RSS feed，获取新帖子信息
2. **评论监控**: 追踪特定帖子的新评论
3. **AI 过滤**: 使用 Cloudflare Workers AI 进行内容分析和翻译
4. **关键词过滤**: 支持复杂的 AND/OR 关键词匹配规则
5. **多渠道通知**: 支持 Telegram、微信（息知）、自定义 Webhook
6. **Web 管理界面**: 提供配置管理的 Web UI
7. **数据持久化**: 使用 MongoDB 存储帖子和评论数据

## 技术栈

- **语言**: Go 1.21+
- **Web 框架**: Gin
- **数据库**: MongoDB
- **HTML 解析**: goquery (基于 Go's net/html)
- **HTTP 客户端**: 标准库 net/http + cloudflare-bypass
- **配置管理**: viper + godotenv
- **日志**: logrus
- **任务调度**: 自定义 goroutine + ticker

## 项目结构

```
let-monitor-go/
├── cmd/
│   └── app/
│       └── main.go                 # 应用入口
├── internal/
│   ├── config/
│   │   └── config.go              # 配置管理
│   ├── database/
│   │   ├── mongodb.go             # MongoDB 连接和操作
│   │   └── models.go              # 数据模型定义
│   ├── monitor/
│   │   ├── monitor.go             # 核心监控逻辑
│   │   ├── rss.go                 # RSS 解析
│   │   ├── scraper.go             # 网页抓取
│   │   └── comments.go            # 评论处理
│   ├── filter/
│   │   ├── keywords.go            # 关键词过滤器
│   │   └── ai.go                  # AI 过滤器
│   ├── notifier/
│   │   ├── notifier.go            # 通知接口
│   │   ├── telegram.go            # Telegram 通知
│   │   ├── wechat.go              # 微信通知
│   │   └── custom.go              # 自定义通知
│   ├── server/
│   │   ├── server.go              # Web 服务器
│   │   ├── handlers.go            # HTTP 处理器
│   │   └── middleware.go          # 中间件（认证等）
│   └── utils/
│       ├── time.go                # 时间工具
│       ├── http.go                # HTTP 工具
│       └── message.go             # 消息格式化
├── web/
│   └── templates/
│       └── index.html             # Web UI 模板
├── data/                          # 数据目录（gitignore）
│   ├── .env                       # 环境变量
│   └── config.json                # 配置文件
├── docker-compose.yml             # Docker 编排
├── Dockerfile                     # Docker 镜像
├── go.mod                         # Go 模块
├── go.sum                         # 依赖锁定
├── .env.example                   # 环境变量示例
├── config.example.json            # 配置文件示例
├── README.md                      # 项目说明
└── ARCHITECTURE.md                # 本文档

```

## 核心模块设计

### 1. Config 模块 (`internal/config`)

**职责**: 配置文件和环境变量的加载、验证和热更新

**数据结构**:
```go
type Config struct {
    URLs           []string `json:"urls"`
    ExtraURLs      []string `json:"extra_urls"`
    OnlyExtra      bool     `json:"only_extra"`
    Frequency      int      `json:"frequency"`
    CommentFilter  string   `json:"comment_filter"` // "by_role" | "by_author"
    
    // 关键词过滤
    UseKeywordsFilter bool   `json:"use_keywords_filter"`
    KeywordsRule      string `json:"keywords_rule"`
    
    // AI 过滤
    UseAIFilter   bool   `json:"use_ai_filter"`
    CFAccountID   string `json:"cf_account_id"`
    CFToken       string `json:"cf_token"`
    Model         string `json:"model"`
    ThreadPrompt  string `json:"thread_prompt"`
    CommentPrompt string `json:"comment_prompt"`
    
    // 通知配置
    NoticeType   string `json:"notice_type"` // "telegram" | "wechat" | "custom"
    TelegramBot  string `json:"telegrambot"`
    ChatID       string `json:"chat_id"`
    WeChatKey    string `json:"wechat_key"`
    CustomURL    string `json:"custom_url"`
}
```

**关键方法**:
- `LoadConfig(path string) (*Config, error)`: 从文件加载配置
- `SaveConfig(cfg *Config, path string) error`: 保存配置到文件
- `Validate() error`: 验证配置的有效性
- `Reload()`: 热更新配置

### 2. Database 模块 (`internal/database`)

**职责**: MongoDB 数据库连接和 CRUD 操作

**数据模型**:
```go
type Thread struct {
    ID          primitive.ObjectID `bson:"_id,omitempty"`
    Domain      string            `bson:"domain"`
    Category    string            `bson:"category"`
    Title       string            `bson:"title"`
    Link        string            `bson:"link" unique:"true"`
    Description string            `bson:"description"`
    Creator     string            `bson:"creator"`
    PubDate     time.Time         `bson:"pub_date"`
    CreatedAt   time.Time         `bson:"created_at"`
    LastPage    int               `bson:"last_page"`
}

type Comment struct {
    ID               primitive.ObjectID `bson:"_id,omitempty"`
    CommentID        string            `bson:"comment_id" unique:"true"`
    ThreadURL        string            `bson:"thread_url"`
    Author           string            `bson:"author"`
    Message          string            `bson:"message"`
    CreatedAt        time.Time         `bson:"created_at"`
    CreatedAtRecorded time.Time        `bson:"created_at_recorded"`
    URL              string            `bson:"url"`
}
```

**关键接口**:
```go
type Database interface {
    // Thread 操作
    InsertThread(thread *Thread) error
    FindThread(link string) (*Thread, error)
    UpdateThreadLastPage(link string, page int) error
    
    // Comment 操作
    InsertComment(comment *Comment) error
    FindComment(commentID string) (*Comment, error)
    CommentExists(commentID string) bool
    
    // 连接管理
    Connect(uri string) error
    Disconnect() error
    Ping() error
}
```

### 3. Monitor 模块 (`internal/monitor`)

**职责**: 核心监控逻辑，协调 RSS 抓取、页面解析、评论监控

**核心组件**:

#### ForumMonitor
主监控器，协调所有监控任务

```go
type ForumMonitor struct {
    config   *config.Config
    db       database.Database
    notifier notifier.Notifier
    filters  *FilterChain
    scraper  *Scraper
    
    stopCh   chan struct{}
    running  atomic.Bool
}
```

**关键方法**:
- `Start()`: 启动监控循环
- `Stop()`: 停止监控
- `CheckRSSFeeds(urls []string)`: 检查 RSS feeds
- `CheckExtraURLs(urls []string)`: 检查额外 URL
- `ProcessThread(thread *Thread)`: 处理单个帖子
- `ProcessComments(thread *Thread)`: 处理帖子评论

#### RSS Parser
解析 RSS feed 并转换为 Thread 结构

```go
type RSSParser struct{}

func (p *RSSParser) Parse(xmlData []byte) ([]*Thread, error)
func (p *RSSParser) ParseItem(item *RSSItem) (*Thread, error)
```

#### Scraper
使用 goquery 解析 HTML 页面

```go
type Scraper struct {
    client *http.Client
}

func (s *Scraper) FetchThreadPage(url string) (*Thread, error)
func (s *Scraper) FetchComments(threadURL string, page int) ([]*Comment, error)
```

### 4. Filter 模块 (`internal/filter`)

**职责**: 内容过滤（关键词、AI）

#### KeywordFilter
实现复杂的 AND/OR 关键词匹配

```go
type KeywordFilter struct {
    rule string // "keyword1+keyword2,keyword3"
}

// 匹配逻辑: OR 组用逗号分隔，AND 关键词用 + 分隔
func (f *KeywordFilter) Match(text string) bool
```

#### AIFilter
调用 Cloudflare Workers AI 进行内容分析

```go
type AIFilter struct {
    accountID string
    token     string
    model     string
}

func (f *AIFilter) Filter(content string, prompt string) (string, error)
func (f *AIFilter) IsValid(result string) bool // 判断是否返回 "FALSE"
```

**AI API 调用流程**:
1. 构建请求体 (messages array)
2. POST 到 Cloudflare API
3. 解析响应，提取 content
4. 检查是否为 "FALSE" 或有效内容

### 5. Notifier 模块 (`internal/notifier`)

**职责**: 多渠道消息通知

**接口设计**:
```go
type Notifier interface {
    Send(message string) error
    SendThread(thread *Thread, aiDescription string) error
    SendComment(thread *Thread, comment *Comment, aiDescription string) error
}
```

**实现类**:

#### TelegramNotifier
```go
type TelegramNotifier struct {
    botToken string
    chatID   string
}

func (t *TelegramNotifier) Send(message string) error {
    // POST https://api.telegram.org/bot{token}/sendMessage
}
```

#### WeChatNotifier (息知)
```go
type WeChatNotifier struct {
    apiKey string
}

func (w *WeChatNotifier) Send(message string) error {
    // GET https://xizhi.qqoq.net/{key}.send
}
```

#### CustomNotifier
```go
type CustomNotifier struct {
    webhookURL string // 支持 {message} 占位符
}

func (c *CustomNotifier) Send(message string) error {
    // 替换 {message} 后发送 GET 请求
}
```

### 6. Server 模块 (`internal/server`)

**职责**: Web API 服务器，提供配置管理界面

**路由设计**:
```
GET  /                    -> 返回 Web UI (index.html)
GET  /api/config          -> 获取当前配置 (需认证)
POST /api/config          -> 更新配置 (需认证)
GET  /api/health          -> 健康检查
```

**中间件**:
- `AuthMiddleware`: Bearer Token 认证
- `CORSMiddleware`: 跨域支持
- `LoggingMiddleware`: 请求日志

**Handler**:
```go
type ConfigHandler struct {
    monitor *monitor.ForumMonitor
    cfgPath string
}

func (h *ConfigHandler) GetConfig(c *gin.Context)
func (h *ConfigHandler) UpdateConfig(c *gin.Context)
```

### 7. Utils 模块 (`internal/utils`)

**职责**: 通用工具函数

- **Time Utils**: UTC 时间处理、格式化
- **HTTP Utils**: Cloudflare bypass、重试逻辑
- **Message Utils**: 消息格式化（Thread/Comment 转文本）

## 数据流

### 新帖监控流程

```
定时器触发
    ↓
检查 RSS Feeds
    ↓
解析 RSS XML → Thread 对象
    ↓
检查数据库是否已存在 (通过 link)
    ↓
不存在 → 插入数据库
    ↓
检查是否在 24 小时内
    ↓
是 → 应用过滤器
    ↓
关键词过滤 (可选)
    ↓
AI 过滤 (可选)
    ↓
通过 → 格式化消息
    ↓
发送通知 (Telegram/微信/Custom)
    ↓
抓取该帖子的评论
```

### 评论监控流程

```
获取 Thread 的 last_page
    ↓
循环: 从 last_page 开始抓取
    ↓
请求 {thread_url}/p{page}
    ↓
解析 HTML → Comment 对象列表
    ↓
对每个 Comment:
    ↓
检查数据库是否已存在 (通过 comment_id)
    ↓
不存在 → 插入数据库
    ↓
应用评论过滤器 (by_role / by_author)
    ↓
通过 → 检查是否在 24 小时内
    ↓
是 → 应用关键词/AI 过滤
    ↓
通过 → 发送通知
    ↓
page++, 继续循环直到 404
    ↓
更新 Thread.last_page
```

## 并发模型

### Goroutine 使用

1. **主监控循环**: 单独 goroutine
2. **Web 服务器**: Gin 自动管理 goroutine pool
3. **RSS 检查**: 
每个 RSS URL 并发抓取（使用 `sync.WaitGroup`）
4. **评论抓取**: 串行处理（避免过度请求导致被封）

### 同步机制

- **配置更新**: 使用 `sync.RWMutex` 保护配置读写
- **数据库操作**: MongoDB driver 自带连接池，天然支持并发
- **HTTP 请求**: 共享 `http.Client`，设置合理的超时和连接限制

### 优雅关闭

```go
// 监听系统信号
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

// 收到信号后
<-sigCh
monitor.Stop()        // 停止监控循环
server.Shutdown()     // 关闭 Web 服务器
db.Disconnect()       // 断开数据库连接
```

## 错误处理策略

### 1. 网络错误
- **重试机制**: 最多重试 3 次，指数退避（1s, 2s, 4s）
- **Cloudflare 挑战**: 使用 `cloudflarebp` 库绕过
- **超时设置**: 
  - RSS 抓取: 30s
  - 页面抓取: 30s
  - AI API: 60s
  - 通知发送: 10s

### 2. 数据库错误
- **连接失败**: 记录日志，等待下一次循环重试
- **唯一性冲突**: 忽略（已存在的数据）
- **写入失败**: 记录日志并继续处理其他数据

### 3. AI API 错误
- **配额超限**: 记录日志，跳过 AI 过滤
- **格式错误**: 降级为不过滤，发送原始内容
- **超时**: 记录日志，跳过该项

### 4. 日志级别
- **Error**: 严重错误（数据库连接失败、配置加载失败）
- **Warn**: 可恢复错误（单次抓取失败、AI 调用失败）
- **Info**: 正常流程（发现新帖、发送通知）
- **Debug**: 详细信息（HTTP 请求、响应体）

## 性能优化

### 1. HTTP 连接复用
```go
client := &http.Client{
    Transport: &http.Transport{
        MaxIdleConns:        100,
        MaxIdleConnsPerHost: 10,
        IdleConnTimeout:     90 * time.Second,
    },
    Timeout: 30 * time.Second,
}
```

### 2. MongoDB 索引
```go
// threads 集合
db.threads.createIndex({ "link": 1 }, { unique: true })
db.threads.createIndex({ "pub_date": -1 })

// comments 集合
db.comments.createIndex({ "comment_id": 1 }, { unique: true })
db.comments.createIndex({ "thread_url": 1, "created_at": -1 })
```

### 3. 内存控制
- RSS 解析使用流式处理
- HTML 解析完成后立即释放 Document 对象
- 定期触发 `runtime.GC()` （可选）

### 4. 批量操作
- 评论批量插入（使用 `InsertMany`）
- 通知消息批量发送（如果 API 支持）

## 配置文件示例

### config.json
```json
{
    "config": {
        "urls": [
            "https://lowendspirit.com/categories/offers/feed.rss",
            "https://lowendtalk.com/categories/offers/feed.rss"
        ],
        "extra_urls": [],
        "only_extra": false,
        "frequency": 300,
        "comment_filter": "by_role",
        
        "use_keywords_filter": true,
        "keywords_rule": "giveaway,sale+vps,discount+hosting",
        
        "use_ai_filter": false,
        "cf_account_id": "",
        "cf_token": "",
        "model": "@cf/qwen/qwen3-30b-a3b-fp8",
        "thread_prompt": "设定：你是一位精通 VPS 相关信息的中文助手...",
        "comment_prompt": "设定：你是一位精通 VPS 相关信息的中文助手...",
        
        "notice_type": "telegram",
        "telegrambot": "123456:ABCdefGHIjklMNOpqrSTUvwxYZ",
        "chat_id": "-1001234567890",
        "wechat_key": "",
        "custom_url": ""
    }
}
```

### .env
```env
MONGO_HOST=mongodb://localhost:27017/
ACCESS_TOKEN=your_secure_token_here
PORT=5556
GIN_MODE=release
```

## 部署方案

### Docker Compose 部署 (推荐)

```yaml
services:
  app:
    build: .
    ports:
      - "5556:5556"
    volumes:
      - ./data:/app/data
    environment:
      - MONGO_HOST=mongodb://mongo:27017/
      - ACCESS_TOKEN=${ACCESS_TOKEN}
      - GIN_MODE=release
    depends_on:
      - mongo
    restart: unless-stopped

  mongo:
    image: mongo:7
    volumes:
      - ./data/db:/data/db
    restart: unless-stopped
```

## 依赖包

### 核心依赖

```go
require (
    github.com/gin-gonic/gin v1.10.0
    go.mongodb.org/mongo-driver v1.13.1
    github.com/PuerkitoBio/goquery v1.8.1
    github.com/mmcdole/gofeed v1.2.1
    github.com/spf13/viper v1.18.2
    github.com/joho/godotenv v1.5.1
    github.com/sirupsen/logrus v1.9.3
)
```

## 安全建议

1. **ACCESS_TOKEN**: 使用强随机字符串（>32 字符）
2. **HTTPS**: 生产环境使用反向代理（Nginx + Let's Encrypt）
3. **MongoDB**: 启用认证，绑定到 localhost 或内网
4. **API Keys**: 使用环境变量，不要硬编码

## 故障排查

### 常见问题

**Q: 监控不工作，没有收到通知**
- 检查配置文件是否正确加载
- 检查 MongoDB 连接状态
- 验证 Telegram Bot Token 和 Chat ID

**Q: AI 过滤不生效**
- 验证 Cloudflare API 凭证
- 检查 API 配额是否用完

**Q: 数据库写入失败**
- 检查 MongoDB 连接
- 检查磁盘空间

## 总结

let-monitor-go 项目采用模块化设计，清晰的职责分离，完整的错误处理和日志记录。通过 Go 语言的并发特性和高性能 HTTP 客户端，相比 Python 版本有显著的性能提升和更低的资源占用。