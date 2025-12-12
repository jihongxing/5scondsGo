package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/fiveseconds/server/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	authService *service.AuthService
	roomService *service.RoomService
	fundService *service.FundService
}

func NewHandler(authService *service.AuthService, roomService *service.RoomService, fundService *service.FundService) *Handler {
	return &Handler{
		authService: authService,
		roomService: roomService,
		fundService: fundService,
	}
}

// ===== Auth =====

func (h *Handler) Register(c *gin.Context) {
	var req model.RegisterReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.Register(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		if errors.Is(err, service.ErrInvalidInviteCode) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite code"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	resp, err := h.authService.Login(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrInvalidCredentials) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *Handler) GetMe(c *gin.Context) {
	userID := GetUserID(c)
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, user)
}

// UpdateLanguage 更新用户语言偏好
func (h *Handler) UpdateLanguage(c *gin.Context) {
	userID := GetUserID(c)
	var req model.UpdateLanguageReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 验证语言代码
	validLanguages := map[string]bool{
		"en": true, "zh": true, "zh-TW": true, "ja": true, "ko": true,
	}
	if !validLanguages[req.Language] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "unsupported language"})
		return
	}

	if err := h.authService.UpdateLanguage(c.Request.Context(), userID, req.Language); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"language": req.Language})
}

