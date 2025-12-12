package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/fiveseconds/server/internal/game"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/service"
	"github.com/fiveseconds/server/internal/ws"
	"github.com/fiveseconds/server/pkg/metrics"
	"github.com/fiveseconds/server/pkg/trace"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

// TokenValidator 用于验证 JWT token
type TokenValidator interface {
	ValidateToken(tokenString string) (*service.Claims, error)
}

// UserGetter 用于获取用户信息
type UserGetter interface {
	GetUserByID(ctx context.Context, userID int64) (*model.User, error)
}

// ChatServiceInterface 聊天服务接口
type ChatServiceInterface interface {
	SendMessage(ctx context.Context, roomID, userID int64, username, content string) (*model.ChatMessage, error)
	GetHistory(ctx context.Context, roomID int64, limit int) ([]*model.ChatMessage, error)
	ValidateEmoji(emoji string) error
	CheckEmojiRateLimit(userID int64) error
}

type WSHandler struct {
	hub            *ws.Hub
	manager        *game.Manager
	tokenValidator TokenValidator
	userGetter     UserGetter
	chatService    ChatServiceInterface
	logger         *zap.Logger
}

func NewWSHandler(hub *ws.Hub, manager *game.Manager, tokenValidator TokenValidator, userGetter UserGetter, chatService ChatServiceInterface, logger *zap.Logger) *WSHandler {
	return &WSHandler{
		hub:            hub,
		manager:        manager,
		tokenValidator: tokenValidator,
		userGetter:     userGetter,
		chatService:    chatService,
		logger:         logger,
	}
}

// HandleWS 处理 WebSocket 连接
func (h *WSHandler) HandleWS(c *gin.Context) {
	// 从 query 获取 token
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
		return
	}

	claims, err := h.tokenValidator.ValidateToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		h.logger.Error("WebSocket upgrade failed", zap.Error(err))
		return
	}

	// Generate session ID for this WebSocket connection
	sessionID := trace.NewSessionID()

	// Create context with trace info
	ctx := trace.WithSessionID(context.Background(), sessionID)
	ctx = trace.WithUserID(ctx, claims.UserID)

	// Record WebSocket connection metric
	metrics.RecordWSConnection(1)

	client := &wsClient{
		conn:        conn,
		userID:      claims.UserID,
		username:    claims.Username,
		sessionID:   sessionID,
		ctx:         ctx,
		hub:         h.hub,
		manager:     h.manager,
		userGetter:  h.userGetter,
		chatService: h.chatService,
		logger:      h.logger.With(zap.Int64("user_id", claims.UserID), zap.String("session_id", sessionID)),
	}

	h.logger.Info("WebSocket connection established",
		zap.Int64("user_id", claims.UserID),
		zap.String("session_id", sessionID),
	)

	go client.readPump()
	go client.writePump()
}

type wsClient struct {
	conn        *websocket.Conn
	userID      int64
	username    string
	sessionID   string
	ctx         context.Context
	roomID      int64
	hub         *ws.Hub
	manager     *game.Manager
	userGetter  UserGetter
	chatService ChatServiceInterface
	writeMu     sync.Mutex // 保护 WebSocket 写操作
	logger      *zap.Logger
}

// safeConn 是一个线程安全的连接包装器，用于 Hub 广播
type safeConn struct {
	conn    *websocket.Conn
	writeMu *sync.Mutex
}

func (s *safeConn) WriteJSON(v interface{}) error {
	s.writeMu.Lock()
	defer s.writeMu.Unlock()
	return s.conn.WriteJSON(v)
}

func (s *safeConn) Close() error {
	return s.conn.Close()
}

// writeJSON 线程安全地写入 JSON 消息
func (c *wsClient) writeJSON(v interface{}) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	return c.conn.WriteJSON(v)
}

