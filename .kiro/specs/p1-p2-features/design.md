# Design Document: P1-P2 Features

## Overview

本设计文档描述 5SecondsGo 游戏的 P1（体验增强）和 P2（扩展优化）功能的技术实现方案。这些功能将在现有 MVP 架构基础上进行扩展，保持系统的高性能和可维护性。

## Architecture

### 系统架构扩展

```
┌─────────────────────────────────────────────────────────────────┐
│                        Flutter App / Admin                       │
├─────────────────────────────────────────────────────────────────┤
│  新增模块:                                                        │
│  - ChatWidget (聊天组件)                                          │
│  - EmojiPicker (表情选择器)                                       │
│  - GameHistoryPage (游戏记录页)                                   │
│  - FriendListPage (好友列表页)                                    │
│  - WalletPage (钱包页)                                            │
│  - ThemeSelector (主题选择器)                                     │
│  - MonitoringDashboard (监控仪表盘)                               │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                        Go Backend                                │
├─────────────────────────────────────────────────────────────────┤
│  新增服务:                                                        │
│  - ChatService (聊天服务)                                         │
│  - FriendService (好友服务)                                       │
│  - GameHistoryService (游戏记录服务)                              │
│  - MonitoringService (监控服务)                                   │
│  - RiskControlService (风控服务)                                  │
│  - ThemeService (主题服务)                                        │
│  - NotificationService (通知服务)                                 │
├─────────────────────────────────────────────────────────────────┤
│  扩展模块:                                                        │
│  - RoomProcessor (添加观战者、聊天、表情支持)                      │
│  - BalanceCache (余额缓存层)                                      │
│  - AlertManager (告警管理器)                                      │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────────┐
│                     Data Layer                                   │
├─────────────────────────────────────────────────────────────────┤
│  新增表:                                                          │
│  - chat_messages (聊天消息)                                       │
│  - friends (好友关系)                                             │
│  - friend_requests (好友请求)                                     │
│  - room_invitations (房间邀请)                                    │
│  - invite_links (邀请链接)                                        │
│  - room_themes (房间主题)                                         │
│  - risk_flags (风控标记)                                          │
│  - alerts (告警记录)                                              │
│  - metrics_snapshots (指标快照)                                   │
├─────────────────────────────────────────────────────────────────┤
│  Redis 扩展:                                                      │
│  - balance_cache:{user_id} (余额缓存)                             │
│  - online_status:{user_id} (在线状态)                             │
│  - rate_limit:chat:{user_id} (聊天限流)                           │
│  - rate_limit:emoji:{user_id} (表情限流)                          │
│  - metrics:realtime (实时指标)                                    │
└─────────────────────────────────────────────────────────────────┘
```

## Components and Interfaces

### 1. 观战模式组件

```go
// SpectatorManager 观战者管理
type SpectatorManager struct {
    spectators map[int64]*SpectatorState  // userID -> state
    maxSpectators int                      // 最大观战人数 (50)
}

type SpectatorState struct {
    UserID    int64
    Username  string
    JoinedAt  time.Time
}

// RoomProcessor 扩展
func (rp *RoomProcessor) AddSpectator(user *model.User) error
func (rp *RoomProcessor) RemoveSpectator(userID int64)
func (rp *RoomProcessor) SpectatorToParticipant(userID int64) error
func (rp *RoomProcessor) GetSpectatorCount() int
```

### 2. 聊天服务组件

```go
// ChatService 聊天服务
type ChatService struct {
    repo        *repository.ChatRepo
    broadcaster Broadcaster
    filter      *ContentFilter
    rateLimiter *RateLimiter
}

type ChatMessage struct {
    ID        int64
    RoomID    int64
    UserID    int64
    Username  string
    Content   string
    CreatedAt time.Time
}

func (s *ChatService) SendMessage(ctx context.Context, roomID, userID int64, content string) error
func (s *ChatService) GetHistory(ctx context.Context, roomID int64, limit int) ([]*ChatMessage, error)
```

### 3. 好友服务组件

