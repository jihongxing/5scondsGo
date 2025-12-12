-- 添加 balance_version 列用于乐观锁
-- 版本: 2.0.0

-- 为 users 表添加 balance_version 列
ALTER TABLE users ADD COLUMN IF NOT EXISTS balance_version BIGINT NOT NULL DEFAULT 0;

-- 为 users 表添加语言偏好列
ALTER TABLE users ADD COLUMN IF NOT EXISTS language VARCHAR(10) NOT NULL DEFAULT 'zh';

-- 为 users 表添加设备指纹列
ALTER TABLE users ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(255);

-- 创建设备指纹索引
CREATE INDEX IF NOT EXISTS idx_users_device_fingerprint ON users(device_fingerprint);
