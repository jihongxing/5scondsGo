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

// FriendHandler 好友处理器
type FriendHandler struct {
	friendService *service.FriendService
}

// NewFriendHandler 创建好友处理器
func NewFriendHandler(friendService *service.FriendService) *FriendHandler {
	return &FriendHandler{
		friendService: friendService,
	}
}

// SendFriendRequest 发送好友请求
func (h *FriendHandler) SendFriendRequest(c *gin.Context) {
	var req model.SendFriendRequestReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	friendReq, err := h.friendService.SendFriendRequest(c.Request.Context(), userID, req.ToUserID)
	if err != nil {
		switch {
		case errors.Is(err, service.ErrCannotAddSelf):
			c.JSON(http.StatusBadRequest, gin.H{"error": "cannot add yourself as friend", "code": 6001})
		case errors.Is(err, repository.ErrAlreadyFriends):
			c.JSON(http.StatusBadRequest, gin.H{"error": "already friends", "code": 6002})
		case errors.Is(err, repository.ErrFriendRequestExists):
			c.JSON(http.StatusBadRequest, gin.H{"error": "friend request already exists", "code": 6001})
		case errors.Is(err, repository.ErrFriendLimitReached):
			c.JSON(http.StatusBadRequest, gin.H{"error": "friend limit reached", "code": 6003})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusCreated, friendReq)
}

// AcceptFriendRequest 接受好友请求
func (h *FriendHandler) AcceptFriendRequest(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	userID := GetUserID(c)
	if err := h.friendService.AcceptFriendRequest(c.Request.Context(), requestID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrFriendRequestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "friend request not found", "code": 6004})
		case errors.Is(err, repository.ErrFriendLimitReached):
			c.JSON(http.StatusBadRequest, gin.H{"error": "friend limit reached", "code": 6003})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend request accepted"})
}

// RejectFriendRequest 拒绝好友请求
func (h *FriendHandler) RejectFriendRequest(c *gin.Context) {
	requestID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request id"})
		return
	}

	userID := GetUserID(c)
	if err := h.friendService.RejectFriendRequest(c.Request.Context(), requestID, userID); err != nil {
		switch {
		case errors.Is(err, repository.ErrFriendRequestNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "friend request not found", "code": 6004})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend request rejected"})
}

// RemoveFriend 删除好友
func (h *FriendHandler) RemoveFriend(c *gin.Context) {
	friendID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid friend id"})
		return
	}

	userID := GetUserID(c)
	if err := h.friendService.RemoveFriend(c.Request.Context(), userID, friendID); err != nil {
		switch {
		case errors.Is(err, repository.ErrNotFriends):
			c.JSON(http.StatusBadRequest, gin.H{"error": "not friends"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "friend removed"})
}

// GetFriendList 获取好友列表
func (h *FriendHandler) GetFriendList(c *gin.Context) {
	userID := GetUserID(c)
	friends, err := h.friendService.GetFriendList(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, friends)
}

// GetPendingRequests 获取待处理的好友请求
func (h *FriendHandler) GetPendingRequests(c *gin.Context) {
	userID := GetUserID(c)
	requests, err := h.friendService.GetPendingRequests(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, requests)
}
