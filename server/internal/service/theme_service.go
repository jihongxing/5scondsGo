package service

import (
	"context"
	"errors"

	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"go.uber.org/zap"
)

var (
	ErrInvalidTheme = errors.New("invalid theme name")
)

// ThemeBroadcaster 主题广播接口
type ThemeBroadcaster interface {
	BroadcastToRoom(roomID int64, msg *model.WSMessage)
}

// ThemeService 主题服务
type ThemeService struct {
	themeRepo   *repository.ThemeRepo
	roomRepo    *repository.RoomRepo
	broadcaster ThemeBroadcaster
	logger      *zap.Logger
}

// NewThemeService 创建主题服务
func NewThemeService(
	themeRepo *repository.ThemeRepo,
	roomRepo *repository.RoomRepo,
	broadcaster ThemeBroadcaster,
	logger *zap.Logger,
) *ThemeService {
	return &ThemeService{
		themeRepo:   themeRepo,
		roomRepo:    roomRepo,
		broadcaster: broadcaster,
		logger:      logger.With(zap.String("service", "theme")),
	}
}

// GetRoomTheme 获取房间主题
func (s *ThemeService) GetRoomTheme(ctx context.Context, roomID int64) (*model.RoomTheme, error) {
	theme, err := s.themeRepo.GetByRoomID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	
	// 如果没有设置主题，返回默认主题
	if theme == nil {
		return &model.RoomTheme{
			RoomID:    roomID,
			ThemeName: model.ThemeClassic,
		}, nil
	}
	
	return theme, nil
}

// UpdateRoomTheme 更新房间主题
func (s *ThemeService) UpdateRoomTheme(ctx context.Context, roomID int64, ownerID int64, themeName string) (*model.RoomTheme, error) {
	// 验证主题名称
	if !model.IsValidTheme(themeName) {
		return nil, ErrInvalidTheme
	}
	
	// 验证房间所有权
	room, err := s.roomRepo.GetByID(ctx, roomID)
	if err != nil {
		return nil, err
	}
	if room.OwnerID != ownerID {
		return nil, errors.New("not room owner")
	}
	
	// 更新主题
	theme, err := s.themeRepo.Upsert(ctx, roomID, model.ThemeName(themeName))
	if err != nil {
		return nil, err
	}
	
	// 广播主题变更
	s.broadcastThemeChange(roomID, themeName)
	
	s.logger.Info("Room theme updated",
		zap.Int64("room_id", roomID),
		zap.String("theme", themeName))
	
	return theme, nil
}

// broadcastThemeChange 广播主题变更
func (s *ThemeService) broadcastThemeChange(roomID int64, themeName string) {
	if s.broadcaster == nil {
		return
	}
	
	s.broadcaster.BroadcastToRoom(roomID, &model.WSMessage{
		Type: model.WSTypeThemeChange,
		Payload: &model.WSThemeChange{
			RoomID:    roomID,
			ThemeName: themeName,
		},
	})
}

// GetAllThemes 获取所有可用主题
func (s *ThemeService) GetAllThemes() []model.ThemeConfig {
	return model.GetThemeConfigs()
}
