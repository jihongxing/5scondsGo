# 5SecondsGo - Go 服务端

5SecondsGo 游戏的 Go 后端服务，提供 REST API 和 WebSocket 实时通信。

## 功能特性

- 🎮 游戏引擎 - 5阶段循环游戏逻辑
- 🎲 可验证随机 - Commit-Reveal 算法
- 🔌 WebSocket - 实时双向通信
- 💰 资金管理 - 余额、冻结、结算
- 🛡️ 风控系统 - 异常检测和标记
- 📊 监控指标 - Prometheus metrics
- 📝 结构化日志 - Zap logger

## 环境要求

- Go 1.21+
- PostgreSQL 15+
- Redis 7+

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 配置

复制并修改配置文件:

```bash
cp config/config.yaml config/config.local.yaml
```

主要配置项:

```yaml
server:
  host: 0.0.0.0
  port: 8080
  mode: debug  # debug/release

database:
  host: localhost
  port: 5450
  user: fiveseconds
  password: fiveseconds123
  dbname: fiveseconds

redis:
  host: localhost
  port: 6450

auth:
  jwt_secret: "change-this-in-production"
```

### 3. 初始化数据库

```bash
# 使用 psql
psql -h localhost -p 5450 -U fiveseconds -d fiveseconds -f migrations/init.sql

# 或使用 Docker
docker exec -i postgres psql -U fiveseconds -d fiveseconds < migrations/init.sql
```

### 4. 启动服务

```bash
go run cmd/server/main.go
```

## 项目结构

```
server/
├── cmd/
│   ├── server/           # 主服务入口
│   │   └── main.go
│   ├── testbot/          # 测试机器人
│   │   └── main.go
│   └── genhash/          # 密码哈希工具
│       └── main.go
│
├── config/
│   └── config.yaml       # 配置文件
│
├── internal/
│   ├── config/           # 配置加载
│   │   └── config.go
│   │
│   ├── handler/          # HTTP/WS 处理器
│   │   ├── handler.go       # 主路由
│   │   ├── middleware.go    # 中间件
│   │   ├── ws_handler.go    # WebSocket
│   │   ├── wallet_handler.go
│   │   ├── friend_handler.go
│   │   └── ...
│   │
│   ├── service/          # 业务逻辑
│   │   ├── auth_service.go
│   │   ├── room_service.go
│   │   ├── wallet_service.go
│   │   ├── friend_service.go
│   │   ├── risk_service.go
│   │   └── ...
│   │
│   ├── repository/       # 数据访问
│   │   ├── db.go            # 数据库连接
│   │   ├── user_repo.go
│   │   ├── room_repo.go
│   │   └── ...
│   │
│   ├── model/            # 数据模型
│   │   ├── user.go
│   │   ├── room.go
│   │   ├── game.go
│   │   └── ...
│   │
│   ├── game/             # 游戏引擎
│   │   ├── manager.go       # 游戏管理器
│   │   ├── room_processor.go # 房间处理器
│   │   ├── random.go        # 随机算法
│   │   └── errors.go
│   │
│   ├── ws/               # WebSocket
│   │   └── hub.go           # 连接管理
│   │
│   ├── cache/            # 缓存层
│   │   ├── redis.go
│   │   └── balance_cache.go
│   │
│   └── middleware/       # 中间件
│       └── logging.go
│
├── migrations/           # 数据库迁移
│   ├── init.sql
│   └── ...
│
└── pkg/                  # 公共包
    ├── logger/          # 日志
    ├── metrics/         # 监控指标
    └── httpclient/      # HTTP 客户端
```

## API 文档

详细 API 文档请参考 [API-Reference.md](../docs/API-Reference.md)

### 主要接口

#### 认证
- `POST /api/register` - 用户注册
- `POST /api/login` - 用户登录

#### 房间
- `GET /api/rooms` - 房间列表
- `POST /api/rooms` - 创建房间
- `GET /api/rooms/:id` - 房间详情

#### 钱包
- `GET /api/wallet/balance` - 查询余额
- `GET /api/wallet/transactions` - 交易记录
- `POST /api/wallet/withdraw` - 提现申请

#### WebSocket
- `GET /ws?token=xxx` - WebSocket 连接

### WebSocket 消息

```json
// 加入房间
{"type": "join_room", "payload": {"room_id": 1}}

// 离开房间
{"type": "leave_room", "payload": {}}

// 发送聊天
{"type": "send_chat", "payload": {"content": "Hello"}}

// 发送表情
{"type": "send_emoji", "payload": {"emoji": "😀"}}
```

## 游戏引擎

### 游戏阶段

```
waiting → countdown → betting → in_game → settlement → waiting
   5s        5s          5s        5s         5s
```

### Commit-Reveal 随机算法

1. **Commit 阶段**: 服务器生成随机种子，计算哈希并广播
2. **Reveal 阶段**: 结算时公开种子，客户端可验证

```go
// 生成承诺
seed := generateRandomSeed()
commitHash := sha256(seed)

// 验证
isValid := sha256(revealSeed) == commitHash
```

## 监控指标

服务暴露 Prometheus 指标在 `:9091/metrics`:

| 指标 | 类型 | 描述 |
|------|------|------|
| game_rounds_total | Counter | 游戏轮次总数 |
| game_bets_total | Counter | 下注总数 |
| game_pool_amount | Gauge | 当前奖池金额 |
| ws_connections | Gauge | WebSocket 连接数 |
| http_requests_total | Counter | HTTP 请求总数 |
| http_request_duration | Histogram | 请求延迟 |

## 测试

```bash
# 单元测试
go test ./...

# 集成测试
go test ./internal/integration_test/...

# 覆盖率
go test -cover ./...
```

### 测试机器人

```bash
# 启动 5 个机器人加入房间 1
go run cmd/testbot/main.go -room 1 -bots 5 -interval 100
```

## 构建

```bash
# 构建可执行文件
go build -o server cmd/server/main.go

# 交叉编译 Linux
GOOS=linux GOARCH=amd64 go build -o server-linux cmd/server/main.go
```

## Docker

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go build -o server cmd/server/main.go

FROM alpine:latest
COPY --from=builder /app/server /server
COPY --from=builder /app/config /config
EXPOSE 8080 9091
CMD ["/server"]
```

## 性能优化

### 数据库连接池

```yaml
database:
  max_open_conns: 100
  max_idle_conns: 10
  conn_max_lifetime: 1h
```

### Redis 缓存

- 余额缓存: 减少数据库查询
- 乐观锁: 防止并发冲突

### WebSocket 优化

- 心跳检测: 30秒间隔
- 增量广播: 只发送变化数据
- 消息压缩: 大消息自动压缩

## 日志

使用 Zap 结构化日志:

```go
logger.Info("user joined room",
    zap.Int("user_id", userID),
    zap.Int("room_id", roomID),
)
```

日志级别: debug, info, warn, error

## 常见问题

### 数据库连接失败
- 检查 PostgreSQL 是否启动
- 检查配置文件中的连接信息
- 检查防火墙设置

### Redis 连接失败
- 检查 Redis 是否启动
- 检查端口和密码配置

### WebSocket 断开
- 检查心跳配置
- 检查代理超时设置
- 查看服务器日志
