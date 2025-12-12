package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"

	"go.uber.org/zap"
)

// 聊天相关错误
var (
	ErrChatMessageTooLong = errors.New("chat message too long")
	ErrChatRateLimited    = errors.New("chat rate limited")
	ErrEmojiRateLimited   = errors.New("emoji rate limited")
	ErrInvalidEmoji       = errors.New("invalid emoji")
)

const (
	MaxChatMessageLength = 200
	ChatRateLimitPerSec  = 1
	EmojiRateLimitPerSec = 3
	MaxChatHistory       = 100
)

// ContentFilter 内容过滤器
type ContentFilter struct {
	prohibitedWords []string
}

// NewContentFilter 创建内容过滤器
func NewContentFilter() *ContentFilter {
	return &ContentFilter{
		prohibitedWords: []string{
			// 可以从配置或数据库加载敏感词
			"敏感词1", "敏感词2",
		},
	}
}

// Filter 过滤敏感内容
func (f *ContentFilter) Filter(content string) string {
	result := content
	for _, word := range f.prohibitedWords {
		if strings.Contains(result, word) {
			replacement := strings.Repeat("*", utf8.RuneCountInString(word))
			result = strings.ReplaceAll(result, word, replacement)
		}
	}
	return result
}

// RateLimiter 限流器
type RateLimiter struct {
	mu          sync.RWMutex
	limits      map[string][]time.Time // key -> timestamps
	maxCount    int
	window      time.Duration
	lastCleanup time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(maxCount int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		limits:      make(map[string][]time.Time),
		maxCount:    maxCount,
		window:      window,
		lastCleanup: time.Now(),
	}
	// 启动定期清理协程
	go rl.periodicCleanup()
	return rl
}

// periodicCleanup 定期清理过期记录
func (r *RateLimiter) periodicCleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		r.cleanup()
	}
}

// cleanup 清理所有过期记录
func (r *RateLimiter) cleanup() {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window * 2) // 保留2倍窗口时间的记录

	for key, timestamps := range r.limits {
		var valid []time.Time
		for _, t := range timestamps {
			if t.After(cutoff) {
				valid = append(valid, t)
			}
		}
		if len(valid) == 0 {
			delete(r.limits, key)
		} else {
			r.limits[key] = valid
		}
	}
	r.lastCleanup = now
}

// Allow 检查是否允许操作
func (r *RateLimiter) Allow(key string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-r.window)

	// 清理过期记录
	timestamps := r.limits[key]
	var valid []time.Time
	for _, t := range timestamps {
		if t.After(cutoff) {
			valid = append(valid, t)
		}
	}

	// 检查是否超限
	if len(valid) >= r.maxCount {
		r.limits[key] = valid
		return false
	}

	// 记录本次操作
	valid = append(valid, now)
	r.limits[key] = valid
	return true
}

// ChatService 聊天服务
type ChatService struct {
	repo            *repository.ChatRepo
	filter          *ContentFilter
	chatRateLimiter *RateLimiter
	emojiRateLimiter *RateLimiter
	logger          *zap.Logger
}

// NewChatService 创建聊天服务
func NewChatService(repo *repository.ChatRepo, logger *zap.Logger) *ChatService {
	return &ChatService{
		repo:            repo,
		filter:          NewContentFilter(),
		chatRateLimiter: NewRateLimiter(ChatRateLimitPerSec, time.Second),
		emojiRateLimiter: NewRateLimiter(EmojiRateLimitPerSec, time.Second),
		logger:          logger,
	}
}


// SendMessage 发送聊天消息
func (s *ChatService) SendMessage(ctx context.Context, roomID, userID int64, username, content string) (*model.ChatMessage, error) {
	// 检查限流
	key := fmt.Sprintf("chat:%d", userID)
	if !s.chatRateLimiter.Allow(key) {
		return nil, ErrChatRateLimited
	}

	// 截断消息
	if utf8.RuneCountInString(content) > MaxChatMessageLength {
		runes := []rune(content)
		content = string(runes[:MaxChatMessageLength])
	}

	// 过滤敏感内容
	content = s.filter.Filter(content)

	// 保存消息
	msg := &model.ChatMessage{
		RoomID:   roomID,
		UserID:   userID,
		Username: username,
		Content:  content,
	}

	if err := s.repo.Create(ctx, msg); err != nil {
		s.logger.Error("Failed to save chat message", zap.Error(err))
		return nil, err
	}

	// 异步清理旧消息
	go func() {
		if err := s.repo.DeleteOldMessages(context.Background(), roomID, MaxChatHistory); err != nil {
			s.logger.Warn("Failed to delete old messages", zap.Error(err))
		}
	}()

	return msg, nil
}

// GetHistory 获取聊天历史
func (s *ChatService) GetHistory(ctx context.Context, roomID int64, limit int) ([]*model.ChatMessage, error) {
	if limit <= 0 || limit > MaxChatHistory {
		limit = MaxChatHistory
	}
	return s.repo.GetHistory(ctx, roomID, limit)
}

// ValidateEmoji 验证表情
func (s *ChatService) ValidateEmoji(emoji string) error {
	if !model.IsValidEmoji(emoji) {
		return ErrInvalidEmoji
	}
	return nil
}

// CheckEmojiRateLimit 检查表情限流
func (s *ChatService) CheckEmojiRateLimit(userID int64) error {
	key := fmt.Sprintf("emoji:%d", userID)
	if !s.emojiRateLimiter.Allow(key) {
		return ErrEmojiRateLimited
	}
	return nil
}