```go
// FriendService 好友服务
type FriendService struct {
    repo         *repository.FriendRepo
    notification *NotificationService
}

type FriendRequest struct {
    ID         int64
    FromUserID int64
    ToUserID   int64
    Status     string  // pending/accepted/rejected
    CreatedAt  time.Time
}

func (s *FriendService) SendRequest(ctx context.Context, fromID, toID int64) error
func (s *FriendService) AcceptRequest(ctx context.Context, requestID int64) error
func (s *FriendService) RejectRequest(ctx context.Context, requestID int64) error
func (s *FriendService) RemoveFriend(ctx context.Context, userID, friendID int64) error
func (s *FriendService) GetFriendList(ctx context.Context, userID int64) ([]*Friend, error)
```

### 4. 监控服务组件

```go
// MonitoringService 监控服务
type MonitoringService struct {
    metricsRepo *repository.MetricsRepo
    redis       *redis.Client
}

type RealtimeMetrics struct {
    OnlinePlayers    int
    ActiveRooms      int
    GamesPerMinute   float64
    APILatencyP95    float64
    WSLatencyP95     float64
    DBLatencyP95     float64
    DailyActiveUsers int
    DailyVolume      decimal.Decimal
    PlatformRevenue  decimal.Decimal
}

func (s *MonitoringService) GetRealtimeMetrics(ctx context.Context) (*RealtimeMetrics, error)
func (s *MonitoringService) GetHistoricalMetrics(ctx context.Context, from, to time.Time) ([]*MetricsSnapshot, error)
```

### 5. 风控服务组件

```go
// RiskControlService 风控服务
type RiskControlService struct {
    repo         *repository.RiskRepo
    alertManager *AlertManager
}

type RiskFlag struct {
    ID        int64
    UserID    int64
    Type      string  // consecutive_wins/high_win_rate/multi_account
    Details   string
    Status    string  // pending/reviewed/confirmed/dismissed
    CreatedAt time.Time
}

func (s *RiskControlService) CheckConsecutiveWins(ctx context.Context, userID int64) error
func (s *RiskControlService) CheckWinRate(ctx context.Context, userID int64) error
func (s *RiskControlService) CheckDeviceFingerprint(ctx context.Context, fingerprint string) error
func (s *RiskControlService) ReviewFlag(ctx context.Context, flagID int64, action string) error
```

### 6. 余额缓存组件

```go
// BalanceCache 余额缓存
type BalanceCache struct {
    redis *redis.Client
    repo  *repository.UserRepo
}

type CachedBalance struct {
    Balance        decimal.Decimal
    FrozenBalance  decimal.Decimal
    Version        int64
    CachedAt       time.Time
}

func (c *BalanceCache) Get(ctx context.Context, userID int64) (*CachedBalance, error)
func (c *BalanceCache) Set(ctx context.Context, userID int64, balance *CachedBalance) error
func (c *BalanceCache) Invalidate(ctx context.Context, userID int64) error
func (c *BalanceCache) UpdateWithVersion(ctx context.Context, userID int64, delta decimal.Decimal, expectedVersion int64) error
```

## Data Models

### 新增数据库表