func (c *wsClient) readPump() {
	defer func() {
		c.handleDisconnect()
		c.conn.Close()
	}()

	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.logger.Error("WebSocket read error", zap.Error(err))
			} else {
				c.logger.Debug("WebSocket read ended", zap.Error(err))
			}
			break
		}

		// 收到任何消息都重置读取超时
		newDeadline := time.Now().Add(60 * time.Second)
		if err := c.conn.SetReadDeadline(newDeadline); err != nil {
			c.logger.Error("Failed to set read deadline", zap.Error(err))
		}

		var msg model.WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			c.logger.Warn("Invalid message format", zap.Error(err))
			continue
		}

		c.handleMessage(&msg)
	}
}

func (c *wsClient) writePump() {
	ticker := time.NewTicker(30 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
		c.logger.Debug("writePump exited")
	}()

	for range ticker.C {
		// 使用 WriteControl 发送 ping，它有自己的超时控制，不会阻塞太久
		if err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second)); err != nil {
			c.logger.Debug("Failed to send ping", zap.Error(err))
			return
		}
		c.logger.Debug("Ping sent successfully")
	}
}

func (c *wsClient) handleMessage(msg *model.WSMessage) {
	// Record WebSocket message metric
	metrics.RecordWSMessage(string(msg.Type), "inbound")

	switch msg.Type {
	case model.WSTypeHeartbeat:
		// 心跳,记录日志
		c.logger.Debug("Heartbeat received")

	case model.WSTypeJoinRoom:
		c.handleJoinRoom(msg.Payload)

	case model.WSTypeLeaveRoom:
		c.handleLeaveRoom()

	case model.WSTypeSetAutoReady:
		c.handleSetAutoReady(msg.Payload)

	case model.WSTypeJoinAsSpectator:
		c.handleJoinAsSpectator(msg.Payload)

	case model.WSTypeSwitchToParticipant:
		c.handleSwitchToParticipant()

	case model.WSTypeSendChat:
		c.handleSendChat(msg.Payload)

	case model.WSTypeSendEmoji:
		c.handleSendEmoji(msg.Payload)
	}
}

func (c *wsClient) handleJoinRoom(payload interface{}) {
	data, _ := json.Marshal(payload)
	var req model.WSJoinRoom
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError(400, "invalid payload")
		return
	}

	// 如果已在其他房间,先离开
	if c.roomID != 0 && c.roomID != req.RoomID {
		c.handleLeaveRoom()
	}

	processor, err := c.manager.GetOrCreateRoom(context.Background(), req.RoomID)
	if err != nil {
		c.sendError(404, "room not found")
		return
	}

	// 检查是否已是参与者
	isExistingPlayer := processor.IsParticipant(c.userID)

	// 如果不是已有参与者，检查房间状态
	if !isExistingPlayer {
		state := processor.GetRoomState()
		// 游戏进行中（非等待阶段）只能以观战者身份加入
		if state.Phase != model.PhaseWaiting {
			// 获取用户信息
			user, err := c.userGetter.GetUserByID(context.Background(), c.userID)
			if err != nil {
				c.sendError(500, "failed to get user")
				return
			}

			// 自动以观战者身份加入
			if err := processor.AddSpectator(user); err != nil {
				switch err {
				case game.ErrAlreadySpectator:
					// 已是观战者，继续
				case game.ErrSpectatorLimitReached:
					c.sendError(400, "room is busy, spectator limit reached")
					return
				default:
					c.sendError(500, err.Error())
					return
				}
			}

			c.roomID = req.RoomID
			c.hub.AddConn(req.RoomID, c.userID, &safeConn{conn: c.conn, writeMu: &c.writeMu})

			// 发送房间状态（包含观战者标识）
			roomState := processor.GetRoomStateForUser(c.userID)
			c.writeJSON(&model.WSMessage{
				Type:    model.WSTypeRoomState,
				Payload: roomState,
			})

			c.logger.Info("User joined room as spectator (game in progress)", zap.Int64("room_id", req.RoomID))
			return
		}

		// 等待阶段，检查房间是否已满
		if len(state.Players) >= processor.Room.MaxPlayers {
			c.sendError(400, "room is full")
			return
		}

		// 获取用户信息并添加为玩家
		user, err := c.userGetter.GetUserByID(context.Background(), c.userID)
		if err != nil {
			c.sendError(500, "failed to get user")
			return
		}
		processor.AddPlayer(user)
	}

	c.roomID = req.RoomID
	c.hub.AddConn(req.RoomID, c.userID, &safeConn{conn: c.conn, writeMu: &c.writeMu})

	// 标记在线
	processor.SetPlayerOnline(c.userID, true)

	// 发送房间状态
	state := processor.GetRoomStateForUser(c.userID)
	c.writeJSON(&model.WSMessage{
		Type:    model.WSTypeRoomState,
		Payload: state,
	})

	c.logger.Info("User joined room", zap.Int64("room_id", req.RoomID))
}

