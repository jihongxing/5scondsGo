-- 5SecondsGo 数据库初始化脚本
-- 版本: 1.0.0

-- ========================================
-- 1. 用户表
-- ========================================
CREATE TABLE IF NOT EXISTS users (
    id                          BIGSERIAL PRIMARY KEY,
    username                    VARCHAR(50) UNIQUE NOT NULL,
    password_hash               VARCHAR(255) NOT NULL,
    role                        VARCHAR(20) NOT NULL DEFAULT 'player',  -- admin/owner/player
    invited_by                  BIGINT REFERENCES users(id),            -- 邀请人(房主/admin)
    invite_code                 VARCHAR(6) UNIQUE,                      -- 房主/admin的邀请码
    
    -- 玩家余额
    balance                     DECIMAL(18,2) NOT NULL DEFAULT 0,
    frozen_balance              DECIMAL(18,2) NOT NULL DEFAULT 0,
    
    -- 房主专属字段
    owner_room_balance          DECIMAL(18,2) NOT NULL DEFAULT 0,
    owner_withdrawable_balance  DECIMAL(18,2) NOT NULL DEFAULT 0,
    owner_frozen_balance        DECIMAL(18,2) NOT NULL DEFAULT 0,
    owner_custody_quota         DECIMAL(18,2) NOT NULL DEFAULT 0,
    owner_margin_balance        DECIMAL(18,2) NOT NULL DEFAULT 0,
    
    status                      VARCHAR(20) NOT NULL DEFAULT 'active',  -- active/disabled
    created_at                  TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at                  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_users_role ON users(role);
CREATE INDEX IF NOT EXISTS idx_users_invited_by ON users(invited_by);
CREATE INDEX IF NOT EXISTS idx_users_invite_code ON users(invite_code);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);

-- ========================================
-- 2. 房间表
-- ========================================
CREATE TABLE IF NOT EXISTS rooms (
    id                      BIGSERIAL PRIMARY KEY,
    code                    VARCHAR(10) UNIQUE NOT NULL,
    owner_id                BIGINT NOT NULL REFERENCES users(id),
    
    -- 房间配置
    bet_amount              DECIMAL(18,2) NOT NULL,
    min_players             INT NOT NULL DEFAULT 2,
    max_players             INT NOT NULL DEFAULT 10,
    winner_count            INT NOT NULL DEFAULT 1,
    owner_commission        DECIMAL(5,4) NOT NULL DEFAULT 0.03,
    platform_commission     DECIMAL(5,4) NOT NULL DEFAULT 0.02,
    password                VARCHAR(50),
    
    -- 状态
    status                  VARCHAR(20) NOT NULL DEFAULT 'active',
    current_round           INT NOT NULL DEFAULT 0,
    state_json              JSONB,
    
    created_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at              TIMESTAMP NOT NULL DEFAULT NOW(),
    
    -- 约束
    CONSTRAINT chk_bet_amount CHECK (bet_amount IN (5, 10, 20, 50, 100, 200)),
    CONSTRAINT chk_min_players CHECK (min_players >= 2 AND min_players <= 100),
    CONSTRAINT chk_max_players CHECK (max_players >= min_players AND max_players <= 100),
    CONSTRAINT chk_winner_count CHECK (winner_count >= 1 AND winner_count < max_players),
    CONSTRAINT chk_commission CHECK (owner_commission + platform_commission <= 0.10)
);

CREATE INDEX IF NOT EXISTS idx_rooms_owner ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_rooms_code ON rooms(code);
CREATE INDEX IF NOT EXISTS idx_rooms_status ON rooms(status);

-- ========================================
-- 3. 游戏回合表
-- ========================================
CREATE TABLE IF NOT EXISTS game_rounds (
    id                  BIGSERIAL PRIMARY KEY,
    room_id             BIGINT NOT NULL REFERENCES rooms(id),
    round_number        INT NOT NULL,
    
    -- 参与信息
    participant_ids     BIGINT[] NOT NULL DEFAULT '{}',
    skipped_ids         BIGINT[] NOT NULL DEFAULT '{}',
    winner_ids          BIGINT[] NOT NULL DEFAULT '{}',
    
    -- 金额
    bet_amount          DECIMAL(18,2) NOT NULL,
    pool_amount         DECIMAL(18,2) NOT NULL DEFAULT 0,
    prize_per_winner    DECIMAL(18,2),
    owner_earning       DECIMAL(18,2),
    platform_earning    DECIMAL(18,2),
    residual_amount     DECIMAL(18,2) DEFAULT 0,
    
    -- Commit-Reveal 随机
    commit_hash         VARCHAR(64),
    reveal_seed         VARCHAR(64),
    
    -- 状态
    status              VARCHAR(20) NOT NULL DEFAULT 'betting',
    failure_reason      VARCHAR(100),
    
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    settled_at          TIMESTAMP,
    
    UNIQUE(room_id, round_number)
);

CREATE INDEX IF NOT EXISTS idx_rounds_room ON game_rounds(room_id);
CREATE INDEX IF NOT EXISTS idx_rounds_status ON game_rounds(status);
CREATE INDEX IF NOT EXISTS idx_rounds_created ON game_rounds(created_at);

