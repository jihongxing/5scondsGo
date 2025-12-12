-- 5SecondsGo P1/P2 功能数据库迁移脚本
-- 版本: 2.0.0
-- 日期: 2025-12-08

-- ========================================
-- 1. 聊天消息表
-- ========================================
CREATE TABLE IF NOT EXISTS chat_messages (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    content     VARCHAR(200) NOT NULL,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_chat_room_time ON chat_messages(room_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chat_user ON chat_messages(user_id);

-- ========================================
-- 2. 好友关系表
-- ========================================
CREATE TABLE IF NOT EXISTS friends (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(user_id, friend_id)
);

CREATE INDEX IF NOT EXISTS idx_friends_user ON friends(user_id);
CREATE INDEX IF NOT EXISTS idx_friends_friend ON friends(friend_id);

-- ========================================
-- 3. 好友请求表
-- ========================================
CREATE TABLE IF NOT EXISTS friend_requests (
    id           BIGSERIAL PRIMARY KEY,
    from_user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    to_user_id   BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/accepted/rejected
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(from_user_id, to_user_id)
);

CREATE INDEX IF NOT EXISTS idx_friend_req_to ON friend_requests(to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_friend_req_from ON friend_requests(from_user_id);

-- ========================================
-- 4. 房间邀请表
-- ========================================
CREATE TABLE IF NOT EXISTS room_invitations (
    id           BIGSERIAL PRIMARY KEY,
    room_id      BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    from_user_id BIGINT NOT NULL REFERENCES users(id),
    to_user_id   BIGINT NOT NULL REFERENCES users(id),
    status       VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/accepted/declined/expired
    created_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at   TIMESTAMP NOT NULL
);

CREATE INDEX IF NOT EXISTS idx_invitation_to ON room_invitations(to_user_id, status);
CREATE INDEX IF NOT EXISTS idx_invitation_room ON room_invitations(room_id);

-- ========================================
-- 5. 邀请链接表
-- ========================================
CREATE TABLE IF NOT EXISTS invite_links (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    code        VARCHAR(32) UNIQUE NOT NULL,
    created_by  BIGINT NOT NULL REFERENCES users(id),
    created_at  TIMESTAMP NOT NULL DEFAULT NOW(),
    expires_at  TIMESTAMP NOT NULL,
    use_count   INT NOT NULL DEFAULT 0,
    max_uses    INT DEFAULT NULL  -- NULL表示无限制
);

CREATE INDEX IF NOT EXISTS idx_invite_code ON invite_links(code);
CREATE INDEX IF NOT EXISTS idx_invite_room ON invite_links(room_id);

-- ========================================
-- 6. 房间主题表
-- ========================================
CREATE TABLE IF NOT EXISTS room_themes (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT UNIQUE NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    theme_name  VARCHAR(50) NOT NULL DEFAULT 'classic',  -- classic/neon/ocean/forest/luxury
    updated_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_theme_room ON room_themes(room_id);

-- ========================================
-- 7. 风控标记表
-- ========================================
CREATE TABLE IF NOT EXISTS risk_flags (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    flag_type   VARCHAR(50) NOT NULL,  -- consecutive_wins/high_win_rate/multi_account/large_transaction
    details     JSONB,
    status      VARCHAR(20) NOT NULL DEFAULT 'pending',  -- pending/reviewed/confirmed/dismissed
    reviewed_by BIGINT REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_risk_user ON risk_flags(user_id);
CREATE INDEX IF NOT EXISTS idx_risk_status ON risk_flags(status);
CREATE INDEX IF NOT EXISTS idx_risk_type ON risk_flags(flag_type);

-- ========================================
-- 8. 告警记录表
-- ========================================
CREATE TABLE IF NOT EXISTS alerts (
    id              BIGSERIAL PRIMARY KEY,
    alert_type      VARCHAR(50) NOT NULL,  -- negative_balance/negative_custody/large_transaction/settlement_failure/conservation_failure
    severity        VARCHAR(20) NOT NULL,  -- info/warning/critical
    title           VARCHAR(200) NOT NULL,
    details         JSONB,
    related_user_id BIGINT REFERENCES users(id),
    related_room_id BIGINT REFERENCES rooms(id),
    status          VARCHAR(20) NOT NULL DEFAULT 'active',  -- active/acknowledged/resolved
    acknowledged_by BIGINT REFERENCES users(id),
    acknowledged_at TIMESTAMP,
    resolved_at     TIMESTAMP,
    created_at      TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_alerts_status ON alerts(status, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_alerts_type ON alerts(alert_type);
CREATE INDEX IF NOT EXISTS idx_alerts_severity ON alerts(severity);

-- ========================================
-- 9. 指标快照表
-- ========================================
CREATE TABLE IF NOT EXISTS metrics_snapshots (
    id                  BIGSERIAL PRIMARY KEY,
    online_players      INT NOT NULL DEFAULT 0,
    active_rooms        INT NOT NULL DEFAULT 0,
    games_per_minute    DECIMAL(10,2) NOT NULL DEFAULT 0,
    api_latency_p95     DECIMAL(10,2) NOT NULL DEFAULT 0,
    ws_latency_p95      DECIMAL(10,2) NOT NULL DEFAULT 0,
    db_latency_p95      DECIMAL(10,2) NOT NULL DEFAULT 0,
    daily_active_users  INT NOT NULL DEFAULT 0,
    daily_volume        DECIMAL(18,2) NOT NULL DEFAULT 0,
    platform_revenue    DECIMAL(18,2) NOT NULL DEFAULT 0,
    created_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_metrics_time ON metrics_snapshots(created_at DESC);

-- ========================================
-- 10. 房间观战者表
-- ========================================
CREATE TABLE IF NOT EXISTS room_spectators (
    id          BIGSERIAL PRIMARY KEY,
    room_id     BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    joined_at   TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_spectator_room ON room_spectators(room_id);
CREATE INDEX IF NOT EXISTS idx_spectator_user ON room_spectators(user_id);

-- ========================================
-- 11. 用户表扩展字段
-- ========================================
ALTER TABLE users ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(64);
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_version BIGINT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(10) NOT NULL DEFAULT 'zh';
ALTER TABLE users ADD COLUMN IF NOT EXISTS consecutive_wins INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_win_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_rounds INT NOT NULL DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS total_wins INT NOT NULL DEFAULT 0;

CREATE INDEX IF NOT EXISTS idx_users_fingerprint ON users(device_fingerprint);
CREATE INDEX IF NOT EXISTS idx_users_language ON users(language);

-- ========================================
-- 12. 房间表扩展字段
-- ========================================
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS max_spectators INT NOT NULL DEFAULT 50;
ALTER TABLE rooms ADD COLUMN IF NOT EXISTS name VARCHAR(100);

-- 更新现有房间的名称
UPDATE rooms SET name = 'Room ' || code WHERE name IS NULL;

-- ========================================
-- 13. 更新触发器
-- ========================================
DROP TRIGGER IF EXISTS friend_requests_updated_at ON friend_requests;
CREATE TRIGGER friend_requests_updated_at
    BEFORE UPDATE ON friend_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS room_themes_updated_at ON room_themes;
CREATE TRIGGER room_themes_updated_at
    BEFORE UPDATE ON room_themes
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