func (c *wsClient) handleLeaveRoom() {
	if c.roomID == 0 {
		return
	}

	c.hub.RemoveConn(c.roomID, c.userID)

	if processor := c.manager.GetRoom(c.roomID); processor != nil {
		// 检查是观战者还是参与者
		if processor.IsSpectator(c.userID) {
			processor.RemoveSpectator(c.userID)
		} else {
			// 主动离开房间：完全移除玩家（不同于断线只标记离线）
			processor.RemovePlayer(c.userID)
		}
	}

	c.logger.Info("User left room (removed)", zap.Int64("room_id", c.roomID))
	c.roomID = 0
}

func (c *wsClient) handleSetAutoReady(payload interface{}) {
	data, _ := json.Marshal(payload)
	var req model.WSSetAutoReady
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError(400, "invalid payload")
		return
	}

	if c.roomID == 0 {
		c.sendError(400, "not in room")
		return
	}

	if processor := c.manager.GetRoom(c.roomID); processor != nil {
		processor.SetAutoReady(c.userID, req.AutoReady)
	}
}

func (c *wsClient) handleDisconnect() {
	// Record WebSocket disconnection metric
	metrics.RecordWSConnection(-1)

	if c.roomID != 0 {
		c.hub.RemoveConn(c.roomID, c.userID)
		if processor := c.manager.GetRoom(c.roomID); processor != nil {
			// 检查是观战者还是参与者
			if processor.IsSpectator(c.userID) {
				processor.RemoveSpectator(c.userID)
			} else {
				processor.SetPlayerOnline(c.userID, false)
			}
		}
	}
	c.logger.Info("User disconnected", zap.String("session_id", c.sessionID))
}

func (c *wsClient) sendError(code int, message string) {
	c.writeJSON(&model.WSMessage{
		Type: model.WSTypeError,
		Payload: &model.WSError{
			Code:    code,
			Message: message,
		},
	})
}

// handleJoinAsSpectator 处理以观战者身份加入房间
func (c *wsClient) handleJoinAsSpectator(payload interface{}) {
	data, _ := json.Marshal(payload)
	var req model.WSJoinRoom
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError(400, "invalid payload")
		return
	}

	// 如果已在其他房间,先离开
	if c.roomID != 0 && c.roomID != req.RoomID {
		c.handleLeaveRoom()
	}

	processor, err := c.manager.GetOrCreateRoom(context.Background(), req.RoomID)
	if err != nil {
		c.sendError(404, "room not found")
		return
	}

	// 获取用户信息
	user, err := c.userGetter.GetUserByID(context.Background(), c.userID)
	if err != nil {
		c.sendError(500, "failed to get user")
		return
	}

	// 添加为观战者
	if err := processor.AddSpectator(user); err != nil {
		switch err {
		case game.ErrAlreadyParticipant:
			c.sendError(400, "already a participant")
		case game.ErrAlreadySpectator:
			c.sendError(400, "already a spectator")
		case game.ErrSpectatorLimitReached:
			c.sendError(400, "spectator limit reached")
		default:
			c.sendError(500, err.Error())
		}
		return
	}

	c.roomID = req.RoomID
	c.hub.AddConn(req.RoomID, c.userID, &safeConn{conn: c.conn, writeMu: &c.writeMu})

	// 发送房间状态（包含观战者标识）
	state := processor.GetRoomStateForUser(c.userID)
	c.writeJSON(&model.WSMessage{
		Type:    model.WSTypeRoomState,
		Payload: state,
	})

	c.logger.Info("User joined room as spectator", zap.Int64("room_id", req.RoomID))
}

