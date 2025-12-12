package model

import "time"

// ThemeName 主题名称
type ThemeName string

const (
	ThemeClassic ThemeName = "classic"
	ThemeNeon    ThemeName = "neon"
	ThemeOcean   ThemeName = "ocean"
	ThemeForest  ThemeName = "forest"
	ThemeLuxury  ThemeName = "luxury"
)

// ValidThemes 有效主题列表
var ValidThemes = []ThemeName{
	ThemeClassic,
	ThemeNeon,
	ThemeOcean,
	ThemeForest,
	ThemeLuxury,
}

// IsValidTheme 检查主题是否有效
func IsValidTheme(name string) bool {
	for _, t := range ValidThemes {
		if string(t) == name {
			return true
		}
	}
	return false
}

// RoomTheme 房间主题
type RoomTheme struct {
	ID        int64     `json:"id" db:"id"`
	RoomID    int64     `json:"room_id" db:"room_id"`
	ThemeName ThemeName `json:"theme_name" db:"theme_name"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// ThemeConfig 主题配置
type ThemeConfig struct {
	Name            ThemeName `json:"name"`
	DisplayName     string    `json:"display_name"`
	PrimaryColor    string    `json:"primary_color"`
	SecondaryColor  string    `json:"secondary_color"`
	BackgroundColor string    `json:"background_color"`
	TextColor       string    `json:"text_color"`
}

// GetThemeConfigs 获取所有主题配置
func GetThemeConfigs() []ThemeConfig {
	return []ThemeConfig{
		{
			Name:            ThemeClassic,
			DisplayName:     "Classic",
			PrimaryColor:    "#6C5CE7",
			SecondaryColor:  "#A29BFE",
			BackgroundColor: "#FFFFFF",
			TextColor:       "#2D3436",
		},
		{
			Name:            ThemeNeon,
			DisplayName:     "Neon",
			PrimaryColor:    "#00F5FF",
			SecondaryColor:  "#FF00FF",
			BackgroundColor: "#0D0D0D",
			TextColor:       "#FFFFFF",
		},
		{
			Name:            ThemeOcean,
			DisplayName:     "Ocean",
			PrimaryColor:    "#0984E3",
			SecondaryColor:  "#74B9FF",
			BackgroundColor: "#DFE6E9",
			TextColor:       "#2D3436",
		},
		{
			Name:            ThemeForest,
			DisplayName:     "Forest",
			PrimaryColor:    "#00B894",
			SecondaryColor:  "#55EFC4",
			BackgroundColor: "#F0FFF4",
			TextColor:       "#2D3436",
		},
		{
			Name:            ThemeLuxury,
			DisplayName:     "Luxury",
			PrimaryColor:    "#FDCB6E",
			SecondaryColor:  "#F39C12",
			BackgroundColor: "#1A1A2E",
			TextColor:       "#FFFFFF",
		},
	}
}

// UpdateThemeReq 更新主题请求
type UpdateThemeReq struct {
	ThemeName string `json:"theme_name" binding:"required"`
}
