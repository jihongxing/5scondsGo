package game

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	mrand "math/rand"
)

// CommitReveal 实现可验证随机算法
type CommitReveal struct{}

func NewCommitReveal() *CommitReveal {
	return &CommitReveal{}
}

// GenerateCommit 生成随机种子和 commit hash
// 返回: (seed, commitHash, error)
func (cr *CommitReveal) GenerateCommit() ([]byte, string, error) {
	// 生成32字节随机种子
	seed := make([]byte, 32)
	if _, err := rand.Read(seed); err != nil {
		return nil, "", fmt.Errorf("generate seed: %w", err)
	}

	// 计算 SHA-256 hash 作为 commit
	hash := sha256.Sum256(seed)
	commitHash := hex.EncodeToString(hash[:])

	return seed, commitHash, nil
}

// Reveal 公开种子并验证
// 返回种子的十六进制字符串
func (cr *CommitReveal) Reveal(seed []byte) string {
	return hex.EncodeToString(seed)
}

// Verify 验证 reveal 是否匹配 commit
func (cr *CommitReveal) Verify(revealSeed, commitHash string) bool {
	seedBytes, err := hex.DecodeString(revealSeed)
	if err != nil {
		return false
	}
	hash := sha256.Sum256(seedBytes)
	return hex.EncodeToString(hash[:]) == commitHash
}

// SelectWinners 使用种子确定性选择赢家
// playerIDs: 参与者ID列表
// winnerCount: 需要选出的赢家数量
// seed: 随机种子
func (cr *CommitReveal) SelectWinners(playerIDs []int64, winnerCount int, seed []byte) []int64 {
	if len(playerIDs) == 0 || winnerCount <= 0 {
		return nil
	}

	// 如果赢家数量大于等于参与者数量,所有人都是赢家
	if winnerCount >= len(playerIDs) {
		result := make([]int64, len(playerIDs))
		copy(result, playerIDs)
		return result
	}

	// 使用种子创建确定性随机数生成器
	// 使用完整的32字节种子生成多个int64，提高随机性
	rng := mrand.New(newSeedSource(seed))

	// Fisher-Yates 洗牌,但只洗前 winnerCount 个
	shuffled := make([]int64, len(playerIDs))
	copy(shuffled, playerIDs)

	for i := 0; i < winnerCount; i++ {
		j := i + rng.Intn(len(shuffled)-i)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	return shuffled[:winnerCount]
}

// seedSource 使用完整种子的随机源
type seedSource struct {
	seed  []byte
	index int
}

// newSeedSource 创建新的种子源
func newSeedSource(seed []byte) *seedSource {
	return &seedSource{seed: seed, index: 0}
}

// Int63 实现 rand.Source 接口
func (s *seedSource) Int63() int64 {
	// 使用种子的不同部分生成随机数
	// 通过 SHA256 哈希种子+索引来生成更多随机数
	h := sha256.Sum256(append(s.seed, byte(s.index), byte(s.index>>8)))
	s.index++

	var result int64
	for i := 0; i < 8; i++ {
		result = (result << 8) | int64(h[i])
	}
	// 确保返回非负数
	return result & 0x7FFFFFFFFFFFFFFF
}

// Seed 实现 rand.Source 接口（不使用）
func (s *seedSource) Seed(seed int64) {
	// 不支持重新设置种子
}

// GenerateInviteCode 生成6位字母数字邀请码
func GenerateInviteCode() (string, error) {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	code := make([]byte, 6)
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", err
	}
	for i := range code {
		code[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(code), nil
}