// handleSwitchToParticipant 处理观战者切换为参与者
func (c *wsClient) handleSwitchToParticipant() {
	if c.roomID == 0 {
		c.sendError(400, "not in room")
		return
	}

	processor := c.manager.GetRoom(c.roomID)
	if processor == nil {
		c.sendError(404, "room not found")
		return
	}

	// 检查是否是观战者
	if !processor.IsSpectator(c.userID) {
		c.sendError(400, "not a spectator")
		return
	}

	// 获取用户信息
	user, err := c.userGetter.GetUserByID(context.Background(), c.userID)
	if err != nil {
		c.sendError(500, "failed to get user")
		return
	}

	// 切换为参与者
	if err := processor.SpectatorToParticipant(user); err != nil {
		switch err {
		case game.ErrNotSpectator:
			c.sendError(400, "not a spectator")
		case game.ErrRoomFull:
			c.sendError(400, "room is full")
		default:
			c.sendError(500, err.Error())
		}
		return
	}

	// 发送更新后的房间状态
	state := processor.GetRoomStateForUser(c.userID)
	c.writeJSON(&model.WSMessage{
		Type:    model.WSTypeRoomState,
		Payload: state,
	})

	c.logger.Info("Spectator switched to participant", zap.Int64("room_id", c.roomID))
}

// handleSendChat 处理发送聊天消息
func (c *wsClient) handleSendChat(payload interface{}) {
	if c.roomID == 0 {
		c.sendError(400, "not in room")
		return
	}

	if c.chatService == nil {
		c.sendError(500, "chat service unavailable")
		return
	}

	data, _ := json.Marshal(payload)
	var req model.WSSendChat
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError(400, "invalid payload")
		return
	}

	if req.Content == "" {
		c.sendError(400, "empty message")
		return
	}

	// 发送消息
	msg, err := c.chatService.SendMessage(context.Background(), c.roomID, c.userID, c.username, req.Content)
	if err != nil {
		switch err.Error() {
		case "chat rate limited":
			c.sendError(5005, "chat rate limit exceeded")
		default:
			c.sendError(500, err.Error())
		}
		return
	}

	// 广播消息给房间所有成员
	c.hub.BroadcastToRoom(c.roomID, &model.WSMessage{
		Type: model.WSTypeChatMessage,
		Payload: &model.WSChatMessage{
			ID:        msg.ID,
			UserID:    msg.UserID,
			Username:  msg.Username,
			Content:   msg.Content,
			Timestamp: msg.CreatedAt.UnixMilli(),
		},
	})

	c.logger.Debug("Chat message sent", zap.String("content", msg.Content))
}

// handleSendEmoji 处理发送表情
func (c *wsClient) handleSendEmoji(payload interface{}) {
	if c.roomID == 0 {
		c.sendError(400, "not in room")
		return
	}

	if c.chatService == nil {
		c.sendError(500, "chat service unavailable")
		return
	}

	data, _ := json.Marshal(payload)
	var req model.WSSendEmoji
	if err := json.Unmarshal(data, &req); err != nil {
		c.sendError(400, "invalid payload")
		return
	}

	// 验证表情
	if err := c.chatService.ValidateEmoji(req.Emoji); err != nil {
		c.sendError(400, "invalid emoji")
		return
	}

	// 检查限流
	if err := c.chatService.CheckEmojiRateLimit(c.userID); err != nil {
		c.sendError(5006, "emoji rate limit exceeded")
		return
	}

	// 广播表情给房间所有成员
	c.hub.BroadcastToRoom(c.roomID, &model.WSMessage{
		Type: model.WSTypeEmojiReaction,
		Payload: &model.WSEmojiReaction{
			UserID:   c.userID,
			Username: c.username,
			Emoji:    req.Emoji,
		},
	})

	c.logger.Debug("Emoji sent", zap.String("emoji", req.Emoji))
}
