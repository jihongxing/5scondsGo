// Package integration_test 集成测试
package integration_test

import (
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/ws"
)

// MockConn 模拟 WebSocket 连接
type MockConn struct {
	mu       sync.Mutex
	messages []interface{}
	closed   bool
}

func NewMockConn() *MockConn {
	return &MockConn{
		messages: make([]interface{}, 0),
	}
}

func (c *MockConn) WriteJSON(v interface{}) error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return nil
	}
	c.messages = append(c.messages, v)
	return nil
}

func (c *MockConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *MockConn) GetMessages() []interface{} {
	c.mu.Lock()
	defer c.mu.Unlock()
	result := make([]interface{}, len(c.messages))
	copy(result, c.messages)
	return result
}

func (c *MockConn) ClearMessages() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.messages = make([]interface{}, 0)
}

// TestWSBroadcastToAllRoomMembers 测试消息广播到房间所有成员
func TestWSBroadcastToAllRoomMembers(t *testing.T) {
	hub := ws.NewHub()

	// 创建多个用户连接
	roomID := int64(1)
	userCount := 5
	conns := make([]*MockConn, userCount)

	for i := 0; i < userCount; i++ {
		conns[i] = NewMockConn()
		hub.AddConn(roomID, int64(i+1), conns[i])
	}

	// 广播消息
	testMsg := &model.WSMessage{
		Type: model.WSTypeChatMessage,
		Payload: &model.WSChatMessage{
			ID:        1,
			UserID:    1,
			Username:  "testuser",
			Content:   "Hello, World!",
			Timestamp: time.Now().UnixMilli(),
		},
	}

	hub.BroadcastToRoom(roomID, testMsg)

	// 验证所有用户都收到消息
	for i, conn := range conns {
		messages := conn.GetMessages()
		if len(messages) != 1 {
			t.Errorf("User %d: expected 1 message, got %d", i+1, len(messages))
			continue
		}

		// 验证消息类型
		msg, ok := messages[0].(*model.WSMessage)
		if !ok {
			t.Errorf("User %d: message type mismatch", i+1)
			continue
		}

		if msg.Type != model.WSTypeChatMessage {
			t.Errorf("User %d: expected type %s, got %s", i+1, model.WSTypeChatMessage, msg.Type)
		}
	}
}


// TestWSBroadcastWithSpectators 测试消息广播包含观战者
func TestWSBroadcastWithSpectators(t *testing.T) {
	hub := ws.NewHub()

	roomID := int64(1)

	// 添加参与者
	participantConns := make([]*MockConn, 3)
	for i := 0; i < 3; i++ {
		participantConns[i] = NewMockConn()
		hub.AddConn(roomID, int64(i+1), participantConns[i])
	}

	// 添加观战者
	spectatorConns := make([]*MockConn, 2)
	for i := 0; i < 2; i++ {
		spectatorConns[i] = NewMockConn()
		hub.AddConn(roomID, int64(100+i), spectatorConns[i])
	}

	// 广播游戏状态更新
	testMsg := &model.WSMessage{
		Type: model.WSTypePhaseChange,
		Payload: &model.WSPhaseChange{
			Phase:        model.PhaseBetting,
			PhaseEndTime: time.Now().Add(5 * time.Second).UnixMilli(),
			Round:        1,
		},
	}

	hub.BroadcastToRoom(roomID, testMsg)

	// 验证参与者收到消息
	for i, conn := range participantConns {
		messages := conn.GetMessages()
		if len(messages) != 1 {
			t.Errorf("Participant %d: expected 1 message, got %d", i+1, len(messages))
		}
	}

	// 验证观战者也收到消息
	for i, conn := range spectatorConns {
		messages := conn.GetMessages()
		if len(messages) != 1 {
			t.Errorf("Spectator %d: expected 1 message, got %d", i+1, len(messages))
		}
	}
}

// TestWSPrivateMessage 测试私发消息
func TestWSPrivateMessage(t *testing.T) {
	hub := ws.NewHub()

	roomID := int64(1)
	targetUserID := int64(2)

	// 添加多个用户
	conn1 := NewMockConn()
	conn2 := NewMockConn()
	conn3 := NewMockConn()

	hub.AddConn(roomID, 1, conn1)
	hub.AddConn(roomID, targetUserID, conn2)
	hub.AddConn(roomID, 3, conn3)

	// 私发消息给用户2
	privateMsg := &model.WSMessage{
		Type: model.WSTypeRoomInvitation,
		Payload: &model.WSRoomInvitation{
			InvitationID: 1,
			RoomID:       2,
			RoomName:     "Test Room",
			BetAmount:    "10.00",
			PlayerCount:  3,
			FromUserID:   1,
			FromUsername: "user1",
		},
	}

	hub.SendToUser(targetUserID, privateMsg)

	// 验证只有目标用户收到消息
	if len(conn1.GetMessages()) != 0 {
		t.Error("User 1 should not receive private message")
	}
	if len(conn2.GetMessages()) != 1 {
		t.Error("User 2 should receive private message")
	}
	if len(conn3.GetMessages()) != 0 {
		t.Error("User 3 should not receive private message")
	}
}

