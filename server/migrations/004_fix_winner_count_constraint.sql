-- 修复 winner_count 约束
-- 原约束: winner_count < min_players (过于严格)
-- 新约束: winner_count < max_players (更合理)

ALTER TABLE rooms DROP CONSTRAINT IF EXISTS chk_winner_count;
ALTER TABLE rooms ADD CONSTRAINT chk_winner_count CHECK (winner_count >= 1 AND winner_count < max_players);
