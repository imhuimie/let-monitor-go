# Let-Monitor-Go 部署指南

## 📋 更新摘要

本次更新优化了Docker配置和CI/CD流程，主要改进包括：

### 1. Dockerfile 更新 ✅
- **Go版本升级**: 从 Go 1.21 升级到 Go 1.23
- **SQLite支持**: 启用CGO以支持 `mattn/go-sqlite3`
- **静态链接**: 使用静态链接减少运行时依赖
- **模板文件**: 正确复制Web UI模板文件

### 2. docker-compose.yml 优化 ✅
- **默认使用SQLite**: 更轻量级，适合小型部署
- **环境变量优化**: 所有配置支持通过环境变量覆盖
- **MongoDB可选**: MongoDB服务默认注释，需要时可启用
- **容器命名**: 添加明确的容器名称便于管理

### 3. GitHub Actions 工作流 ✅
- **自动构建**: 推送到main分支时自动构建Docker镜像
- **GHCR发布**: 发布到GitHub Container Registry
- **构建缓存**: 使用GitHub Actions缓存加速构建
- **手动触发**: 支持 workflow_dispatch 手动触发

### 4. .dockerignore 文件 ✅
- **优化构建**: 排除不必要的文件减少构建上下文
- **保护敏感数据**: 排除 .env 和数据文件
- **加速构建**: 减少传输到Docker daemon的数据量

---

## 🚀 快速开始

### 使用 Docker Compose（推荐）

#### SQLite模式（默认）
```bash
# 1. 克隆仓库
git clone <repository-url>
cd let-monitor-go

# 2. 创建环境配置
cp .env.example data/.env
# 编辑 data/.env 设置你的 ACCESS_TOKEN

# 3. 启动服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f app
```

#### MongoDB模式
```bash
# 1. 编辑 docker-compose.yml
# - 取消 mongo 服务的注释
# - 在 app 服务中启用 depends_on

# 2. 编辑 data/.env 设置：
DB_TYPE=mongodb
MONGO_HOST=mongodb://mongo:27017/

# 3. 启动服务
docker-compose up -d

# 4. 查看日志
docker-compose logs -f
```

### 使用 Docker（不使用 Compose）

```bash
# 构建镜像
docker build -t let-monitor-go .

# 运行容器（SQLite模式）
docker run -d \
  --name let-monitor-go \
  -p 5556:5556 \
  -v $(pwd)/data:/app/data \
  -e DB_TYPE=sqlite \
  -e ACCESS_TOKEN=your_secure_token \
  let-monitor-go

# 查看日志
docker logs -f let-monitor-go
```

### 从 GitHub Container Registry 拉取

```bash
# 拉取最新镜像
docker pull ghcr.io/<owner>/let-monitor-go:latest

# 运行
docker run -d \
  --name let-monitor-go \
  -p 5556:5556 \
  -v $(pwd)/data:/app/data \
  -e DB_TYPE=sqlite \
  -e ACCESS_TOKEN=your_secure_token \
  ghcr.io/<owner>/let-monitor-go:latest
```

---

## ⚙️ 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `DB_TYPE` | 数据库类型 (`sqlite` 或 `mongodb`) | `sqlite` | 否 |
| `SQLITE_PATH` | SQLite数据库文件路径 | `data/forum_monitor.db` | 否 |
| `MONGO_HOST` | MongoDB连接字符串 | `mongodb://localhost:27017/` | 使用MongoDB时 |
| `ACCESS_TOKEN` | API访问令牌 | `your_access_token_here` | **是** |
| `PORT` | Web服务器端口 | `5556` | 否 |
| `GIN_MODE` | Gin框架模式 (`debug`/`release`) | `release` | 否 |
| `LOG_LEVEL` | 日志级别 (`debug`/`info`/`warn`/`error`) | `info` | 否 |

### 配置文件

应用启动时会从 `data/.env` 加载环境变量，配置文件位于 `data/config.json`。

---

## 🔧 GitHub Actions 配置

### 启用自动构建

1. **创建个人访问令牌（可选但推荐）**
   - 访问: Settings → Developer settings → Personal access tokens → Tokens (classic)
   - 创建新令牌，权限选择: `write:packages`, `read:packages`
   - 复制生成的令牌

2. **添加到仓库密钥**
   - 仓库 Settings → Secrets and variables → Actions
   - 点击 "New repository secret"
   - Name: `GHCR_PAT`
   - Secret: 粘贴你的个人访问令牌

3. **触发构建**
   - 推送代码到 `main` 分支自动触发
   - 或在 Actions 标签页手动触发

### 工作流说明

```yaml
# .github/workflows/publish.yml
# - 在推送到main分支时自动运行
# - 支持手动触发 (workflow_dispatch)
# - 构建并推送到 ghcr.io/<owner>/let-monitor-go:latest
```

---

## 📊 数据持久化

### 目录结构
```
data/
├── .env                    # 环境变量配置
├── config.json            # 应用配置文件
├── forum_monitor.db       # SQLite数据库（SQLite模式）
└── db/                    # MongoDB数据（MongoDB模式）
```

### 备份建议
```bash
# SQLite备份
cp data/forum_monitor.db data/forum_monitor.db.backup

# MongoDB备份
docker-compose exec mongo mongodump --out /data/db/backup
```

---

## 🔍 故障排查

### 常见问题

**1. 容器无法启动**
```bash
# 检查日志
docker-compose logs app

# 检查配置文件
cat data/.env
```

**2. 数据库连接失败**
```bash
# SQLite: 检查文件权限
ls -la data/

# MongoDB: 检查服务状态
docker-compose ps mongo
docker-compose logs mongo
```

**3. Web界面无法访问**
```bash
# 检查端口映射
docker-compose ps
netstat -tuln | grep 5556

# 检查ACCESS_TOKEN配置
grep ACCESS_TOKEN data/.env
```

**4. GitHub Actions构建失败**
- 检查 Dockerfile 语法
- 验证 go.mod 和 go.sum 是否同步
- 查看 Actions 日志详细错误信息

---

## 🔄 升级指南

### 从旧版本升级

```bash
# 1. 停止现有服务
docker-compose down

# 2. 备份数据
cp -r data data.backup

# 3. 拉取最新代码
git pull origin main

# 4. 重新构建并启动
docker-compose build --no-cache
docker-compose up -d

# 5. 验证服务
docker-compose logs -f app
```

---

## 📝 技术细节

### Dockerfile 构建优化
- **多阶段构建**: 使用builder阶段编译，最终镜像仅包含运行时文件
- **CGO启用**: 安装gcc、musl-dev、sqlite-dev支持SQLite
- **静态链接**: 减少运行时依赖，提高可移植性
- **Alpine基础镜像**: 最小化镜像大小

### 构建命令解析
```bash
# CGO_ENABLED=1: 启用CGO支持SQLite
# -ldflags '-linkmode external -extldflags "-static"': 静态链接
CGO_ENABLED=1 GOOS=linux go build -a -ldflags '-linkmode external -extldflags "-static"' -o let-monitor-go ./cmd/app
```

---

## 📞 支持

如遇问题，请提交 Issue 并包含：
- 操作系统和Docker版本
- docker-compose.yml 配置
- 相关错误日志
- 重现步骤

---

**最后更新**: 2025-11-25  
**版本**: v2.0.0