// TestWSMultipleRoomsBroadcast 测试多房间广播隔离
func TestWSMultipleRoomsBroadcast(t *testing.T) {
	hub := ws.NewHub()

	// 房间1的用户
	room1Conns := make([]*MockConn, 3)
	for i := 0; i < 3; i++ {
		room1Conns[i] = NewMockConn()
		hub.AddConn(1, int64(i+1), room1Conns[i])
	}

	// 房间2的用户
	room2Conns := make([]*MockConn, 2)
	for i := 0; i < 2; i++ {
		room2Conns[i] = NewMockConn()
		hub.AddConn(2, int64(100+i), room2Conns[i])
	}

	// 向房间1广播
	msg := &model.WSMessage{
		Type:    model.WSTypeChatMessage,
		Payload: map[string]string{"content": "Room 1 message"},
	}
	hub.BroadcastToRoom(1, msg)

	// 验证房间1用户收到消息
	for i, conn := range room1Conns {
		if len(conn.GetMessages()) != 1 {
			t.Errorf("Room 1 User %d: expected 1 message, got %d", i+1, len(conn.GetMessages()))
		}
	}

	// 验证房间2用户没有收到消息
	for i, conn := range room2Conns {
		if len(conn.GetMessages()) != 0 {
			t.Errorf("Room 2 User %d: should not receive Room 1 message", i+1)
		}
	}
}

// TestWSConcurrentBroadcast 测试并发广播
func TestWSConcurrentBroadcast(t *testing.T) {
	hub := ws.NewHub()

	roomID := int64(1)
	userCount := 10
	messageCount := 100

	conns := make([]*MockConn, userCount)
	for i := 0; i < userCount; i++ {
		conns[i] = NewMockConn()
		hub.AddConn(roomID, int64(i+1), conns[i])
	}

	// 并发广播消息
	var wg sync.WaitGroup
	for i := 0; i < messageCount; i++ {
		wg.Add(1)
		go func(msgID int) {
			defer wg.Done()
			msg := &model.WSMessage{
				Type: model.WSTypePhaseTick,
				Payload: &model.WSPhaseTick{
					ServerTime: time.Now().UnixMilli(),
				},
			}
			hub.BroadcastToRoom(roomID, msg)
		}(i)
	}

	wg.Wait()

	// 验证所有用户都收到了所有消息
	for i, conn := range conns {
		messages := conn.GetMessages()
		if len(messages) != messageCount {
			t.Errorf("User %d: expected %d messages, got %d", i+1, messageCount, len(messages))
		}
	}
}

// TestWSUserLeaveAndRejoin 测试用户离开和重新加入
func TestWSUserLeaveAndRejoin(t *testing.T) {
	hub := ws.NewHub()

	roomID := int64(1)
	userID := int64(1)

	conn1 := NewMockConn()
	hub.AddConn(roomID, userID, conn1)

	// 广播第一条消息
	msg1 := &model.WSMessage{Type: model.WSTypeChatMessage, Payload: "msg1"}
	hub.BroadcastToRoom(roomID, msg1)

	if len(conn1.GetMessages()) != 1 {
		t.Error("User should receive first message")
	}

	// 用户离开
	hub.RemoveConn(roomID, userID)

	// 广播第二条消息
	msg2 := &model.WSMessage{Type: model.WSTypeChatMessage, Payload: "msg2"}
	hub.BroadcastToRoom(roomID, msg2)

	// 用户不应该收到第二条消息
	if len(conn1.GetMessages()) != 1 {
		t.Error("User should not receive message after leaving")
	}

	// 用户重新加入
	conn2 := NewMockConn()
	hub.AddConn(roomID, userID, conn2)

	// 广播第三条消息
	msg3 := &model.WSMessage{Type: model.WSTypeChatMessage, Payload: "msg3"}
	hub.BroadcastToRoom(roomID, msg3)

	// 新连接应该收到第三条消息
	if len(conn2.GetMessages()) != 1 {
		t.Error("User should receive message after rejoining")
	}
}

// TestWSMessageSerialization 测试消息序列化
func TestWSMessageSerialization(t *testing.T) {
	// 测试聊天消息序列化
	chatMsg := &model.WSMessage{
		Type: model.WSTypeChatMessage,
		Payload: &model.WSChatMessage{
			ID:        123,
			UserID:    456,
			Username:  "testuser",
			Content:   "Hello, 世界!",
			Timestamp: 1699999999000,
		},
	}

	data, err := json.Marshal(chatMsg)
	if err != nil {
		t.Fatalf("Failed to marshal chat message: %v", err)
	}

	var decoded model.WSMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal chat message: %v", err)
	}

	if decoded.Type != model.WSTypeChatMessage {
		t.Errorf("Expected type %s, got %s", model.WSTypeChatMessage, decoded.Type)
	}

	// 测试表情消息序列化
	emojiMsg := &model.WSMessage{
		Type: model.WSTypeEmojiReaction,
		Payload: &model.WSEmojiReaction{
			UserID:   1,
			Username: "user1",
			Emoji:    "😀",
		},
	}

	data, err = json.Marshal(emojiMsg)
	if err != nil {
		t.Fatalf("Failed to marshal emoji message: %v", err)
	}

	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Failed to unmarshal emoji message: %v", err)
	}

	if decoded.Type != model.WSTypeEmojiReaction {
		t.Errorf("Expected type %s, got %s", model.WSTypeEmojiReaction, decoded.Type)
	}
}
