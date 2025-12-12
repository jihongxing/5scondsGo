package repository

import (
	"context"
	"errors"

	"github.com/fiveseconds/server/internal/model"
	"github.com/jackc/pgx/v5"
)

// ThemeRepo 主题仓库
type ThemeRepo struct{}

// NewThemeRepo 创建主题仓库
func NewThemeRepo() *ThemeRepo {
	return &ThemeRepo{}
}

// GetByRoomID 根据房间ID获取主题
func (r *ThemeRepo) GetByRoomID(ctx context.Context, roomID int64) (*model.RoomTheme, error) {
	sql := `SELECT id, room_id, theme_name, updated_at
		FROM room_themes WHERE room_id = $1`
	
	theme := &model.RoomTheme{}
	err := DB.QueryRow(ctx, sql, roomID).Scan(
		&theme.ID, &theme.RoomID, &theme.ThemeName, &theme.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil // 没有设置主题，返回nil
	}
	if err != nil {
		return nil, err
	}
	return theme, nil
}

// Upsert 创建或更新主题
func (r *ThemeRepo) Upsert(ctx context.Context, roomID int64, themeName model.ThemeName) (*model.RoomTheme, error) {
	sql := `INSERT INTO room_themes (room_id, theme_name, updated_at)
		VALUES ($1, $2, NOW())
		ON CONFLICT (room_id) DO UPDATE SET theme_name = $2, updated_at = NOW()
		RETURNING id, room_id, theme_name, updated_at`
	
	theme := &model.RoomTheme{}
	err := DB.QueryRow(ctx, sql, roomID, themeName).Scan(
		&theme.ID, &theme.RoomID, &theme.ThemeName, &theme.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return theme, nil
}

// Delete 删除主题
func (r *ThemeRepo) Delete(ctx context.Context, roomID int64) error {
	sql := `DELETE FROM room_themes WHERE room_id = $1`
	_, err := DB.Exec(ctx, sql, roomID)
	return err
}
