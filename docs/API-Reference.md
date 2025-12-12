# 5SecondsGo API 参考文档

> 版本: V2.0 (P1/P2 功能)
> 更新日期: 2025-12-09

---

## 目录

1. [认证 API](#1-认证-api)
2. [房间 API](#2-房间-api)
3. [观战 API](#3-观战-api)
4. [好友 API](#4-好友-api)
5. [邀请 API](#5-邀请-api)
6. [游戏记录 API](#6-游戏记录-api)
7. [钱包 API](#7-钱包-api)
8. [主题 API](#8-主题-api)
9. [风控 API (Admin)](#9-风控-api-admin)
10. [监控 API (Admin)](#10-监控-api-admin)
11. [告警 API (Admin)](#11-告警-api-admin)
12. [WebSocket 事件](#12-websocket-事件)

---

## 1. 认证 API

### POST /api/auth/register
注册新用户（需要邀请码）

**请求体:**
```json
{
  "username": "string",
  "password": "string",
  "invite_code": "string"
}
```

**响应:**
```json
{
  "user_id": 123,
  "username": "string",
  "role": "player",
  "token": "jwt_token"
}
```

### POST /api/auth/login
用户登录

**请求体:**
```json
{
  "username": "string",
  "password": "string"
}
```

**响应:**
```json
{
  "user_id": 123,
  "username": "string",
  "role": "player",
  "token": "jwt_token"
}
```

### GET /api/auth/me
获取当前用户信息

**响应:**
```json
{
  "user_id": 123,
  "username": "string",
  "role": "player",
  "balance": "1000.00",
  "frozen_balance": "0.00",
  "language": "zh"
}
```

---

## 2. 房间 API

### POST /api/rooms
创建房间（房主）

**请求体:**
```json
{
  "bet_amount": 10,
  "min_players": 2,
  "max_players": 10,
  "winner_count": 1,
  "owner_commission": 0.03
}
```

### GET /api/rooms
获取房间列表

**查询参数:**
- `status`: 房间状态 (active/locked/closed)
- `page`: 页码
- `limit`: 每页数量

### GET /api/rooms/:code
获取房间详情

### POST /api/rooms/:code/join
加入房间

### POST /api/rooms/:code/leave
离开房间

---

## 3. 观战 API

### POST /api/rooms/:id/spectate
以观战者身份加入房间

**响应:**
```json
{
  "success": true,
  "room_id": 123,
  "spectator_count": 5
}
```

**错误码:**
- `5001`: 房间观战人数已满
- `5002`: 已是观战者

### POST /api/rooms/:id/switch-to-participant
观战者切换为参与者

**响应:**
```json
{
  "success": true,
  "room_id": 123
}
```

**错误码:**
- `5003`: 不是观战者
- `2002`: 房间已满

---

## 4. 好友 API

### GET /api/friends
获取好友列表

**响应:**
```json
{
  "friends": [
    {
      "user_id": 123,
      "username": "friend1",
      "is_online": true,
      "current_room_id": 456
    }
  ]
}
```

### POST /api/friends/request
发送好友请求

**请求体:**
```json
{
  "to_user_id": 123
}
```

**错误码:**
- `6001`: 好友请求已存在
- `6002`: 已是好友
- `6003`: 好友数量已达上限

### GET /api/friends/requests
获取待处理的好友请求

**响应:**
```json
{
  "requests": [
    {
      "request_id": 1,
      "from_user_id": 123,
      "from_username": "user1",
      "created_at": "2025-12-09T10:00:00Z"
    }
  ]
}
```

### POST /api/friends/accept
接受好友请求

**请求体:**
```json
{
  "request_id": 1
}
```

### POST /api/friends/reject
拒绝好友请求

**请求体:**
```json
{
  "request_id": 1
}
```

### DELETE /api/friends/:id
删除好友

---

## 5. 邀请 API

### POST /api/rooms/:id/invite
邀请好友加入房间

**请求体:**
```json
{
  "to_user_id": 123
}
```

**响应:**
```json
{
  "invitation_id": 1,
  "expires_at": "2025-12-09T11:00:00Z"
}
```

### POST /api/rooms/:id/invite-link
生成邀请链接

**响应:**
```json
{
  "code": "ABC123XY",
  "link": "https://app.5secondsgo.com/invite/ABC123XY",
  "expires_at": "2025-12-10T10:00:00Z"
}
```

### POST /api/invite/:code/join
通过邀请链接加入房间

**错误码:**
- `7001`: 邀请链接已过期
- `7002`: 邀请链接无效
- `2002`: 房间已满

---

## 6. 游戏记录 API

### GET /api/game-history
获取游戏历史记录

**查询参数:**
- `page`: 页码 (默认 1)
- `limit`: 每页数量 (默认 20, 最大 100)
- `room_id`: 按房间筛选
- `start_date`: 开始日期 (YYYY-MM-DD)
- `end_date`: 结束日期 (YYYY-MM-DD)

**响应:**
```json
{
  "records": [
    {
      "round_id": 123,
      "room_id": 1,
      "room_name": "测试房间",
      "round_number": 5,
      "bet_amount": "10.00",
      "result": "win",
      "prize_amount": "18.00",
      "created_at": "2025-12-09T10:00:00Z"
    }
  ],
  "total": 100,
  "page": 1,
  "limit": 20
}
```

### GET /api/game-history/:id
获取回合详情

**响应:**
```json
{
  "round_id": 123,
  "room_id": 1,
  "round_number": 5,
  "participants": [1, 2, 3, 4, 5],
  "winners": [2, 4],
  "bet_amount": "10.00",
  "pool_amount": "50.00",
  "prize_per_winner": "22.75",
  "commit_hash": "abc123...",
  "reveal_seed": "def456...",
  "created_at": "2025-12-09T10:00:00Z",
  "settled_at": "2025-12-09T10:00:25Z"
}
```

### GET /api/game-stats
获取游戏统计

**查询参数:**
- `period`: 统计周期 (day/week/month/all)

**响应:**
```json
{
  "total_rounds": 500,
  "win_count": 150,
  "lose_count": 350,
  "win_rate": 0.30,
  "total_wagered": "5000.00",
  "total_won": "2700.00",
  "net_profit": "-2300.00"
}
```

### GET /api/game-rounds/:id/replay
获取回放数据

**响应:**
```json
{
  "round_id": 123,
  "phases": [
    {
      "phase": "betting",
      "timestamp": 1699999990000,
      "data": {"pool_amount": "0.00"}
    },
    {
      "phase": "in_game",
      "timestamp": 1699999995000,
      "data": {"commit_hash": "abc123..."}
    },
    {
      "phase": "settlement",
      "timestamp": 1700000000000,
      "data": {"winners": [2, 4], "reveal_seed": "def456..."}
    }
  ],
  "verification": {
    "commit_hash": "abc123...",
    "reveal_seed": "def456...",
    "computed_winners": [2, 4],
    "is_valid": true
  }
}
```

---

## 7. 钱包 API

### GET /api/wallet
获取钱包信息

**响应:**
```json
{
  "available_balance": "1000.00",
  "frozen_balance": "100.00",
  "total_balance": "1100.00"
}
```

### GET /api/wallet/transactions
获取交易历史

**查询参数:**
- `page`: 页码
- `limit`: 每页数量
- `type`: 交易类型 (deposit/withdraw/bet/win/lose)

**响应:**
```json
{
  "transactions": [
    {
      "id": 1,
      "type": "win",
      "amount": "18.00",
      "balance_after": "1018.00",
      "created_at": "2025-12-09T10:00:00Z",
      "remark": "房间#123 第5轮获胜"
    }
  ],
  "total": 50,
  "page": 1
}
```

### GET /api/wallet/earnings
获取收益统计

**查询参数:**
- `period`: 统计周期 (day/week/month)

**响应:**
```json
{
  "period": "week",
  "total_winnings": "500.00",
  "total_losses": "300.00",
  "net_profit": "200.00",
  "daily_breakdown": [
    {"date": "2025-12-09", "profit": "50.00"},
    {"date": "2025-12-08", "profit": "-20.00"}
  ]
}
```

### POST /api/wallet/withdraw
发起提现申请

**请求体:**
```json
{
  "amount": "100.00",
  "payment_account": "支付宝: xxx@xxx.com"
}
```

---

## 8. 主题 API

### GET /api/themes
获取可用主题列表

**响应:**
```json
{
  "themes": [
    {
      "name": "classic",
      "display_name": "经典",
      "primary_color": "#1976D2",
      "background_color": "#FFFFFF"
    },
    {
      "name": "neon",
      "display_name": "霓虹",
      "primary_color": "#FF00FF",
      "background_color": "#000000"
    }
  ]
}
```

### PUT /api/rooms/:id/theme
设置房间主题（房主）

**请求体:**
```json
{
  "theme_name": "neon"
}
```

---

## 9. 风控 API (Admin)

### GET /api/admin/risk-flags
获取风控标记列表

**查询参数:**
- `status`: 状态 (pending/reviewed/confirmed/dismissed)
- `type`: 类型 (consecutive_wins/high_win_rate/multi_account)

**响应:**
```json
{
  "flags": [
    {
      "id": 1,
      "user_id": 123,
      "username": "user1",
      "flag_type": "consecutive_wins",
      "details": {"consecutive_count": 12},
      "status": "pending",
      "created_at": "2025-12-09T10:00:00Z"
    }
  ]
}
```

### POST /api/admin/risk-flags/:id/review
审核风控标记

**请求体:**
```json
{
  "action": "confirm",
  "suspend_account": true,
  "freeze_balance": true,
  "remark": "确认作弊行为"
}
```

---

## 10. 监控 API (Admin)

### GET /api/admin/metrics/realtime
获取实时指标

**响应:**
```json
{
  "online_players": 1500,
  "active_rooms": 120,
  "games_per_minute": 45.5,
  "api_latency_p95": 25.3,
  "ws_latency_p95": 8.2,
  "db_latency_p95": 5.1,
  "daily_active_users": 3500,
  "daily_volume": "125000.00",
  "platform_revenue": "2500.00",
  "timestamp": 1699999999000
}
```

### GET /api/admin/metrics/history
获取历史指标

**查询参数:**
- `range`: 时间范围 (1h/24h/7d/30d)

**响应:**
```json
{
  "snapshots": [
    {
      "timestamp": 1699999999000,
      "online_players": 1500,
      "active_rooms": 120,
      "games_per_minute": 45.5
    }
  ]
}
```

---

## 11. 告警 API (Admin)

### GET /api/admin/alerts
获取告警列表

**查询参数:**
- `status`: 状态 (active/acknowledged)
- `severity`: 严重程度 (info/warning/critical)

**响应:**
```json
{
  "alerts": [
    {
      "id": 1,
      "alert_type": "negative_balance",
      "severity": "critical",
      "title": "玩家余额为负",
      "details": {"user_id": 123, "balance": "-50.00"},
      "status": "active",
      "created_at": "2025-12-09T10:00:00Z"
    }
  ]
}
```

### POST /api/admin/alerts/:id/acknowledge
确认告警

**请求体:**
```json
{
  "remark": "已处理"
}
```

---

## 12. WebSocket 事件

### 连接
```
ws://server:8080/ws?token=<jwt_token>
```

### 客户端 → 服务端

| 事件类型 | 描述 | 载荷 |
|----------|------|------|
| `heartbeat` | 心跳 | `{}` |
| `join_room` | 加入房间 | `{room_id}` |
| `leave_room` | 离开房间 | `{}` |
| `set_auto_ready` | 设置自动准备 | `{auto_ready: bool}` |
| `join_as_spectator` | 以观战者加入 | `{room_id}` |
| `switch_to_participant` | 切换为参与者 | `{}` |
| `send_chat` | 发送聊天 | `{content}` |
| `send_emoji` | 发送表情 | `{emoji}` |
| `send_invite` | 发送邀请 | `{room_id, to_user_id}` |
| `respond_invite` | 响应邀请 | `{invitation_id, accept}` |

### 服务端 → 客户端

| 事件类型 | 描述 | 载荷 |
|----------|------|------|
| `error` | 错误 | `{code, message}` |
| `room_state` | 房间完整状态 | `{room_id, phase, players, spectators, ...}` |
| `phase_change` | 阶段变化 | `{phase, phase_end_time, round}` |
| `phase_tick` | 增量状态更新 | `{server_time, time_remaining, ...}` |
| `player_join` | 玩家加入 | `{user_id, username}` |
| `player_leave` | 玩家离开 | `{user_id}` |
| `player_update` | 玩家状态更新 | `{user_id, balance, auto_ready}` |
| `betting_done` | 下注完成 | `{pool_amount, participants, skipped}` |
| `round_result` | 回合结果 | `{round_id, winners, prize_per_winner, reveal_seed}` |
| `round_failed` | 回合失败 | `{reason, refunded}` |
| `balance_update` | 余额更新 | `{balance, frozen_balance}` |
| `spectator_join` | 观战者加入 | `{user_id, username}` |
| `spectator_leave` | 观战者离开 | `{user_id}` |
| `spectator_switch` | 观战者切换 | `{user_id}` |
| `chat_message` | 聊天消息 | `{id, user_id, username, content, timestamp}` |
| `chat_history` | 聊天历史 | `{messages: [...]}` |
| `emoji_reaction` | 表情反应 | `{user_id, username, emoji}` |
| `friend_request` | 好友请求 | `{request_id, from_user_id, from_username}` |
| `friend_accepted` | 好友接受 | `{friend_id, friend_name}` |
| `friend_online` | 好友上线 | `{friend_id, friend_name, room_id}` |
| `friend_offline` | 好友下线 | `{friend_id}` |
| `room_invitation` | 房间邀请 | `{invitation_id, room_id, room_name, ...}` |
| `invite_response` | 邀请响应 | `{invitation_id, accepted, from_user_id}` |
| `theme_change` | 主题变更 | `{room_id, theme_name}` |
| `alert` | 告警 (Admin) | `{id, type, severity, title, details}` |
| `metrics_update` | 指标更新 (Admin) | `{online_players, active_rooms, ...}` |

---

## 错误码汇总

| 错误码 | 中文 | English |
|--------|------|---------|
| 1001 | 用户名已存在 | Username already exists |
| 1002 | 邀请码无效 | Invalid invite code |
| 1003 | 用户名或密码错误 | Invalid username or password |
| 1004 | 账户已被禁用 | Account disabled |
| 2001 | 房间不存在 | Room not found |
| 2002 | 房间已满 | Room is full |
| 2003 | 房间已锁定 | Room is locked |
| 2004 | 无权操作此房间 | No permission for this room |
| 3001 | 余额不足 | Insufficient balance |
| 3002 | 托管额度不足 | Insufficient custody quota |
| 3003 | 保证金不足 | Insufficient margin balance |
| 3004 | 风险超限 | Risk limit exceeded |
| 5001 | 房间观战人数已满 | Room spectator limit reached |
| 5002 | 已是观战者 | Already a spectator |
| 5003 | 不是观战者 | Not a spectator |
| 5004 | 聊天消息过长 | Chat message too long |
| 5005 | 聊天频率过快 | Chat rate limit exceeded |
| 5006 | 表情发送过快 | Emoji rate limit exceeded |
| 6001 | 好友请求已存在 | Friend request already exists |
| 6002 | 已是好友 | Already friends |
| 6003 | 好友数量已达上限 | Friend limit reached |
| 6004 | 好友请求不存在 | Friend request not found |
| 7001 | 邀请链接已过期 | Invite link expired |
| 7002 | 邀请链接无效 | Invalid invite link |
| 8001 | 不支持的语言 | Unsupported language |
| 9001 | 账户已被标记 | Account flagged for review |
| 9002 | 账户已被冻结 | Account frozen |
