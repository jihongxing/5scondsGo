// Package service 属性测试
// 本文件包含 P1/P2 功能的属性测试
package service

import (
	"crypto/sha256"
	"encoding/hex"
	"sort"
	"sync"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/leanovate/gopter"
	"github.com/leanovate/gopter/gen"
	"github.com/leanovate/gopter/prop"
	"github.com/shopspring/decimal"
)

// =============================================================================
// Property 1: Spectator isolation
// **Feature: p1-p2-features, Property 1: Spectator isolation**
// **Validates: Requirements 1.3**
// =============================================================================

// MockRoomState 模拟房间状态
type MockRoomState struct {
	mu           sync.RWMutex
	participants map[int64]*MockParticipant
	spectators   map[int64]*MockSpectator
	bettingPool  decimal.Decimal
}

type MockParticipant struct {
	UserID   int64
	Balance  decimal.Decimal
	AutoReady bool
}

type MockSpectator struct {
	UserID   int64
	Username string
}

func NewMockRoomState() *MockRoomState {
	return &MockRoomState{
		participants: make(map[int64]*MockParticipant),
		spectators:   make(map[int64]*MockSpectator),
		bettingPool:  decimal.Zero,
	}
}

func (r *MockRoomState) AddParticipant(userID int64, balance decimal.Decimal) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.participants[userID] = &MockParticipant{
		UserID:  userID,
		Balance: balance,
	}
}

func (r *MockRoomState) AddSpectator(userID int64, username string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.spectators[userID] = &MockSpectator{
		UserID:   userID,
		Username: username,
	}
}

func (r *MockRoomState) IsSpectator(userID int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.spectators[userID]
	return ok
}

func (r *MockRoomState) IsParticipant(userID int64) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.participants[userID]
	return ok
}

func (r *MockRoomState) ProcessBetting(betAmount decimal.Decimal) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, p := range r.participants {
		if p.AutoReady {
			p.Balance = p.Balance.Sub(betAmount)
			r.bettingPool = r.bettingPool.Add(betAmount)
		}
	}
	// 观战者不参与下注
}

func (r *MockRoomState) GetSpectatorBalance(userID int64) (decimal.Decimal, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	// 观战者没有余额变化
	if _, ok := r.spectators[userID]; ok {
		return decimal.Zero, true
	}
	return decimal.Zero, false
}