```sql
-- 聊天消息表
CREATE TABLE chat_messages (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id),
    user_id     BIGINT NOT NULL REFERENCES users(id),
    content     VARCHAR(200) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_chat_room_time ON chat_messages(room_id, created_at DESC);

-- 好友关系表
CREATE TABLE friends (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    friend_id   BIGINT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, friend_id)
);
CREATE INDEX idx_friends_user ON friends(user_id);

-- 好友请求表
CREATE TABLE friend_requests (
    id           BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES users(id),
    to_user_id   BIGINT NOT NULL REFERENCES users(id),
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_friend_req_to ON friend_requests(to_user_id, status);

-- 房间邀请表
CREATE TABLE room_invitations (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id),
    from_user_id BIGINT NOT NULL REFERENCES users(id),
    to_user_id  BIGINT NOT NULL REFERENCES users(id),
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMP NOT NULL
);
CREATE INDEX idx_invitation_to ON room_invitations(to_user_id, status);

-- 邀请链接表
CREATE TABLE invite_links (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id),
    code        VARCHAR(32) UNIQUE NOT NULL,
    created_by  BIGINT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMP NOT NULL,
    use_count   INT NOT NULL DEFAULT 0
);
CREATE INDEX idx_invite_code ON invite_links(code);

-- 房间主题表
CREATE TABLE room_themes (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT UNIQUE NOT NULL REFERENCES rooms(id),
    theme_name  VARCHAR(50) NOT NULL DEFAULT 'classic',
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 风控标记表
CREATE TABLE risk_flags (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    flag_type   VARCHAR(50) NOT NULL,
    details     JSONB,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',
    reviewed_by BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_risk_user ON risk_flags(user_id);
CREATE INDEX idx_risk_status ON risk_flags(status);

-- 告警记录表
CREATE TABLE alerts (
    id          BIGSERIAL PRIMARY KEY,
    alert_type  VARCHAR(50) NOT NULL,
    severity    VARCHAR(20) NOT NULL,  -- info/warning/critical
    title       VARCHAR(200) NOT NULL,
    details     JSONB,
    status      VARCHAR(20) NOT NULL DEFAULT 'active',
    acknowledged_by BIGINT REFERENCES users(id),
    acknowledged_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_alerts_status ON alerts(status, created_at DESC);

-- 指标快照表
CREATE TABLE metrics_snapshots (
    id                  BIGSERIAL PRIMARY KEY,
    online_players      INT NOT NULL,
    active_rooms        INT NOT NULL,
    games_per_minute    DECIMAL(10,2) NOT NULL,
    api_latency_p95     DECIMAL(10,2) NOT NULL,
    ws_latency_p95      DECIMAL(10,2) NOT NULL,
    db_latency_p95      DECIMAL(10,2) NOT NULL,
    daily_active_users  INT NOT NULL,
    daily_volume        DECIMAL(18,2) NOT NULL,
    platform_revenue    DECIMAL(18,2) NOT NULL,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_metrics_time ON metrics_snapshots(created_at DESC);

-- 用户表扩展
ALTER TABLE users ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(10) NOT NULL DEFAULT 'zh';
ALTER TABLE users ADD COLUMN IF NOT EXISTS consecutive_wins INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_win_at TIMESTAMP;

CREATE INDEX IF NOT EXISTS idx_users_fingerprint ON users(device_fingerprint);
```

### WebSocket 事件扩展

```go
// 新增 WebSocket 消息类型
const (
    // 观战
    WSTypeSpectatorJoin   = "spectator_join"
    WSTypeSpectatorLeave  = "spectator_leave"
    WSTypeSpectatorSwitch = "spectator_switch"
    
    // 聊天
    WSTypeChatMessage     = "chat_message"
    WSTypeChatHistory     = "chat_history"
    
    // 表情
    WSTypeEmojiReaction   = "emoji_reaction"
    
    // 好友
    WSTypeFriendRequest   = "friend_request"
    WSTypeFriendAccepted  = "friend_accepted"
    WSTypeFriendOnline    = "friend_online"
    WSTypeFriendOffline   = "friend_offline"
    
    // 邀请
    WSTypeRoomInvitation  = "room_invitation"
    WSTypeInviteResponse  = "invite_response"
    
    // 主题
    WSTypeThemeChange     = "theme_change"
    
    // 告警 (Admin)
    WSTypeAlert           = "alert"
    WSTypeMetricsUpdate   = "metrics_update"
)
```

## Correctness Properties

*A property is a characteristic or behavior that should hold true across all valid executions of a system-essentially, a formal statement about what the system should do. Properties serve as the bridge between human-readable specifications and machine-verifiable correctness guarantees.*

### Property 1: Spectator isolation
*For any* room and any spectator, the spectator should never appear in the participants list and should never have their balance deducted during betting phase
**Validates: Requirements 1.3**

### Property 2: Spectator receives all updates
*For any* room with spectators, when a phase change or game event occurs, all spectators should receive the same event data as participants
**Validates: Requirements 1.2**

### Property 3: Chat message broadcast completeness
*For any* chat message sent in a room, all online members (participants and spectators) should receive the message
**Validates: Requirements 2.1**

### Property 4: Chat message truncation
*For any* chat message with length greater than 200 characters, the stored and broadcast message should have exactly 200 characters
**Validates: Requirements 2.3**

### Property 5: Chat rate limiting
*For any* user sending messages faster than 1 per second, messages beyond the first should be rejected
**Validates: Requirements 2.4**

