# 数据库配置指南

let-monitor-go 支持两种数据库：**MongoDB** 和 **SQLite**。您可以根据需求选择合适的数据库。

## 数据库对比

| 特性 | MongoDB | SQLite |
|------|---------|--------|
| **部署复杂度** | 需要独立服务 | 零配置，单文件 |
| **性能** | 高并发优秀 | 单线程写入 |
| **扩展性** | 支持分布式 | 单机使用 |
| **资源占用** | 较高 | 极低 |
| **适用场景** | 生产环境、大规模 | 个人使用、测试开发 |
| **备份** | 需要专门工具 | 直接复制文件 |

## 使用 MongoDB

### Docker Compose 方式

```yaml
# docker-compose.yml
services:
  app:
    environment:
      - DB_TYPE=mongodb
      - MONGO_HOST=mongodb://mongo:27017/
    depends_on:
      - mongo

  mongo:
    image: mongo:7
    volumes:
      - ./data/db:/data/db
```

### 本地开发

```bash
# 1. 启动 MongoDB
docker run -d -p 27017:27017 -v $(pwd)/data/db:/data/db mongo:7

# 2. 配置环境变量
export DB_TYPE=mongodb
export MONGO_HOST=mongodb://localhost:27017/

# 3. 运行应用
go run cmd/app/main.go
```

## 使用 SQLite

### Docker Compose 方式

```yaml
# docker-compose.yml
services:
  app:
    environment:
      - DB_TYPE=sqlite
      - SQLITE_PATH=data/forum_monitor.db
    volumes:
      - ./data:/app/data
    # 移除 depends_on: mongo

# 注释或删除 mongo 服务
```

### 本地开发

```bash
# 1. 配置环境变量
export DB_TYPE=sqlite
export SQLITE_PATH=data/forum_monitor.db

# 2. 确保数据目录存在
mkdir -p data

# 3. 运行应用（自动创建数据库）
go run cmd/app/main.go
```

## 数据库切换

### 从 MongoDB 切换到 SQLite

1. **导出数据**（可选）:
   ```bash
   # 使用 mongoexport 导出数据
   mongoexport --db=forum_monitor --collection=threads --out=threads.json
   mongoexport --db=forum_monitor --collection=comments --out=comments.json
   ```

2. **修改配置**:
   ```bash
   # .env
   DB_TYPE=sqlite
   SQLITE_PATH=data/forum_monitor.db
   ```

3. **重启应用**:
   ```bash
   docker compose restart app
   ```

### 从 SQLite 切换到 MongoDB

1. **备份 SQLite 数据**:
   ```bash
   cp data/forum_monitor.db data/forum_monitor.db.backup
   ```

2. **启动 MongoDB**:
   ```bash
   docker compose up -d mongo
   ```

3. **修改配置**:
   ```bash
   # .env
   DB_TYPE=mongodb
   MONGO_HOST=mongodb://mongo:27017/
   ```

4. **重启应用**:
   ```bash
   docker compose restart app
   ```

## 数据库文件位置

### MongoDB
- 数据目录: `./data/db/`
- 默认端口: `27017`

### SQLite
- 数据文件: `./data/forum_monitor.db`
- 自动创建表和索引

## 备份策略

### MongoDB 备份
```bash
# 完整备份
mongodump --db=forum_monitor --out=./backup

# 恢复
mongorestore --db=forum_monitor ./backup/forum_monitor
```

### SQLite 备份
```bash
# 简单复制文件
cp data/forum_monitor.db backup/forum_monitor_$(date +%Y%m%d).db

# 使用 sqlite3 工具
sqlite3 data/forum_monitor.db ".backup 'backup/forum_monitor.db'"
```

## 性能优化建议

### MongoDB
- 使用索引（已自动创建）
- 启用副本集保证高可用
- 配置合适的连接池大小

### SQLite
- 设置 WAL 模式（已自动）
- 定期 VACUUM 清理
- 避免大量并发写入
- 定期备份数据文件

## 故障排查

### MongoDB 连接失败
```bash
# 检查 MongoDB 是否运行
docker ps | grep mongo

# 查看日志
docker logs let-monitor-go-mongo-1

# 测试连接
mongosh mongodb://localhost:27017/
```

### SQLite 锁定错误
```bash
# 检查文件权限
ls -la data/forum_monitor.db

# 确保没有其他进程占用
lsof data/forum_monitor.db

# 修复数据库
sqlite3 data/forum_monitor.db "PRAGMA integrity_check;"
```

## 常见问题

**Q: 可以在运行时切换数据库吗？**  
A: 不可以。需要停止应用，修改配置，然后重启。

**Q: SQLite 和 MongoDB 数据可以互相迁移吗？**  
A: 数据结构兼容，但需要编写迁移脚本。建议从一开始就选择合适的数据库。

**Q: 哪个数据库更推荐？**  
A: 
- 个人使用、轻量部署 → **SQLite**
- 生产环境、高并发 → **MongoDB**

**Q: SQLite 支持多少数据量？**  
A: SQLite 理论上支持 281TB，对于论坛监控场景（数万条记录）完全足够。