func (h *Handler) ListUsers(c *gin.Context) {
	var query model.UserListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	users, total, err := h.authService.ListUsers(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": users, "total": total})
}

func (h *Handler) CreateOwner(c *gin.Context) {
	var req model.CreateOwnerReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	user, err := h.authService.CreateOwner(c.Request.Context(), &req)
	if err != nil {
		if errors.Is(err, service.ErrUserExists) {
			c.JSON(http.StatusConflict, gin.H{"error": "username already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, user)
}

func (h *Handler) ListOwnerPlayers(c *gin.Context) {
	userID := GetUserID(c)
	players, err := h.authService.ListOwnerPlayers(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, players)
}

// ===== Room =====

// AdminUpdateRoomStatus 管理后台更新房间状态（active/paused/locked）
func (h *Handler) AdminUpdateRoomStatus(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.UpdateRoomStatusReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.roomService.AdminUpdateRoomStatus(c.Request.Context(), roomID, req.Status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) CreateRoom(c *gin.Context) {
	var req model.CreateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bind error: " + err.Error()})
		return
	}

	// Debug: 打印请求参数
	fmt.Printf("[CreateRoom] name=%s, bet_amount=%s, winner_count=%d, max_players=%d, owner_commission=%s\n",
		req.Name, req.BetAmount, req.WinnerCount, req.MaxPlayers, req.OwnerCommissionRate)

	userID := GetUserID(c)
	room, err := h.roomService.CreateRoom(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "create room error: " + err.Error()})
		return
	}

	c.JSON(http.StatusCreated, room)
}

func (h *Handler) GetRoom(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	room, err := h.roomService.GetRoom(c.Request.Context(), roomID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "room not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, room)
}

func (h *Handler) ListRooms(c *gin.Context) {
	var query model.RoomListQuery
	// 手动绑定可选参数，避免 min=1 验证问题
	if ownerID := c.Query("owner_id"); ownerID != "" {
		if id, err := strconv.ParseInt(ownerID, 10, 64); err == nil {
			query.OwnerID = &id
		}
	}
	if status := c.Query("status"); status != "" {
		s := model.RoomStatus(status)
		query.Status = &s
	}
	query.Page = 1
	query.PageSize = 20
	if page := c.Query("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			query.Page = p
		}
	}
	if pageSize := c.Query("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			query.PageSize = ps
		}
	}

	// 获取当前用户
	userID := GetUserID(c)
	user, err := h.authService.GetUserByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 如果是 Player，只能看到关联房主的房间
	if user.Role == model.RolePlayer && user.InvitedBy != nil {
		query.InvitedBy = user.InvitedBy
	} else if user.Role == model.RoleOwner {
		// Owner 只能看自己的房间? 
		// 需求只说了“每个用户进入游戏大厅只能看到用户关联的房主创建的房间”
		// 既然是“进入游戏大厅”，房主自己进大厅应该看自己的。
		// 管理员看所有。
		// 如果 query 已经指定了 owner_id (Admin可能会传)，则不用覆盖。
		// 否则，如果是 Owner，默认看自己的。
		if query.OwnerID == nil {
			// query.OwnerID = &user.ID // 这里有歧义，如果 Owner 想看别人的房间呢？暂且认为 Owner 只看自己的。
			// 但前端 ListMyRooms 是另外一个接口。这里是公共大厅。
			// 保持现状，如果 Owner 进大厅，看所有?
			// 既然题目强调“用户关联的房主”，通常指 Player。Owner 应该能看到自己。
			// 我们假设 Owner 看自己。
			query.OwnerID = &user.ID
		}
	}

	rooms, total, err := h.roomService.ListRooms(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": rooms, "total": total})
}

func (h *Handler) ListMyRooms(c *gin.Context) {
	userID := GetUserID(c)
	rooms, err := h.roomService.ListOwnerRooms(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func (h *Handler) UpdateRoom(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.UpdateRoomReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	if err := h.roomService.UpdateRoom(c.Request.Context(), roomID, userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) JoinRoom(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := GetUserID(c)

	var req model.JoinRoomReq
	// 尝试绑定 Body，如果 Body 为空也不报错（密码可选）
	if err := c.ShouldBindJSON(&req); err != nil {
		// 忽略 Body 为空的错误，或者假设没有 Body 就是没有密码
	}

	fmt.Printf("[JoinRoom] userID=%d, roomID=%d, password=%s\n", userID, roomID, req.Password)

	if err := h.roomService.JoinRoom(c.Request.Context(), userID, roomID, req.Password); err != nil {
		fmt.Printf("[JoinRoom] error: %v\n", err)
		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid password"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "joined"})
}

func (h *Handler) LeaveRoom(c *gin.Context) {
	userID := GetUserID(c)
	if err := h.roomService.LeaveRoom(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "left"})
}

func (h *Handler) SetAutoReady(c *gin.Context) {
	var req model.WSSetAutoReady
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	if err := h.roomService.SetAutoReady(c.Request.Context(), userID, req.AutoReady); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "updated"})
}

func (h *Handler) GetMyRoom(c *gin.Context) {
	userID := GetUserID(c)
	rp, err := h.roomService.GetPlayerRoom(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if rp == nil {
		c.JSON(http.StatusOK, nil)
		return
	}
	room, err := h.roomService.GetRoom(c.Request.Context(), rp.RoomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, room)
}

// JoinAsSpectator 以观战者身份加入房间
func (h *Handler) JoinAsSpectator(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := GetUserID(c)

	var req model.JoinAsSpectatorReq
	// 尝试绑定 Body，如果 Body 为空也不报错（密码可选）
	_ = c.ShouldBindJSON(&req)

	if err := h.roomService.JoinAsSpectator(c.Request.Context(), userID, roomID, req.Password); err != nil {
		if errors.Is(err, service.ErrInvalidPassword) {
			c.JSON(http.StatusForbidden, gin.H{"error": "invalid password"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "joined as spectator"})
}

// SwitchToParticipant 观战者切换为参与者
func (h *Handler) SwitchToParticipant(c *gin.Context) {
	roomID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	userID := GetUserID(c)

	if err := h.roomService.SwitchToParticipant(c.Request.Context(), userID, roomID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "switched to participant"})
}

// ===== Fund =====

func (h *Handler) CreateFundRequest(c *gin.Context) {
	var req model.CreateFundRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	fundReq, err := h.fundService.CreateFundRequest(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, fundReq)
}

func (h *Handler) ListFundRequests(c *gin.Context) {
	var query model.FundRequestListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	// 非管理员只能查看自己的
	if GetRole(c) != model.RoleAdmin {
		userID := GetUserID(c)
		query.UserID = &userID
	}

	reqs, total, err := h.fundService.ListFundRequests(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": reqs, "total": total})
}

func (h *Handler) ProcessFundRequest(c *gin.Context) {
	reqID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.ProcessFundRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	if err := h.fundService.ProcessFundRequest(c.Request.Context(), reqID, userID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "processed"})
}

// ListOwnerFundRequests 获取 owner 下级玩家的资金申请
func (h *Handler) ListOwnerFundRequests(c *gin.Context) {
	var query model.FundRequestListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	// owner 只能查看自己下级玩家的申请
	ownerID := GetUserID(c)
	query.InvitedBy = &ownerID

	reqs, total, err := h.fundService.ListFundRequests(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": reqs, "total": total})
}

// ProcessOwnerFundRequest owner 审批下级玩家的资金申请
func (h *Handler) ProcessOwnerFundRequest(c *gin.Context) {
	reqID, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	var req model.ProcessFundRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ownerID := GetUserID(c)
	
	// 验证这个申请是否属于 owner 的下级玩家
	if err := h.fundService.ValidateOwnerFundRequest(c.Request.Context(), reqID, ownerID); err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	if err := h.fundService.ProcessFundRequest(c.Request.Context(), reqID, ownerID, &req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "processed"})
}

func (h *Handler) ListTransactions(c *gin.Context) {
	var query model.TransactionListQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	// 非管理员只能查看自己的
	if GetRole(c) != model.RoleAdmin {
		userID := GetUserID(c)
		query.UserID = &userID
	}

	txs, total, err := h.fundService.ListTransactions(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": txs, "total": total})
}

func (h *Handler) GetPlatformAccount(c *gin.Context) {
	acc, err := h.fundService.GetPlatformAccount(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acc)
}

func (h *Handler) CheckConservation(c *gin.Context) {
	check, err := h.fundService.CheckConservation(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, check)
}

// GetReconciliationReport 获取详细的资金对账报告
func (h *Handler) GetReconciliationReport(c *gin.Context) {
	report, err := h.fundService.GetReconciliationReport(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, report)
}

func (h *Handler) GetFundSummary(c *gin.Context) {
	var userID *int64
	if GetRole(c) != model.RoleAdmin {
		id := GetUserID(c)
		userID = &id
	}

	summary, err := h.fundService.GetFundSummary(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, summary)
}

// GetBalanceCheckReport 管理后台资金对账详情（资金守恒结果 + 对账类目汇总）
func (h *Handler) GetBalanceCheckReport(c *gin.Context) {
	ctx := c.Request.Context()

	check, err := h.fundService.CheckConservation(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	summary, err := h.fundService.GetFundSummary(ctx, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"check":   check,
		"summary": summary,
	})
}

// ListBalanceCheckHistory 查询资金对账历史（全局 + 房主维度）
// 支持按 scope/owner_id/period_type/time_range 分页过滤
func (h *Handler) ListBalanceCheckHistory(c *gin.Context) {
	var query model.FundConservationHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, total, err := h.fundService.ListConservationHistory(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": items,
		"total": total,
	})
}


// ===== Game History =====

// GameHistoryHandler 游戏历史处理器
type GameHistoryHandler struct {
	gameHistoryService *service.GameHistoryService
}

// NewGameHistoryHandler 创建游戏历史处理器
func NewGameHistoryHandler(gameHistoryService *service.GameHistoryService) *GameHistoryHandler {
	return &GameHistoryHandler{
		gameHistoryService: gameHistoryService,
	}
}

// GetGameHistory 获取游戏历史
func (h *GameHistoryHandler) GetGameHistory(c *gin.Context) {
	var query model.GameHistoryQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	query.UserID = GetUserID(c)
	if query.Page == 0 {
		query.Page = 1
	}
	if query.PageSize == 0 {
		query.PageSize = 20
	}

	items, total, err := h.gameHistoryService.GetHistory(c.Request.Context(), &query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"items": items, "total": total})
}

// GetGameStats 获取游戏统计
func (h *GameHistoryHandler) GetGameStats(c *gin.Context) {
	userID := GetUserID(c)

	stats, err := h.gameHistoryService.GetStats(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetRoundDetail 获取回合详情
func (h *GameHistoryHandler) GetRoundDetail(c *gin.Context) {
	roundID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid round id"})
		return
	}

	detail, err := h.gameHistoryService.GetRoundDetail(c.Request.Context(), roundID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "round not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, detail)
}

// GetReplayData 获取回放数据
func (h *GameHistoryHandler) GetReplayData(c *gin.Context) {
	roundID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid round id"})
		return
	}

	replay, err := h.gameHistoryService.GetReplayData(c.Request.Context(), roundID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "round not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, replay)
}

// VerifyRound 验证回合
func (h *GameHistoryHandler) VerifyRound(c *gin.Context) {
	roundID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid round id"})
		return
	}

	result, err := h.gameHistoryService.VerifyRound(c.Request.Context(), roundID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "round not found"})
			return
		}
		if errors.Is(err, service.ErrInvalidSeed) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "round not yet settled"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}