-- ========================================
-- 4. 账本流水表
-- ========================================
CREATE TABLE IF NOT EXISTS balance_transactions (
    id                  BIGSERIAL PRIMARY KEY,
    tx_type             VARCHAR(30) NOT NULL,
    user_id             BIGINT NOT NULL REFERENCES users(id),
    operator_id         BIGINT REFERENCES users(id),
    room_id             BIGINT REFERENCES rooms(id),
    round_id            BIGINT REFERENCES game_rounds(id),
    
    amount              DECIMAL(18,2) NOT NULL,
    balance_before      DECIMAL(18,2) NOT NULL,
    balance_after       DECIMAL(18,2) NOT NULL,
    balance_field       VARCHAR(50) NOT NULL,
    
    remark              VARCHAR(500),
    created_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tx_user ON balance_transactions(user_id);
CREATE INDEX IF NOT EXISTS idx_tx_room ON balance_transactions(room_id);
CREATE INDEX IF NOT EXISTS idx_tx_round ON balance_transactions(round_id);
CREATE INDEX IF NOT EXISTS idx_tx_type ON balance_transactions(tx_type);
CREATE INDEX IF NOT EXISTS idx_tx_created ON balance_transactions(created_at);

-- ========================================
-- 5. 资金申请表
-- ========================================
CREATE TABLE IF NOT EXISTS fund_requests (
    id                  BIGSERIAL PRIMARY KEY,
    request_type        VARCHAR(30) NOT NULL,
    user_id             BIGINT NOT NULL REFERENCES users(id),
    owner_id            BIGINT REFERENCES users(id),
    
    amount              DECIMAL(18,2) NOT NULL,
    status              VARCHAR(20) NOT NULL DEFAULT 'pending',
    
    payment_account     VARCHAR(200),
    remark              VARCHAR(500),
    
    operator_id         BIGINT REFERENCES users(id),
    created_at          TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fund_user ON fund_requests(user_id);
CREATE INDEX IF NOT EXISTS idx_fund_owner ON fund_requests(owner_id);
CREATE INDEX IF NOT EXISTS idx_fund_status ON fund_requests(status);
CREATE INDEX IF NOT EXISTS idx_fund_type ON fund_requests(request_type);

-- ========================================
-- 6. 平台账户表
-- ========================================
CREATE TABLE IF NOT EXISTS platform_account (
    id                  BIGSERIAL PRIMARY KEY,
    balance             DECIMAL(18,2) NOT NULL DEFAULT 0,
    updated_at          TIMESTAMP NOT NULL DEFAULT NOW()
);

-- ========================================
-- 7. 房间玩家关联表
-- ========================================
CREATE TABLE IF NOT EXISTS room_players (
    id                  BIGSERIAL PRIMARY KEY,
    room_id             BIGINT NOT NULL REFERENCES rooms(id),
    user_id             BIGINT NOT NULL REFERENCES users(id),
    
    auto_ready          BOOLEAN NOT NULL DEFAULT FALSE,
    is_online           BOOLEAN NOT NULL DEFAULT FALSE,
    joined_at           TIMESTAMP NOT NULL DEFAULT NOW(),
    
    UNIQUE(room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_rp_room ON room_players(room_id);
CREATE INDEX IF NOT EXISTS idx_rp_user ON room_players(user_id);

-- ========================================
-- 8. 资金对账历史表
-- ========================================
CREATE TABLE IF NOT EXISTS fund_conservation_history (
    id                          BIGSERIAL PRIMARY KEY,
    scope                       VARCHAR(20) NOT NULL,              -- global/owner
    owner_id                    BIGINT REFERENCES users(id),       -- scope=owner 时必填
    period_type                 VARCHAR(20) NOT NULL,              -- 2h/daily
    period_start                TIMESTAMP NOT NULL,
    period_end                  TIMESTAMP NOT NULL,

    -- 聚合字段
    total_player_balance        DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 名下玩家余额合计
    total_player_frozen         DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 名下玩家冻结余额合计
    total_custody_quota         DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 房主托管额度合计
    total_margin                DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 保证金合计
    owner_room_balance          DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 房主房间收益
    owner_withdrawable_balance  DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 房主可提现
    owner_frozen_balance        DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 房主冻结
    platform_balance            DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 平台账户余额

    difference                  DECIMAL(18,2) NOT NULL DEFAULT 0,  -- 左右两侧差值
    is_balanced                 BOOLEAN NOT NULL,                  -- 是否通过对账

    created_at                  TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_fch_scope_created ON fund_conservation_history(scope, created_at);
CREATE INDEX IF NOT EXISTS idx_fch_owner_created ON fund_conservation_history(owner_id, created_at);
CREATE INDEX IF NOT EXISTS idx_fch_period ON fund_conservation_history(period_type, period_start);

-- ========================================
-- 初始化数据
-- ========================================

-- 初始化平台账户
INSERT INTO platform_account (id, balance) VALUES (1, 0)
ON CONFLICT (id) DO NOTHING;

-- 初始化 Admin 账号
-- 密码: admin123 (bcrypt hash)
INSERT INTO users (username, password_hash, role, invite_code, status)
VALUES ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMy.Mrq4RQmGkT3LT8VvHoYFIxqiDwLBnRa', 'admin', 'ADMIN1', 'active')
ON CONFLICT (username) DO NOTHING;

-- ========================================
-- 更新触发器
-- ========================================
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS users_updated_at ON users;
CREATE TRIGGER users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS rooms_updated_at ON rooms;
CREATE TRIGGER rooms_updated_at
    BEFORE UPDATE ON rooms
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS fund_requests_updated_at ON fund_requests;
CREATE TRIGGER fund_requests_updated_at
    BEFORE UPDATE ON fund_requests
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();

DROP TRIGGER IF EXISTS platform_account_updated_at ON platform_account;
CREATE TRIGGER platform_account_updated_at
    BEFORE UPDATE ON platform_account
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at();
