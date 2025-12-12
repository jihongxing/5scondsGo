-- 修复 platform_account 表的列名
-- 将 balance 改为 platform_balance 以匹配代码

ALTER TABLE platform_account RENAME COLUMN balance TO platform_balance;
