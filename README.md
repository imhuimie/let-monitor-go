# Let-Monitor-Go

一个基于 Go 语言的 LowEndTalk/LowEndSpirit 论坛监控工具。监控新帖子和评论，通过 AI 进行内容分析、翻译和筛选，并推送到 Telegram 等通知渠道。

这是 [let-monitor](https://github.com/vpslog/let-monitor) Python 版本的 Go 语言重构版本。

## 功能特性

- ✅ **RSS 监控**: 定期抓取论坛 RSS feed，获取新帖子
- ✅ **评论监控**: 追踪特定帖子的新评论
- ✅ **AI 过滤**: 使用 Cloudflare Workers AI 进行内容分析和翻译
- ✅ **关键词过滤**: 支持复杂的 AND/OR 关键词匹配规则
- ✅ **多渠道通知**: 支持 Telegram、微信（息知）、自定义 Webhook
- ✅ **Web 管理界面**: 提供配置管理的 Web UI
- ✅ **灵活的数据持久化**: 支持 **MongoDB** 或 **SQLite** 数据库（可选）

## 技术栈

- Go 1.21+
- Gin Web Framework
- MongoDB / SQLite（可选）
- Cloudflare Workers AI

## 快速开始

### Docker Compose 部署（推荐）

```bash
# 1. 克隆项目
git clone https://github.com/imhuimie/let-monitor-go.git
cd let-monitor-go

# 2. 配置环境变量
cp .env.example data/.env
# 编辑 .env 文件，设置 ACCESS_TOKEN

# 3. 启动服务
docker compose up -d

# 4. 访问 Web UI
# 浏览器打开 http://localhost:5556
# 使用 ACCESS_TOKEN 登录并配置
```

### 本地开发

```bash
# 1. 安装依赖
go mod download

# 2. 启动 MongoDB（如果使用 MongoDB）
docker run -d -p 27017:27017 -v $(pwd)/data/db:/data/db mongo:7

# 或者使用 SQLite（无需启动额外服务）
# 只需确保 .env 中设置 DB_TYPE=sqlite

# 3. 配置环境变量
cp .env.example data/.env
cp config.example.json data/config.json

# 4. 运行应用
go run cmd/app/main.go
```

## 配置说明

### 环境变量 (.env)

```env
# 数据库类型选择: mongodb 或 sqlite
DB_TYPE=mongodb

# MongoDB 配置（当 DB_TYPE=mongodb 时使用）
MONGO_HOST=mongodb://localhost:27017/

# SQLite 配置（当 DB_TYPE=sqlite 时使用）
SQLITE_PATH=data/forum_monitor.db

# 其他配置
ACCESS_TOKEN=your_secure_token_here
PORT=5556
GIN_MODE=release
```

#### 数据库选择说明

**MongoDB（推荐用于生产环境）**:
- 适合高并发、大数据量场景
- 支持分布式部署和副本集
- 需要额外的 MongoDB 服务

**SQLite（适合个人使用和测试）**:
- 零配置，无需额外服务
- 轻量级，单文件数据库
- 适合小规模部署和开发测试
- 数据文件存储在 `data/` 目录

#### 使用 SQLite 的 Docker Compose 配置

如果选择使用 SQLite，修改 `docker-compose.yml`:

```yaml
services:
  app:
    environment:
      - DB_TYPE=sqlite
      - SQLITE_PATH=data/forum_monitor.db
      # 移除 MONGO_HOST 和 depends_on
    # 不需要 depends_on mongo
    
# 注释或删除整个 mongo 服务
```

### 应用配置 (data/config.json)

详细配置说明请参考 [config.example.json](config.example.json)

主要配置项：
- `urls`: RSS feed 地址列表
- `extra_urls`: 额外监控的帖子 URL
- `frequency`: 监控间隔（秒）
- `comment_filter`: 评论过滤模式（by_role/by_author）
- `use_keywords_filter`: 是否启用关键词过滤
- `use_ai_filter`: 是否启用 AI 过滤
- `notice_type`: 通知类型（telegram/wechat/custom）

## 架构文档

详细的架构设计和实现说明请参考 [ARCHITECTURE.md](ARCHITECTURE.md)

## 相比 Python 版本的优势

- ⚡ **性能提升**: Go 的并发模型和编译型特性带来显著性能提升
- 💾 **资源占用低**: 内存占用约为 Python 版本的 1/3
- 🚀 **启动速度快**: 编译后的二进制文件启动几乎瞬间完成
- 🔒 **类型安全**: 静态类型系统减少运行时错误
- 📦 **部署简单**: 单一二进制文件，无需安装运行时环境

## 数据迁移

从 Python 版本迁移数据：

```bash
# 1. 停止 Python 版本
cd let-monitor
docker compose down

# 2. 复制数据目录
cp -r let-monitor/data let-monitor-go/data

# 3. 启动 Go 版本
cd let-monitor-go
docker compose up -d
```

数据库 schema 完全兼容，无需额外迁移步骤。

## 开发

```bash
# 运行测试
go test ./...

# 生成覆盖率报告
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# 构建
go build -o let-monitor-go ./cmd/app

# 格式化代码
go fmt ./...

# 静态检查
go vet ./...
```

## 许可证

MIT License

## 致谢

- 原始 Python 版本: [let-monitor](https://github.com/vpslog/let-monitor)
- 社区贡献者

## 联系方式

- GitHub Issues: [提交问题](https://github.com/imhuimie/let-monitor-go/issues)
- Telegram 群组: [加入讨论](https://t.me/vpalogchat)