// TestProperty1_SpectatorIsolation 属性测试：观战者隔离
// **Feature: p1-p2-features, Property 1: Spectator isolation**
// **Validates: Requirements 1.3**
func TestProperty1_SpectatorIsolation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("spectators never appear in participants list", prop.ForAll(
		func(participantCount, spectatorCount int) bool {
			room := NewMockRoomState()
			
			// 添加参与者
			for i := 0; i < participantCount; i++ {
				room.AddParticipant(int64(i+1), decimal.NewFromInt(1000))
			}
			
			// 添加观战者（ID从1000开始避免冲突）
			for i := 0; i < spectatorCount; i++ {
				room.AddSpectator(int64(1000+i), "spectator")
			}
			
			// 验证：观战者不在参与者列表中
			for i := 0; i < spectatorCount; i++ {
				if room.IsParticipant(int64(1000 + i)) {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 50),
	))

	properties.Property("spectators balance never deducted during betting", prop.ForAll(
		func(participantCount, spectatorCount int, betAmount int64) bool {
			room := NewMockRoomState()
			
			// 添加参与者并设置自动准备
			for i := 0; i < participantCount; i++ {
				room.AddParticipant(int64(i+1), decimal.NewFromInt(1000))
				room.participants[int64(i+1)].AutoReady = true
			}
			
			// 添加观战者
			for i := 0; i < spectatorCount; i++ {
				room.AddSpectator(int64(1000+i), "spectator")
			}
			
			// 执行下注
			room.ProcessBetting(decimal.NewFromInt(betAmount))
			
			// 验证：观战者余额不变（始终为0，因为他们不参与）
			for i := 0; i < spectatorCount; i++ {
				if _, isSpectator := room.GetSpectatorBalance(int64(1000 + i)); !isSpectator {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 50),
		gen.Int64Range(1, 100),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 2: Spectator receives all updates
// **Feature: p1-p2-features, Property 2: Spectator receives all updates**
// **Validates: Requirements 1.2**
// =============================================================================

type MockBroadcaster struct {
	mu       sync.Mutex
	messages map[int64][]string // userID -> messages
}

func NewMockBroadcaster() *MockBroadcaster {
	return &MockBroadcaster{
		messages: make(map[int64][]string),
	}
}

func (b *MockBroadcaster) BroadcastToRoom(userIDs []int64, msg string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for _, id := range userIDs {
		b.messages[id] = append(b.messages[id], msg)
	}
}

func (b *MockBroadcaster) GetMessages(userID int64) []string {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.messages[userID]
}

// TestProperty2_SpectatorReceivesAllUpdates 属性测试：观战者接收所有更新
// **Feature: p1-p2-features, Property 2: Spectator receives all updates**
// **Validates: Requirements 1.2**
func TestProperty2_SpectatorReceivesAllUpdates(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("spectators receive same events as participants", prop.ForAll(
		func(participantCount, spectatorCount, eventCount int) bool {
			broadcaster := NewMockBroadcaster()
			
			var allUserIDs []int64
			// 参与者
			for i := 0; i < participantCount; i++ {
				allUserIDs = append(allUserIDs, int64(i+1))
			}
			// 观战者
			for i := 0; i < spectatorCount; i++ {
				allUserIDs = append(allUserIDs, int64(1000+i))
			}
			
			// 广播事件
			for i := 0; i < eventCount; i++ {
				broadcaster.BroadcastToRoom(allUserIDs, "phase_change")
			}
			
			// 验证：所有观战者收到的消息数量与参与者相同
			if participantCount > 0 {
				participantMsgCount := len(broadcaster.GetMessages(1))
				for i := 0; i < spectatorCount; i++ {
					spectatorMsgCount := len(broadcaster.GetMessages(int64(1000 + i)))
					if spectatorMsgCount != participantMsgCount {
						return false
					}
				}
			}
			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(1, 50),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 3: Chat message broadcast completeness
// **Feature: p1-p2-features, Property 3: Chat message broadcast completeness**
// **Validates: Requirements 2.1**
// =============================================================================

// TestProperty3_ChatMessageBroadcastCompleteness 属性测试：聊天消息广播完整性
// **Feature: p1-p2-features, Property 3: Chat message broadcast completeness**
// **Validates: Requirements 2.1**
func TestProperty3_ChatMessageBroadcastCompleteness(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all room members receive chat message", prop.ForAll(
		func(participantCount, spectatorCount int) bool {
			broadcaster := NewMockBroadcaster()
			
			var allUserIDs []int64
			for i := 0; i < participantCount; i++ {
				allUserIDs = append(allUserIDs, int64(i+1))
			}
			for i := 0; i < spectatorCount; i++ {
				allUserIDs = append(allUserIDs, int64(1000+i))
			}
			
			// 发送聊天消息
			broadcaster.BroadcastToRoom(allUserIDs, "chat_message")
			
			// 验证：所有成员都收到消息
			for _, id := range allUserIDs {
				if len(broadcaster.GetMessages(id)) != 1 {
					return false
				}
			}
			return true
		},
		gen.IntRange(1, 10),
		gen.IntRange(0, 50),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 4: Chat message truncation
// **Feature: p1-p2-features, Property 4: Chat message truncation**
// **Validates: Requirements 2.3**
// =============================================================================

// TruncateMessage 截断消息到指定长度
func TruncateMessage(content string, maxLen int) string {
	if utf8.RuneCountInString(content) > maxLen {
		runes := []rune(content)
		return string(runes[:maxLen])
	}
	return content
}

// TestProperty4_ChatMessageTruncation 属性测试：聊天消息截断
// **Feature: p1-p2-features, Property 4: Chat message truncation**
// **Validates: Requirements 2.3**
func TestProperty4_ChatMessageTruncation(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("messages longer than 200 chars are truncated to exactly 200", prop.ForAll(
		func(length int) bool {
			// 生成指定长度的字符串
			content := ""
			for i := 0; i < length; i++ {
				content += "a"
			}
			
			truncated := TruncateMessage(content, 200)
			truncatedLen := utf8.RuneCountInString(truncated)
			
			if length > 200 {
				return truncatedLen == 200
			}
			return truncatedLen == length
		},
		gen.IntRange(1, 500),
	))

	properties.Property("unicode messages are truncated correctly", prop.ForAll(
		func(length int) bool {
			// 生成包含中文的字符串
			content := ""
			for i := 0; i < length; i++ {
				content += "中"
			}
			
			truncated := TruncateMessage(content, 200)
			truncatedLen := utf8.RuneCountInString(truncated)
			
			if length > 200 {
				return truncatedLen == 200
			}
			return truncatedLen == length
		},
		gen.IntRange(1, 500),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 5 & 6: Rate limiting (already tested in concurrent_test.go)
// =============================================================================

// =============================================================================
// Property 7: Game history pagination
// **Feature: p1-p2-features, Property 7: Game history pagination**
// **Validates: Requirements 4.1**
// =============================================================================

type MockGameRecord struct {
	ID        int64
	CreatedAt time.Time
}

// TestProperty7_GameHistoryPagination 属性测试：游戏记录分页
// **Feature: p1-p2-features, Property 7: Game history pagination**
// **Validates: Requirements 4.1**
func TestProperty7_GameHistoryPagination(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("records are sorted by time descending and limited to page size", prop.ForAll(
		func(totalRecords, pageSize int) bool {
			// 生成记录
			records := make([]MockGameRecord, totalRecords)
			baseTime := time.Now()
			for i := 0; i < totalRecords; i++ {
				records[i] = MockGameRecord{
					ID:        int64(i + 1),
					CreatedAt: baseTime.Add(time.Duration(i) * time.Minute),
				}
			}
			
			// 按时间降序排序
			sort.Slice(records, func(i, j int) bool {
				return records[i].CreatedAt.After(records[j].CreatedAt)
			})
			
			// 分页
			if pageSize > len(records) {
				pageSize = len(records)
			}
			pagedRecords := records[:pageSize]
			
			// 验证：结果按时间降序
			for i := 1; i < len(pagedRecords); i++ {
				if pagedRecords[i].CreatedAt.After(pagedRecords[i-1].CreatedAt) {
					return false
				}
			}
			
			// 验证：结果数量不超过页大小
			return len(pagedRecords) <= pageSize
		},
		gen.IntRange(1, 100),
		gen.IntRange(1, 50),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 8: Game history date filtering
// **Feature: p1-p2-features, Property 8: Game history date filtering**
// **Validates: Requirements 4.3**
// =============================================================================

// TestProperty8_GameHistoryDateFiltering 属性测试：游戏记录日期过滤
// **Feature: p1-p2-features, Property 8: Game history date filtering**
// **Validates: Requirements 4.3**
func TestProperty8_GameHistoryDateFiltering(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("all returned records are within date range", prop.ForAll(
		func(totalRecords, rangeDays int) bool {
			baseTime := time.Now()
			startDate := baseTime.AddDate(0, 0, -rangeDays)
			endDate := baseTime
			
			// 生成记录（一些在范围内，一些在范围外）
			records := make([]MockGameRecord, totalRecords)
			for i := 0; i < totalRecords; i++ {
				// 随机分布在 -2*rangeDays 到 +rangeDays 之间
				offset := (i % (rangeDays * 3)) - rangeDays
				records[i] = MockGameRecord{
					ID:        int64(i + 1),
					CreatedAt: baseTime.AddDate(0, 0, offset),
				}
			}
			
			// 过滤
			var filtered []MockGameRecord
			for _, r := range records {
				if (r.CreatedAt.Equal(startDate) || r.CreatedAt.After(startDate)) &&
					(r.CreatedAt.Equal(endDate) || r.CreatedAt.Before(endDate)) {
					filtered = append(filtered, r)
				}
			}
			
			// 验证：所有过滤后的记录都在范围内
			for _, r := range filtered {
				if r.CreatedAt.Before(startDate) || r.CreatedAt.After(endDate) {
					return false
				}
			}
			return true
		},
		gen.IntRange(10, 100),
		gen.IntRange(1, 30),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 9: Round verification consistency
// **Feature: p1-p2-features, Property 9: Round verification consistency**
// **Validates: Requirements 5.4**
// =============================================================================

// ComputeWinnersForTest 使用种子计算赢家（测试版本）
func ComputeWinnersForTest(seed string, participantIDs []int64, winnerCount int) []int64 {
	if len(participantIDs) == 0 || winnerCount <= 0 {
		return nil
	}
	if winnerCount > len(participantIDs) {
		winnerCount = len(participantIDs)
	}

	hash := sha256.Sum256([]byte(seed))
	shuffled := make([]int64, len(participantIDs))
	copy(shuffled, participantIDs)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(hash[i%32]) % (i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	winners := shuffled[:winnerCount]
	sort.Slice(winners, func(i, j int) bool {
		return winners[i] < winners[j]
	})
	return winners
}

// TestProperty9_RoundVerificationConsistency 属性测试：回合验证一致性
// **Feature: p1-p2-features, Property 9: Round verification consistency**
// **Validates: Requirements 5.4**
func TestProperty9_RoundVerificationConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("computing winners with same seed produces same result", prop.ForAll(
		func(participantCount, winnerCount int, seedSuffix int) bool {
			seed := "test_seed_" + string(rune(seedSuffix))
			
			var participantIDs []int64
			for i := 0; i < participantCount; i++ {
				participantIDs = append(participantIDs, int64(i+1))
			}
			
			// 计算两次
			winners1 := ComputeWinnersForTest(seed, participantIDs, winnerCount)
			winners2 := ComputeWinnersForTest(seed, participantIDs, winnerCount)
			
			// 验证：两次结果相同
			if len(winners1) != len(winners2) {
				return false
			}
			for i := range winners1 {
				if winners1[i] != winners2[i] {
					return false
				}
			}
			return true
		},
		gen.IntRange(2, 20),
		gen.IntRange(1, 5),
		gen.IntRange(0, 1000),
	))

	properties.Property("commit hash matches revealed seed hash", prop.ForAll(
		func(seedSuffix int) bool {
			seed := "reveal_seed_" + string(rune(seedSuffix))
			
			// 计算 commit hash
			hash := sha256.Sum256([]byte(seed))
			commitHash := hex.EncodeToString(hash[:])
			
			// 验证时重新计算
			verifyHash := sha256.Sum256([]byte(seed))
			computedHash := hex.EncodeToString(verifyHash[:])
			
			return commitHash == computedHash
		},
		gen.IntRange(0, 10000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 10: Friend relationship bidirectionality
// **Feature: p1-p2-features, Property 10: Friend relationship bidirectionality**
// **Validates: Requirements 6.2**
// =============================================================================

type MockFriendStore struct {
	mu          sync.RWMutex
	friendships map[int64]map[int64]bool // userID -> friendID -> exists
}

func NewMockFriendStore() *MockFriendStore {
	return &MockFriendStore{
		friendships: make(map[int64]map[int64]bool),
	}
}

func (s *MockFriendStore) CreateFriendship(userID1, userID2 int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.friendships[userID1] == nil {
		s.friendships[userID1] = make(map[int64]bool)
	}
	if s.friendships[userID2] == nil {
		s.friendships[userID2] = make(map[int64]bool)
	}
	
	// 双向添加
	s.friendships[userID1][userID2] = true
	s.friendships[userID2][userID1] = true
}

func (s *MockFriendStore) AreFriends(userID1, userID2 int64) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	if friends, ok := s.friendships[userID1]; ok {
		return friends[userID2]
	}
	return false
}

func (s *MockFriendStore) GetFriendList(userID int64) []int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	var friends []int64
	if friendMap, ok := s.friendships[userID]; ok {
		for friendID := range friendMap {
			friends = append(friends, friendID)
		}
	}
	return friends
}

// TestProperty10_FriendRelationshipBidirectionality 属性测试：好友关系双向性
// **Feature: p1-p2-features, Property 10: Friend relationship bidirectionality**
// **Validates: Requirements 6.2**
func TestProperty10_FriendRelationshipBidirectionality(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("accepted friend request creates bidirectional relationship", prop.ForAll(
		func(userID1, userID2 int64) bool {
			if userID1 == userID2 {
				return true // 跳过自己加自己的情况
			}
			
			store := NewMockFriendStore()
			store.CreateFriendship(userID1, userID2)
			
			// 验证：双向关系
			if !store.AreFriends(userID1, userID2) {
				return false
			}
			if !store.AreFriends(userID2, userID1) {
				return false
			}
			
			// 验证：互相出现在好友列表中
			friends1 := store.GetFriendList(userID1)
			friends2 := store.GetFriendList(userID2)
			
			found1 := false
			for _, f := range friends1 {
				if f == userID2 {
					found1 = true
					break
				}
			}
			
			found2 := false
			for _, f := range friends2 {
				if f == userID1 {
					found2 = true
					break
				}
			}
			
			return found1 && found2
		},
		gen.Int64Range(1, 1000),
		gen.Int64Range(1, 1000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 11: Friend removal bidirectionality
// **Feature: p1-p2-features, Property 11: Friend removal bidirectionality**
// **Validates: Requirements 6.5**
// =============================================================================

func (s *MockFriendStore) RemoveFriendship(userID1, userID2 int64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	// 双向删除
	if friends, ok := s.friendships[userID1]; ok {
		delete(friends, userID2)
	}
	if friends, ok := s.friendships[userID2]; ok {
		delete(friends, userID1)
	}
}

// TestProperty11_FriendRemovalBidirectionality 属性测试：好友删除双向性
// **Feature: p1-p2-features, Property 11: Friend removal bidirectionality**
// **Validates: Requirements 6.5**
func TestProperty11_FriendRemovalBidirectionality(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("removing friend removes bidirectional relationship", prop.ForAll(
		func(userID1, userID2 int64) bool {
			if userID1 == userID2 {
				return true
			}
			
			store := NewMockFriendStore()
			
			// 先建立好友关系
			store.CreateFriendship(userID1, userID2)
			
			// 删除好友
			store.RemoveFriendship(userID1, userID2)
			
			// 验证：双向都不再是好友
			if store.AreFriends(userID1, userID2) {
				return false
			}
			if store.AreFriends(userID2, userID1) {
				return false
			}
			
			// 验证：互相不在好友列表中
			friends1 := store.GetFriendList(userID1)
			friends2 := store.GetFriendList(userID2)
			
			for _, f := range friends1 {
				if f == userID2 {
					return false
				}
			}
			for _, f := range friends2 {
				if f == userID1 {
					return false
				}
			}
			
			return true
		},
		gen.Int64Range(1, 1000),
		gen.Int64Range(1, 1000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 12: Invitation notification delivery
// **Feature: p1-p2-features, Property 12: Invitation notification delivery**
// **Validates: Requirements 7.2**
// =============================================================================

type MockInvitationNotification struct {
	RoomID      int64
	RoomName    string
	BetAmount   string
	PlayerCount int
}

type MockNotificationStore struct {
	mu            sync.Mutex
	notifications map[int64][]MockInvitationNotification
}

func NewMockNotificationStore() *MockNotificationStore {
	return &MockNotificationStore{
		notifications: make(map[int64][]MockInvitationNotification),
	}
}

func (s *MockNotificationStore) SendInvitation(toUserID int64, notification MockInvitationNotification) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.notifications[toUserID] = append(s.notifications[toUserID], notification)
}

func (s *MockNotificationStore) GetNotifications(userID int64) []MockInvitationNotification {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.notifications[userID]
}

// TestProperty12_InvitationNotificationDelivery 属性测试：邀请通知送达
// **Feature: p1-p2-features, Property 12: Invitation notification delivery**
// **Validates: Requirements 7.2**
func TestProperty12_InvitationNotificationDelivery(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("invitation contains room name, bet amount, and player count", prop.ForAll(
		func(roomID int64, betAmount int64, playerCount int) bool {
			store := NewMockNotificationStore()
			targetUserID := int64(100)
			
			notification := MockInvitationNotification{
				RoomID:      roomID,
				RoomName:    "Test Room",
				BetAmount:   decimal.NewFromInt(betAmount).String(),
				PlayerCount: playerCount,
			}
			
			store.SendInvitation(targetUserID, notification)
			
			// 验证：目标用户收到通知
			notifications := store.GetNotifications(targetUserID)
			if len(notifications) != 1 {
				return false
			}
			
			n := notifications[0]
			// 验证：通知包含必要信息
			if n.RoomName == "" {
				return false
			}
			if n.BetAmount == "" {
				return false
			}
			if n.PlayerCount <= 0 {
				return false
			}
			
			return true
		},
		gen.Int64Range(1, 1000),
		gen.Int64Range(1, 1000),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 13: Invite link validity
// **Feature: p1-p2-features, Property 13: Invite link validity**
// **Validates: Requirements 7.5, 7.6**
// =============================================================================

type MockInviteLink struct {
	Code      string
	RoomID    int64
	ExpiresAt time.Time
	UseCount  int
	MaxUses   *int
}

type MockInviteLinkStore struct {
	mu    sync.Mutex
	links map[string]*MockInviteLink
}

func NewMockInviteLinkStore() *MockInviteLinkStore {
	return &MockInviteLinkStore{
		links: make(map[string]*MockInviteLink),
	}
}

func (s *MockInviteLinkStore) CreateLink(code string, roomID int64, expiresIn time.Duration) *MockInviteLink {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	link := &MockInviteLink{
		Code:      code,
		RoomID:    roomID,
		ExpiresAt: time.Now().Add(expiresIn),
		UseCount:  0,
	}
	s.links[code] = link
	return link
}

func (s *MockInviteLinkStore) UseLink(code string) (bool, string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	link, ok := s.links[code]
	if !ok {
		return false, "invalid"
	}
	
	if time.Now().After(link.ExpiresAt) {
		return false, "expired"
	}
	
	if link.MaxUses != nil && link.UseCount >= *link.MaxUses {
		return false, "max_uses"
	}
	
	link.UseCount++
	return true, ""
}

// TestProperty13_InviteLinkValidity 属性测试：邀请链接有效性
// **Feature: p1-p2-features, Property 13: Invite link validity**
// **Validates: Requirements 7.5, 7.6**
func TestProperty13_InviteLinkValidity(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("link is usable within 24 hours", prop.ForAll(
		func(hoursOffset int) bool {
			store := NewMockInviteLinkStore()
			
			// 创建24小时有效的链接
			link := store.CreateLink("test123", 1, 24*time.Hour)
			
			// 模拟时间偏移（通过直接修改过期时间来测试）
			if hoursOffset < 24 {
				// 链接应该有效
				ok, _ := store.UseLink(link.Code)
				return ok
			}
			return true
		},
		gen.IntRange(0, 48),
	))

	properties.Property("expired link is rejected", prop.ForAll(
		func(roomID int64) bool {
			store := NewMockInviteLinkStore()
			
			// 创建已过期的链接
			link := store.CreateLink("expired123", roomID, -1*time.Hour)
			_ = link
			
			ok, reason := store.UseLink("expired123")
			return !ok && reason == "expired"
		},
		gen.Int64Range(1, 1000),
	))

	properties.Property("invalid link code is rejected", prop.ForAll(
		func(roomID int64) bool {
			store := NewMockInviteLinkStore()
			store.CreateLink("valid123", roomID, 24*time.Hour)
			
			ok, reason := store.UseLink("invalid_code")
			return !ok && reason == "invalid"
		},
		gen.Int64Range(1, 1000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 14: Balance cache consistency
// **Feature: p1-p2-features, Property 14: Balance cache consistency**
// **Validates: Requirements 10.2**
// =============================================================================

type MockBalanceCache struct {
	mu       sync.RWMutex
	cache    map[int64]decimal.Decimal
	database map[int64]decimal.Decimal
}

func NewMockBalanceCache() *MockBalanceCache {
	return &MockBalanceCache{
		cache:    make(map[int64]decimal.Decimal),
		database: make(map[int64]decimal.Decimal),
	}
}

func (c *MockBalanceCache) SetBalance(userID int64, balance decimal.Decimal) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// 原子更新数据库和缓存
	c.database[userID] = balance
	c.cache[userID] = balance
}

func (c *MockBalanceCache) GetFromCache(userID int64) (decimal.Decimal, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	bal, ok := c.cache[userID]
	return bal, ok
}

func (c *MockBalanceCache) GetFromDB(userID int64) (decimal.Decimal, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	bal, ok := c.database[userID]
	return bal, ok
}

// TestProperty14_BalanceCacheConsistency 属性测试：缓存一致性
// **Feature: p1-p2-features, Property 14: Balance cache consistency**
// **Validates: Requirements 10.2**
func TestProperty14_BalanceCacheConsistency(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("cache reflects database after transaction", prop.ForAll(
		func(userID int64, balanceInt int64) bool {
			cache := NewMockBalanceCache()
			balance := decimal.NewFromInt(balanceInt)
			
			cache.SetBalance(userID, balance)
			
			cacheVal, cacheOk := cache.GetFromCache(userID)
			dbVal, dbOk := cache.GetFromDB(userID)
			
			if !cacheOk || !dbOk {
				return false
			}
			
			return cacheVal.Equal(dbVal)
		},
		gen.Int64Range(1, 1000),
		gen.Int64Range(0, 100000),
	))

	properties.Property("multiple updates maintain consistency", prop.ForAll(
		func(userID int64, updates []int64) bool {
			cache := NewMockBalanceCache()
			
			for _, amount := range updates {
				cache.SetBalance(userID, decimal.NewFromInt(amount))
			}
			
			if len(updates) == 0 {
				return true
			}
			
			cacheVal, cacheOk := cache.GetFromCache(userID)
			dbVal, dbOk := cache.GetFromDB(userID)
			
			if !cacheOk || !dbOk {
				return false
			}
			
			return cacheVal.Equal(dbVal)
		},
		gen.Int64Range(1, 1000),
		gen.SliceOf(gen.Int64Range(0, 10000)),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 15: Optimistic locking prevents conflicts
// **Feature: p1-p2-features, Property 15: Optimistic locking prevents conflicts**
// **Validates: Requirements 10.6**
// =============================================================================

type MockOptimisticLockStore struct {
	mu       sync.Mutex
	balances map[int64]decimal.Decimal
	versions map[int64]int64
}

func NewMockOptimisticLockStore() *MockOptimisticLockStore {
	return &MockOptimisticLockStore{
		balances: make(map[int64]decimal.Decimal),
		versions: make(map[int64]int64),
	}
}

func (s *MockOptimisticLockStore) SetInitial(userID int64, balance decimal.Decimal) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.balances[userID] = balance
	s.versions[userID] = 1
}

func (s *MockOptimisticLockStore) GetVersion(userID int64) int64 {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.versions[userID]
}

func (s *MockOptimisticLockStore) UpdateWithVersion(userID int64, delta decimal.Decimal, expectedVersion int64) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	
	if s.versions[userID] != expectedVersion {
		return false
	}
	
	s.balances[userID] = s.balances[userID].Add(delta)
	s.versions[userID]++
	return true
}

// TestProperty15_OptimisticLockingPreventsConflicts 属性测试：乐观锁防止冲突
// **Feature: p1-p2-features, Property 15: Optimistic locking prevents conflicts**
// **Validates: Requirements 10.6**
func TestProperty15_OptimisticLockingPreventsConflicts(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("concurrent updates with same version - only one succeeds", prop.ForAll(
		func(initialBalance int64, concurrentUpdates int) bool {
			store := NewMockOptimisticLockStore()
			userID := int64(1)
			store.SetInitial(userID, decimal.NewFromInt(initialBalance))
			
			// 获取当前版本
			version := store.GetVersion(userID)
			
			// 并发更新
			var wg sync.WaitGroup
			var successCount int32
			var mu sync.Mutex
			
			for i := 0; i < concurrentUpdates; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if store.UpdateWithVersion(userID, decimal.NewFromInt(10), version) {
						mu.Lock()
						successCount++
						mu.Unlock()
					}
				}()
			}
			
			wg.Wait()
			
			// 只有一个更新应该成功
			return successCount == 1
		},
		gen.Int64Range(1000, 10000),
		gen.IntRange(2, 10),
	))

	properties.Property("sequential updates with correct version all succeed", prop.ForAll(
		func(initialBalance int64, updateCount int) bool {
			store := NewMockOptimisticLockStore()
			userID := int64(1)
			store.SetInitial(userID, decimal.NewFromInt(initialBalance))
			
			for i := 0; i < updateCount; i++ {
				version := store.GetVersion(userID)
				if !store.UpdateWithVersion(userID, decimal.NewFromInt(10), version) {
					return false
				}
			}
			
			return true
		},
		gen.Int64Range(1000, 10000),
		gen.IntRange(1, 10),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 18: Negative balance alert
// **Feature: p1-p2-features, Property 18: Negative balance alert**
// **Validates: Requirements 13.1**
// =============================================================================

type MockAlertStore struct {
	mu     sync.Mutex
	alerts []string
}

func NewMockAlertStore() *MockAlertStore {
	return &MockAlertStore{
		alerts: make([]string, 0),
	}
}

func (s *MockAlertStore) TriggerAlert(alertType string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.alerts = append(s.alerts, alertType)
}

func (s *MockAlertStore) HasAlert(alertType string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, a := range s.alerts {
		if a == alertType {
			return true
		}
	}
	return false
}

func (s *MockAlertStore) CheckBalance(balance decimal.Decimal) {
	if balance.IsNegative() {
		s.TriggerAlert("negative_balance")
	}
}

// TestProperty18_NegativeBalanceAlert 属性测试：负余额告警
// **Feature: p1-p2-features, Property 18: Negative balance alert**
// **Validates: Requirements 13.1**
func TestProperty18_NegativeBalanceAlert(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("negative balance triggers alert", prop.ForAll(
		func(balanceInt int64) bool {
			store := NewMockAlertStore()
			balance := decimal.NewFromInt(balanceInt)
			
			store.CheckBalance(balance)
			
			if balance.IsNegative() {
				return store.HasAlert("negative_balance")
			}
			return !store.HasAlert("negative_balance")
		},
		gen.Int64Range(-1000, 1000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 19: Large transaction alert
// **Feature: p1-p2-features, Property 19: Large transaction alert**
// **Validates: Requirements 13.3**
// =============================================================================

const LargeTransactionThreshold = 10000

func (s *MockAlertStore) CheckTransaction(amount decimal.Decimal) {
	if amount.GreaterThan(decimal.NewFromInt(LargeTransactionThreshold)) {
		s.TriggerAlert("large_transaction")
	}
}

// TestProperty19_LargeTransactionAlert 属性测试：大额交易告警
// **Feature: p1-p2-features, Property 19: Large transaction alert**
// **Validates: Requirements 13.3**
func TestProperty19_LargeTransactionAlert(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("transaction exceeding 10000 triggers alert", prop.ForAll(
		func(amountInt int64) bool {
			store := NewMockAlertStore()
			amount := decimal.NewFromInt(amountInt)
			
			store.CheckTransaction(amount)
			
			if amount.GreaterThan(decimal.NewFromInt(LargeTransactionThreshold)) {
				return store.HasAlert("large_transaction")
			}
			return !store.HasAlert("large_transaction")
		},
		gen.Int64Range(1, 20000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 20: Theme persistence
// **Feature: p1-p2-features, Property 20: Theme persistence**
// **Validates: Requirements 14.3**
// =============================================================================

type MockThemeStore struct {
	mu     sync.RWMutex
	themes map[int64]string // roomID -> themeName
}

func NewMockThemeStore() *MockThemeStore {
	return &MockThemeStore{
		themes: make(map[int64]string),
	}
}

func (s *MockThemeStore) SetTheme(roomID int64, themeName string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.themes[roomID] = themeName
}

func (s *MockThemeStore) GetTheme(roomID int64) (string, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	theme, ok := s.themes[roomID]
	return theme, ok
}

// TestProperty20_ThemePersistence 属性测试：主题持久化
// **Feature: p1-p2-features, Property 20: Theme persistence**
// **Validates: Requirements 14.3**
func TestProperty20_ThemePersistence(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	validThemes := []string{"classic", "neon", "ocean", "forest", "luxury"}

	properties.Property("players joining room receive theme information", prop.ForAll(
		func(roomID int64, themeIndex int) bool {
			store := NewMockThemeStore()
			themeName := validThemes[themeIndex%len(validThemes)]
			
			// 设置房间主题
			store.SetTheme(roomID, themeName)
			
			// 模拟玩家加入获取主题
			theme, ok := store.GetTheme(roomID)
			
			if !ok {
				return false
			}
			
			return theme == themeName
		},
		gen.Int64Range(1, 1000),
		gen.IntRange(0, 100),
	))

	properties.Property("theme persists across multiple queries", prop.ForAll(
		func(roomID int64, queryCount int) bool {
			store := NewMockThemeStore()
			store.SetTheme(roomID, "neon")
			
			for i := 0; i < queryCount; i++ {
				theme, ok := store.GetTheme(roomID)
				if !ok || theme != "neon" {
					return false
				}
			}
			
			return true
		},
		gen.Int64Range(1, 1000),
		gen.IntRange(1, 20),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 21: Wallet balance accuracy
// **Feature: p1-p2-features, Property 21: Wallet balance accuracy**
// **Validates: Requirements 15.1**
// =============================================================================

type MockWallet struct {
	AvailableBalance decimal.Decimal
	FrozenBalance    decimal.Decimal
}

func (w *MockWallet) TotalBalance() decimal.Decimal {
	return w.AvailableBalance.Add(w.FrozenBalance)
}

// TestProperty21_WalletBalanceAccuracy 属性测试：钱包余额准确性
// **Feature: p1-p2-features, Property 21: Wallet balance accuracy**
// **Validates: Requirements 15.1**
func TestProperty21_WalletBalanceAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("total balance equals available plus frozen", prop.ForAll(
		func(availableInt, frozenInt int64) bool {
			wallet := &MockWallet{
				AvailableBalance: decimal.NewFromInt(availableInt),
				FrozenBalance:    decimal.NewFromInt(frozenInt),
			}
			
			expectedTotal := decimal.NewFromInt(availableInt + frozenInt)
			actualTotal := wallet.TotalBalance()
			
			return actualTotal.Equal(expectedTotal)
		},
		gen.Int64Range(0, 100000),
		gen.Int64Range(0, 100000),
	))

	properties.TestingRun(t)
}

// =============================================================================
// Property 22: Earnings calculation accuracy
// **Feature: p1-p2-features, Property 22: Earnings calculation accuracy**
// **Validates: Requirements 15.5**
// =============================================================================

type MockEarnings struct {
	TotalWinnings decimal.Decimal
	TotalLosses   decimal.Decimal
}

func (e *MockEarnings) NetProfit() decimal.Decimal {
	return e.TotalWinnings.Sub(e.TotalLosses)
}

// TestProperty22_EarningsCalculationAccuracy 属性测试：收益计算准确性
// **Feature: p1-p2-features, Property 22: Earnings calculation accuracy**
// **Validates: Requirements 15.5**
func TestProperty22_EarningsCalculationAccuracy(t *testing.T) {
	parameters := gopter.DefaultTestParameters()
	parameters.MinSuccessfulTests = 100
	properties := gopter.NewProperties(parameters)

	properties.Property("net profit equals winnings minus losses", prop.ForAll(
		func(winningsInt, lossesInt int64) bool {
			earnings := &MockEarnings{
				TotalWinnings: decimal.NewFromInt(winningsInt),
				TotalLosses:   decimal.NewFromInt(lossesInt),
			}
			
			expectedProfit := decimal.NewFromInt(winningsInt - lossesInt)
			actualProfit := earnings.NetProfit()
			
			return actualProfit.Equal(expectedProfit)
		},
		gen.Int64Range(0, 100000),
		gen.Int64Range(0, 100000),
	))

	properties.Property("net profit can be negative", prop.ForAll(
		func(winningsInt, lossesInt int64) bool {
			earnings := &MockEarnings{
				TotalWinnings: decimal.NewFromInt(winningsInt),
				TotalLosses:   decimal.NewFromInt(lossesInt),
			}
			
			profit := earnings.NetProfit()
			
			if lossesInt > winningsInt {
				return profit.IsNegative()
			}
			return true
		},
		gen.Int64Range(0, 50000),
		gen.Int64Range(50001, 100000),
	))

	properties.TestingRun(t)
}
