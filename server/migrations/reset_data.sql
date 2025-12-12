-- ============================================
-- 5SecondsGo 数据重置脚本
-- 用于清空所有业务数据，保留表结构和用户基本信息
-- ============================================

BEGIN;

-- 1. 清零所有用户的余额字段
UPDATE users SET 
    balance = 0,
    frozen_balance = 0,
    owner_room_balance = 0,
    owner_withdrawable_balance = 0,
    owner_frozen_balance = 0,
    owner_margin_balance = 0,
    owner_custody_quota = 0,
    consecutive_wins = 0,
    total_rounds = 0,
    total_wins = 0,
    last_win_at = NULL
WHERE role IN ('player', 'owner');

-- 2. 清空交易流水
TRUNCATE TABLE balance_transactions RESTART IDENTITY CASCADE;

-- 3. 清空资金申请
TRUNCATE TABLE fund_requests RESTART IDENTITY CASCADE;

-- 4. 清空游戏回合
TRUNCATE TABLE game_rounds RESTART IDENTITY CASCADE;

-- 5. 清空对账历史
TRUNCATE TABLE fund_conservation_history RESTART IDENTITY CASCADE;

-- 6. 清空告警
TRUNCATE TABLE alerts RESTART IDENTITY CASCADE;

-- 7. 清空风控标记
TRUNCATE TABLE risk_flags RESTART IDENTITY CASCADE;

-- 8. 清空监控快照
TRUNCATE TABLE metrics_snapshots RESTART IDENTITY CASCADE;

-- 9. 清空聊天消息
TRUNCATE TABLE chat_messages RESTART IDENTITY CASCADE;

-- 10. 清零平台账户余额
UPDATE platform_account SET 
    platform_balance = 0,
    balance = 0;

-- 11. 清空房间玩家（可选，如果需要保留房间配置）
TRUNCATE TABLE room_players RESTART IDENTITY CASCADE;
TRUNCATE TABLE room_spectators RESTART IDENTITY CASCADE;

-- 12. 清空邀请相关
TRUNCATE TABLE room_invitations RESTART IDENTITY CASCADE;
TRUNCATE TABLE invite_links RESTART IDENTITY CASCADE;

-- 13. 清空好友关系（可选）
-- TRUNCATE TABLE friends RESTART IDENTITY CASCADE;
-- TRUNCATE TABLE friend_requests RESTART IDENTITY CASCADE;

COMMIT;

-- 验证清空结果
SELECT 'users_balance' as check_item, SUM(balance + frozen_balance + owner_room_balance + owner_margin_balance) as value FROM users
UNION ALL SELECT 'platform_balance', platform_balance FROM platform_account
UNION ALL SELECT 'balance_transactions', COUNT(*)::numeric FROM balance_transactions
UNION ALL SELECT 'fund_requests', COUNT(*)::numeric FROM fund_requests
UNION ALL SELECT 'game_rounds', COUNT(*)::numeric FROM game_rounds
UNION ALL SELECT 'alerts', COUNT(*)::numeric FROM alerts
UNION ALL SELECT 'risk_flags', COUNT(*)::numeric FROM risk_flags;
