package service

import (
	"context"
	"errors"
	"time"

	"github.com/fiveseconds/server/internal/config"
	"github.com/fiveseconds/server/internal/game"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"

	"github.com/golang-jwt/jwt/v5"
	"github.com/shopspring/decimal"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("username already exists")
	ErrInvalidInviteCode  = errors.New("invalid invite code")
	ErrInvalidToken       = errors.New("invalid token")
)

type AuthService struct {
	userRepo    *repository.UserRepo
	cfg         *config.Config
	riskService *RiskControlService
}

func NewAuthService(userRepo *repository.UserRepo, cfg *config.Config) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// SetRiskService 设置风控服务（用于设备指纹检测）
func (s *AuthService) SetRiskService(riskService *RiskControlService) {
	s.riskService = riskService
}

// Claims JWT claims
type Claims struct {
	UserID   int64      `json:"user_id"`
	Username string     `json:"username"`
	Role     model.Role `json:"role"`
	jwt.RegisteredClaims
}

// Register 用户注册
func (s *AuthService) Register(ctx context.Context, req *model.RegisterReq) (*model.User, error) {
	// 检查用户名是否存在
	exists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// 角色判断
	targetRole := model.RolePlayer
	if req.Role == "owner" {
		targetRole = model.RoleOwner
	}

	// 验证邀请码
	var invitedBy *int64
	var userInviteCode *string

	// 玩家必须有邀请码（绑定房主）
	// 房主必须有邀请码（绑定上级/Admin）
	if req.InviteCode == "" {
		// 暂时所有注册都要求必填
		return nil, ErrInvalidInviteCode
	}

	// 查找邀请人
	inviter, err := s.userRepo.GetByInviteCodeAllRoles(ctx, req.InviteCode)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidInviteCode
		}
		return nil, err
	}

	// 校验绑定逻辑
	if targetRole == model.RolePlayer {
		// 玩家只能被房主邀请 (或 Admin?)
		if inviter.Role != model.RoleOwner && inviter.Role != model.RoleAdmin {
			return nil, ErrInvalidInviteCode // 只能填房主/管理员的码
		}
		invitedBy = &inviter.ID
	} else if targetRole == model.RoleOwner {
		// 房主只能被 Admin (或 上级房主/代理) 邀请
		if inviter.Role != model.RoleAdmin {
			// 暂时只允许 Admin 邀请房主
			return nil, ErrInvalidInviteCode
		}
		invitedBy = &inviter.ID

		// 新房主自动生成自己的邀请码
		for {
			code, err := game.GenerateInviteCode()
			if err != nil {
				return nil, err
			}
			exists, err := s.userRepo.InviteCodeExists(ctx, code)
			if err != nil {
				return nil, err
			}
			if !exists {
				userInviteCode = &code
				break
			}
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:           req.Username,
		PasswordHash:       string(hashedPassword),
		Role:               targetRole,
		InvitedBy:          invitedBy,
		InviteCode:         userInviteCode,
		Balance:            decimal.Zero,
		FrozenBalance:      decimal.Zero,
		OwnerRoomBalance:   decimal.Zero,
		OwnerMarginBalance: decimal.Zero,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *AuthService) Login(ctx context.Context, req *model.LoginReq) (*model.LoginResp, error) {
	user, err := s.userRepo.GetByUsername(ctx, req.Username)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrInvalidCredentials
		}
		return nil, err
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, ErrInvalidCredentials
	}

	// 更新设备指纹并进行风控检测
	if req.DeviceFingerprint != "" {
		// 更新用户的设备指纹
		if err := s.userRepo.UpdateDeviceFingerprint(ctx, user.ID, req.DeviceFingerprint); err != nil {
			// 记录错误但不阻止登录
			// 可以添加日志记录
		}

		// 进行设备指纹风控检测（多账户检测）
		if s.riskService != nil {
			// 异步执行风控检测，不阻塞登录流程
			go func() {
				checkCtx := context.Background()
				s.riskService.CheckDeviceFingerprint(checkCtx, user.ID, req.DeviceFingerprint)
			}()
		}
	}

	// 生成 JWT
	token, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	return &model.LoginResp{
		Token: token,
		User:  user,
	}, nil
}

// generateToken 生成 JWT
func (s *AuthService) generateToken(user *model.User) (string, error) {
	claims := &Claims{
		UserID:   user.ID,
		Username: user.Username,
		Role:     user.Role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(s.cfg.Auth.JWTExpire)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.Auth.JWTSecret))
}

// ValidateToken 验证 JWT
func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.cfg.Auth.JWTSecret), nil
	})

	if err != nil {
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// CreateOwner 创建房主(仅管理员可用)
func (s *AuthService) CreateOwner(ctx context.Context, req *model.CreateOwnerReq) (*model.User, error) {
	// 检查用户名
	exists, err := s.userRepo.UsernameExists(ctx, req.Username)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserExists
	}

	// 生成邀请码
	var inviteCode string
	for {
		code, err := game.GenerateInviteCode()
		if err != nil {
			return nil, err
		}
		exists, err := s.userRepo.InviteCodeExists(ctx, code)
		if err != nil {
			return nil, err
		}
		if !exists {
			inviteCode = code
			break
		}
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	user := &model.User{
		Username:           req.Username,
		PasswordHash:       string(hashedPassword),
		Role:               model.RoleOwner,
		InviteCode:         &inviteCode,
		Balance:            decimal.Zero,
		FrozenBalance:      decimal.Zero,
		OwnerRoomBalance:   decimal.Zero,
		OwnerMarginBalance: decimal.Zero,
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	return user, nil
}

// GetUserByID 获取用户
func (s *AuthService) GetUserByID(ctx context.Context, id int64) (*model.User, error) {
	return s.userRepo.GetByID(ctx, id)
}

// ListUsers 列表用户
func (s *AuthService) ListUsers(ctx context.Context, query *model.UserListQuery) ([]*model.User, int64, error) {
	return s.userRepo.List(ctx, query)
}


// ListOwnerPlayers 获取房主名下玩家列表
func (s *AuthService) ListOwnerPlayers(ctx context.Context, ownerID int64) ([]*model.PlayerStat, error) {
	return s.userRepo.ListOwnerPlayers(ctx, ownerID)
}

// UpdateLanguage 更新用户语言偏好
func (s *AuthService) UpdateLanguage(ctx context.Context, userID int64, language string) error {
	return s.userRepo.UpdateLanguage(ctx, userID, language)
}
