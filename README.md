# 5SecondsGo - 多人在线高频押注小游戏

## 项目简介

5SecondsGo 是一款轻量级、高实时性的多人即时押注小游戏。玩家通过房主创建的房间参与游戏，每轮包含5个阶段（等待→倒计时→下注→游戏→结算），每阶段固定5秒，系统自动下注、自动结算，体验流畅刺激。

## 主要功能

### 核心游戏
- 🎮 实时多人押注游戏
- 🎲 Commit-Reveal 可验证随机算法
- 💰 自动下注与结算
- 📊 完整的资金流水记录

### P1 体验增强 (v2.0)
- 👀 **观战模式** - 观看游戏不参与下注，支持切换为参与者
- 💬 **房间聊天** - 实时聊天，支持敏感词过滤和限流
- 😀 **表情反应** - 12种预设表情快速互动
- 📜 **游戏记录** - 查看历史记录，支持回放验证
- 👥 **好友系统** - 添加好友，查看在线状态
- 📨 **邀请功能** - 邀请好友加入房间，支持邀请链接
- 💼 **钱包功能** - 完整的余额管理和收益统计
- 🎨 **房间主题** - 5种内置主题皮肤
- 🌐 **多语言** - 支持中/英/日/韩语

### P2 扩展优化 (v2.0)
- ⚡ **余额缓存** - Redis缓存提升性能，乐观锁防并发冲突
- 📡 **增量广播** - Phase Tick优化，减少网络带宽
- 🛡️ **风控系统** - 检测异常行为，自动标记审核
- 🚨 **告警系统** - 实时告警推送，快速响应问题
- 📈 **监控仪表盘** - 实时指标监控，历史趋势分析

## 技术栈

| 模块 | 技术 |
|------|------|
| 后端 | Go 1.21+ |
| 前端App | Flutter 3.16+ (iOS/Android) |
| 管理后台 | Flutter Web |
| 数据库 | PostgreSQL 15 |
| 缓存 | Redis 7 |
| 实时通信 | WebSocket |
| 日志 | Zap + Loki |
| 监控 | Prometheus + Grafana |

## 项目结构

```
5SecondsGo/
├── server/         # Go 后端
├── app/            # Flutter 移动端App
├── admin/          # Flutter Web 管理后台
├── docs/           # 文档
└── docker-compose.yml
```

## 快速开始

### 1. 启动基础服务

```bash
# 启动 PostgreSQL, Redis, Loki, Prometheus, Grafana
docker-compose up -d
```

### 2. 启动后端服务

```bash
cd server
go mod download
go run cmd/server/main.go
```

### 3. 启动前端App (开发模式)

```bash
cd app
flutter pub get
flutter run
```

### 4. 启动管理后台 (开发模式)

```bash
cd admin
flutter pub get
flutter run -d chrome
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Go Server | 8080 | HTTP + WebSocket |
| Go Metrics | 9091 | Prometheus metrics |
| PostgreSQL | 5432 | 数据库 |
| Redis | 6379 | 缓存 |
| Grafana | 3000 | 监控面板 |
| Prometheus | 9090 | 指标收集 |
| Loki | 3100 | 日志收集 |

## 默认账号

| 角色 | 用户名 | 密码 | 邀请码 |
|------|--------|------|--------|
| Admin | admin | admin123 | ADMIN1 |

## 开发文档

- [需求文档](docs/多人在线高频押注小游戏.md)
- [技术设计文档](docs/技术设计文档.md)
- [API参考文档](docs/API-Reference.md)
- [UI设计文档](docs/UI-Redesign-Card-Based.md)

## 功能使用说明

### 观战模式
1. 进入房间时选择"观战"按钮
2. 观战者可以看到所有游戏状态，但不参与下注
3. 点击"切换为参与者"可加入游戏（房间未满时）

### 聊天与表情
- 在房间内点击聊天图标发送消息
- 消息限制：200字符，每秒1条
- 点击表情按钮选择预设表情，每秒最多3次

### 好友系统
1. 在好友页面搜索用户名发送好友请求
2. 好友列表显示在线状态和当前房间
3. 点击好友可快速邀请加入房间

### 邀请功能
- 直接邀请：选择在线好友发送邀请
- 链接邀请：生成24小时有效的邀请链接分享

### 游戏记录
- 查看历史游戏记录，支持按日期/房间筛选
- 点击记录可查看详情和回放
- 使用 reveal_seed 验证游戏公平性

### 钱包功能
- 查看可用余额、冻结余额、总余额
- 查看交易历史和收益统计
- 发起提现申请

### 房间主题（房主）
- 在房间设置中选择主题
- 可选：Classic、Neon、Ocean、Forest、Luxury

### 管理后台功能
- **监控仪表盘**：实时查看在线人数、活跃房间、交易量等指标
- **风控管理**：审核被标记的可疑账户
- **告警中心**：查看和处理系统告警

## WebSocket 连接

```javascript
// 连接示例
const ws = new WebSocket('ws://localhost:8080/ws?token=YOUR_JWT_TOKEN');

// 加入房间
ws.send(JSON.stringify({type: 'join_room', payload: {room_id: 123}}));

// 发送聊天
ws.send(JSON.stringify({type: 'send_chat', payload: {content: 'Hello!'}}));

// 发送表情
ws.send(JSON.stringify({type: 'send_emoji', payload: {emoji: '😀'}}));
```

## 测试

```bash
# 运行后端测试
cd server
go test ./...

# 运行集成测试
go test ./internal/integration_test/...
```

## License

MIT