### Property 6: Emoji rate limiting
*For any* user sending emoji reactions faster than 3 per second, reactions beyond the third should be ignored
**Validates: Requirements 3.3**

### Property 7: Game history pagination
*For any* game history query, the returned records should be sorted by time descending and limited to the specified page size
**Validates: Requirements 4.1**

### Property 8: Game history date filtering
*For any* game history query with date range filter, all returned records should have timestamps within the specified range
**Validates: Requirements 4.3**

### Property 9: Round verification consistency
*For any* completed game round, computing winners using the revealed seed and participant list should produce the same winner list as stored
**Validates: Requirements 5.4**

### Property 10: Friend relationship bidirectionality
*For any* accepted friend request, both users should appear in each other's friend list
**Validates: Requirements 6.2**

### Property 11: Friend removal bidirectionality
*For any* friend removal operation, neither user should appear in the other's friend list after removal
**Validates: Requirements 6.5**

### Property 12: Invitation notification delivery
*For any* room invitation sent, the target user should receive a notification containing room name, bet amount, and player count
**Validates: Requirements 7.2**

### Property 13: Invite link validity
*For any* invite link, it should be usable only within 24 hours of creation and only if the room has available space
**Validates: Requirements 7.5, 7.6**

### Property 14: Balance cache consistency
*For any* balance-changing transaction, the cache should reflect the same value as the database after the transaction completes
**Validates: Requirements 10.2**

### Property 15: Optimistic locking prevents conflicts
*For any* concurrent balance updates with the same expected version, only one should succeed and the other should fail
**Validates: Requirements 10.6**

### Property 16: Consecutive win detection
*For any* player who wins more than 10 consecutive rounds, a risk flag should be created
**Validates: Requirements 12.1**

### Property 17: Win rate detection
*For any* player with win rate exceeding 80% over 50 rounds, a risk flag should be created
**Validates: Requirements 12.2**

### Property 18: Negative balance alert
*For any* transaction that would result in negative player balance, an alert should be triggered
**Validates: Requirements 13.1**

### Property 19: Large transaction alert
*For any* single transaction exceeding 10000, an alert should be triggered
**Validates: Requirements 13.3**

### Property 20: Theme persistence
*For any* room with a selected theme, players joining the room should receive the theme information
**Validates: Requirements 14.3**

### Property 21: Wallet balance accuracy
*For any* wallet view, the displayed total balance should equal available balance plus frozen balance
**Validates: Requirements 15.1**

### Property 22: Earnings calculation accuracy
*For any* earnings summary, net profit should equal total winnings minus total losses
**Validates: Requirements 15.5**

## Error Handling

### 错误码扩展

| 错误码 | 中文 | English |
|--------|------|---------|
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

### 降级策略

1. **聊天服务降级**: 当聊天服务不可用时，禁用聊天功能但不影响游戏进行
2. **好友服务降级**: 当好友服务不可用时，显示离线状态但保留好友列表
3. **监控服务降级**: 当监控服务不可用时，使用本地缓存的最后一次指标
4. **余额缓存降级**: 当 Redis 不可用时，直接读取数据库

## Testing Strategy

### 单元测试

- 测试聊天消息截断逻辑
- 测试限流器行为
- 测试好友关系的双向性
- 测试余额缓存的一致性
- 测试风控检测算法
- 测试告警触发条件

### 属性测试

使用 `github.com/leanovate/gopter` 进行属性测试：

1. **观战者隔离测试**: 生成随机房间状态，验证观战者永远不参与下注
2. **聊天截断测试**: 生成随机长度字符串，验证截断行为
3. **限流测试**: 生成随机时间序列的请求，验证限流行为
4. **好友双向性测试**: 生成随机好友操作序列，验证关系一致性
5. **缓存一致性测试**: 生成随机余额操作，验证缓存与数据库一致
6. **风控检测测试**: 生成随机游戏结果序列，验证检测触发
7. **告警触发测试**: 生成随机交易，验证告警条件

### 集成测试

- WebSocket 消息广播测试
- 多用户并发操作测试
- 数据库事务一致性测试
- Redis 缓存失效测试

### 性能测试

- 聊天消息广播延迟测试 (目标 < 100ms)
- 余额缓存命中率测试 (目标 > 95%)
- 监控指标采集延迟测试
- 大量观战者场景压力测试
