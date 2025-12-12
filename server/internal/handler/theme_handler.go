package handler

import (
	"net/http"
	"strconv"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/service"
	"github.com/gin-gonic/gin"
)

// ThemeHandler 主题处理器
type ThemeHandler struct {
	themeService *service.ThemeService
}

// NewThemeHandler 创建主题处理器
func NewThemeHandler(themeService *service.ThemeService) *ThemeHandler {
	return &ThemeHandler{
		themeService: themeService,
	}
}

// GetRoomTheme 获取房间主题
func (h *ThemeHandler) GetRoomTheme(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	theme, err := h.themeService.GetRoomTheme(c.Request.Context(), roomID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, theme)
}

// UpdateRoomTheme 更新房间主题
func (h *ThemeHandler) UpdateRoomTheme(c *gin.Context) {
	roomID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid room id"})
		return
	}

	var req model.UpdateThemeReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := GetUserID(c)
	theme, err := h.themeService.UpdateRoomTheme(c.Request.Context(), roomID, userID, req.ThemeName)
	if err != nil {
		if err == service.ErrInvalidTheme {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid theme name"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, theme)
}

// GetAllThemes 获取所有可用主题
func (h *ThemeHandler) GetAllThemes(c *gin.Context) {
	themes := h.themeService.GetAllThemes()
	c.JSON(http.StatusOK, themes)
}
