package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/fiveseconds/server/internal/service"

	"github.com/gin-gonic/gin"
)

// InvitationHandler 邀请处理器
type InvitationHandler struct {
	invitationService *service.InvitationService
	roomService       *service.RoomService
}

// NewInvitationHandler 创建邀请处理器
func NewInvitationHandler(invitationService *service.InvitationService, roomService *service.RoomService) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		roomService:       roomService,
	}
}

// SendInvitation 发送房间邀请
func (h *InvitationHandler) SendInvitation(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req model.SendInvitationReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	inv, err := h.invitationService.SendInvitation(c.Request.Context(), roomID, userID, req.ToUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotInviteSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot invite yourself"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, inv)
}

// AcceptInvitation 接受邀请
func (h *InvitationHandler) AcceptInvitation(c *gin.Context) {
	invitationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}

	userID := GetUserID(c)
	inv, err := h.invitationService.AcceptInvitation(c.Request.Context(), invitationID, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInvitationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
		case errors.Is(err, service.ErrInvitationExpired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invitation expired", "code": 7001})
		case errors.Is(err, service.ErrNotInvitationTarget):
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	// 自动加入房间
	if err := h.roomService.JoinRoom(c.Request.Context(), userID, inv.RoomID, ""); err != nil {
		// 加入失败，但邀请已接受
		c.JSON(http.StatusOK, gin.H{
			"message":    "invitation accepted but failed to join room",
			"error":      err.Error(),
			"invitation": inv,
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation accepted and joined room", "invitation": inv})
}

// DeclineInvitation 拒绝邀请
func (h *InvitationHandler) DeclineInvitation(c *gin.Context) {
	invitationID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invitation id"})
		return
	}

	userID := GetUserID(c)
	if err := h.invitationService.DeclineInvitation(c.Request.Context(), invitationID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrInvitationNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "invitation not found"})
		case errors.Is(err, service.ErrNotInvitationTarget):
			c.JSON(http.StatusForbidden, gin.H{"error": "not authorized"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "invitation declined"})
}

// GetPendingInvitations 获取待处理的邀请
func (h *InvitationHandler) GetPendingInvitations(c *gin.Context) {
	userID := GetUserID(c)
	invitations, err := h.invitationService.GetPendingInvitations(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, invitations)
}

// CreateInviteLink 创建邀请链接
func (h *InvitationHandler) CreateInviteLink(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req model.CreateInviteLinkReq
	_ = c.ShouldBindJSON(&req) // 可选参数

	userID := GetUserID(c)
	link, err := h.invitationService.CreateInviteLink(c.Request.Context(), roomID, userID, req.MaxUses)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, &model.InviteLinkResponse{
		Code:      link.Code,
		Link:      "/invite/" + link.Code,
		ExpiresAt: link.ExpiresAt,
	})
}

// JoinByInviteLink 通过邀请链接加入房间
func (h *InvitationHandler) JoinByInviteLink(c *gin.Context) {
	code := c.Param("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid invite code", "code": 7002})
		return
	}

	userID := GetUserID(c)
	link, err := h.invitationService.UseInviteLink(c.Request.Context(), code, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrInviteLinkInvalid):
			c.JSON(http.StatusNotFound, gin.H{"error": "invalid invite link", "code": 7002})
		case errors.Is(err, repository.ErrInviteLinkExpired):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invite link expired", "code": 7001})
		case errors.Is(err, repository.ErrInviteLinkMaxUses):
			c.JSON(http.StatusBadRequest, gin.H{"error": "invite link max uses reached"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	// 加入房间
	if err := h.roomService.JoinRoom(c.Request.Context(), userID, link.RoomID, ""); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "joined room via invite link", "room_id": link.RoomID})
}
