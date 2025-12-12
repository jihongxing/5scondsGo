package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/fiveseconds/server/internal/cache"
	"github.com/fiveseconds/server/internal/config"
	"github.com/fiveseconds/server/internal/game"
	"github.com/fiveseconds/server/internal/handler"
	"github.com/fiveseconds/server/internal/middleware"
	"github.com/fiveseconds/server/internal/model"
	"github.com/fiveseconds/server/internal/repository"
	"github.com/fiveseconds/server/internal/service"
	"github.com/fiveseconds/server/internal/ws"
	pkglogger "github.com/fiveseconds/server/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	// 加载配置
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Failed to load config: %v", err))
	}

	// 初始化结构化日志
	logLevel := "info"
	logFormat := "json"
	if cfg.Server.Mode != "release" {
		logLevel = "debug"
		logFormat = "console"
	}
	structuredLogger, err := pkglogger.New(&pkglogger.Config{
		Level:       logLevel,
		Format:      logFormat,
		ServiceName: "fiveseconds",
		Environment: cfg.Server.Mode,
	})
	if err != nil {
		panic(fmt.Sprintf("Failed to init structured logger: %v", err))
	}
	defer structuredLogger.Sync()

	// 初始化传统日志（用于兼容现有代码）
	zapLogger := initLogger(cfg.Server.Mode)
	defer zapLogger.Sync()

	// 初始化数据库
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		cfg.Database.User, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	if err := repository.InitDB(dsn, zapLogger); err != nil {
		zapLogger.Fatal("Failed to init database", zap.Error(err))
	}
	defer repository.CloseDB()

	// 初始化数据库日志器
	dbLogger := repository.NewDBLogger(structuredLogger)

	// 初始化 Redis
	if err := cache.InitRedis(&cfg.Redis, zapLogger); err != nil {
		zapLogger.Warn("Failed to init Redis, balance cache will be disabled", zap.Error(err))
	} else {
		defer cache.CloseRedis()
	}

	// 初始化房间活动日志器
	roomActivityLogger := service.NewRoomActivityLogger(structuredLogger)

	// 初始化对账日志器（alertManager 稍后设置）
	reconciliationLogger := service.NewReconciliationLogger(structuredLogger, nil)

	// Keep dbLogger and roomActivityLogger and reconciliationLogger for future use
	_ = dbLogger
	_ = roomActivityLogger
	_ = reconciliationLogger

	// 初始化仓库
	userRepo := repository.NewUserRepo()
	roomRepo := repository.NewRoomRepo()
	gameRepo := repository.NewGameRepo()
	txRepo := repository.NewTransactionRepo()
	fundRepo := repository.NewFundRequestRepo()
	platformRepo := repository.NewPlatformRepo()
	conservationRepo := repository.NewConservationRepo()
	chatRepo := repository.NewChatRepo()
	riskRepo := repository.NewRiskRepo()
	alertRepo := repository.NewAlertRepo()
	metricsRepo := repository.NewMetricsRepo()
	themeRepo := repository.NewThemeRepo()
	friendRepo := repository.NewFriendRepo()
	invitationRepo := repository.NewInvitationRepo()

	// 初始化 WebSocket Hub
	hub := ws.NewHub()

	// 初始化余额缓存
	var balanceCache *cache.BalanceCache
	if cache.RedisClient != nil {
		balanceCache = cache.NewBalanceCache(cache.RedisClient, userRepo, zapLogger)
	}

	// 初始化告警管理器和风控服务
	alertManager := service.NewAlertManager(alertRepo, nil, zapLogger) // broadcaster 稍后设置
	riskService := service.NewRiskControlService(riskRepo, alertManager, zapLogger)

	// 初始化资金异常日志器
	fundAnomalyLogger := service.NewFundAnomalyLogger(structuredLogger, alertManager)
	_ = fundAnomalyLogger

	// 初始化游戏管理器
	manager := game.NewManager(hub, userRepo, roomRepo, gameRepo, txRepo, platformRepo, balanceCache, riskService, zapLogger)

	// 设置 Hub 断开连接回调，用于清理游戏引擎中的用户状态
	hub.SetDisconnectCallback(func(roomID, userID int64, reason string) {
		zapLogger.Info("Hub disconnect callback triggered",
			zap.Int64("room_id", roomID),
			zap.Int64("user_id", userID),
			zap.String("reason", reason),
		)
		if processor := manager.GetRoom(roomID); processor != nil {
			// 检查是观战者还是参与者
			if processor.IsSpectator(userID) {
				processor.RemoveSpectator(userID)
			} else {
				// 标记玩家离线（不完全移除，允许重连）
				processor.SetPlayerOnline(userID, false)
			}
		}
	})

	// 初始化服务
	authService := service.NewAuthService(userRepo, cfg)
	authService.SetRiskService(riskService) // 设置风控服务用于设备指纹检测
	roomService := service.NewRoomService(roomRepo, userRepo, manager)
	fundService := service.NewFundService(userRepo, fundRepo, txRepo, platformRepo, conservationRepo, cfg)
	fundService.SetHub(hub) // 设置 WebSocket Hub 用于发送余额更新通知
	chatService := service.NewChatService(chatRepo, zapLogger)

	// 启动资金守恒自动对账任务（每2小时一次）
	startConservationAutoCheck(fundService, zapLogger)
	// 启动每日按房主维度对账任务
	startDailyOwnerConservationCheck(fundService, zapLogger)

	// 初始化游戏历史服务
	gameHistoryService := service.NewGameHistoryService(gameRepo, zapLogger)

	// 初始化监控服务
	monitoringService := service.NewMonitoringService(metricsRepo, nil, zapLogger)
	monitoringService.Start()
	defer monitoringService.Stop()

	// 初始化主题服务
	themeService := service.NewThemeService(themeRepo, roomRepo, hub, zapLogger)

	// 初始化好友服务
	friendService := service.NewFriendService(friendRepo)

	// 初始化邀请服务
	invitationService := service.NewInvitationService(invitationRepo, friendRepo, roomRepo, userRepo, hub)

	// 初始化钱包服务
	walletService := service.NewWalletService(userRepo, txRepo)

	// 初始化处理器
	h := handler.NewHandler(authService, roomService, fundService)
	walletHandler := handler.NewWalletHandler(walletService)
	gameHistoryHandler := handler.NewGameHistoryHandler(gameHistoryService)
	monitoringHandler := handler.NewMonitoringHandler(monitoringService)
	themeHandler := handler.NewThemeHandler(themeService)
	riskHandler := handler.NewRiskHandler(riskService)
	alertHandler := handler.NewAlertHandler(alertManager)
	friendHandler := handler.NewFriendHandler(friendService)
	invitationHandler := handler.NewInvitationHandler(invitationService, roomService)
	authMiddleware := handler.NewMiddleware(authService)
	wsHandler := handler.NewWSHandler(hub, manager, authService, authService, chatService, zapLogger)

	// 初始化日志级别处理器
	logLevelHandler := handler.NewLogLevelHandler(structuredLogger)

	// 初始化 Gin
	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}
	r := gin.New()
	r.Use(middleware.RecoveryWithLogging(structuredLogger))
	r.Use(middleware.RequestLogging(structuredLogger))
	r.Use(handler.CORS())

	// 路由
	setupRoutes(r, h, walletHandler, gameHistoryHandler, monitoringHandler, themeHandler, riskHandler, alertHandler, friendHandler, invitationHandler, authMiddleware, wsHandler, logLevelHandler)

	// Prometheus 指标
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// 启动服务器
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Server.Port),
		Handler: r,
	}

	go func() {
		zapLogger.Info("Server starting", zap.Int("port", cfg.Server.Port))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// 优雅关闭
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	zapLogger.Info("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	manager.Shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	zapLogger.Info("Server exited")
}

func setupRoutes(r *gin.Engine, h *handler.Handler, wh *handler.WalletHandler, gh *handler.GameHistoryHandler, mh *handler.MonitoringHandler, th *handler.ThemeHandler, rh *handler.RiskHandler, ah *handler.AlertHandler, fh *handler.FriendHandler, ih *handler.InvitationHandler, m *handler.Middleware, wsHandler *handler.WSHandler, llh *handler.LogLevelHandler) {
	api := r.Group("/api")
	{
		// 公开接口
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)

		// 需要认证的接口
		auth := api.Group("")
		auth.Use(m.Auth())
		{
			// 用户
			auth.GET("/me", h.GetMe)
			auth.PUT("/me/language", h.UpdateLanguage)

			// 房间
			auth.GET("/rooms", h.ListRooms)
			auth.GET("/rooms/:id", h.GetRoom)
			auth.POST("/rooms/:id/join", h.JoinRoom)
			auth.POST("/rooms/:id/spectate", h.JoinAsSpectator)
			auth.POST("/rooms/:id/switch-to-participant", h.SwitchToParticipant)
			auth.POST("/rooms/leave", h.LeaveRoom)
			auth.POST("/rooms/auto-ready", h.SetAutoReady)
			auth.GET("/rooms/my", h.GetMyRoom)
			auth.GET("/rooms/:id/theme", th.GetRoomTheme)
			auth.GET("/themes", th.GetAllThemes)

			// 资金
			auth.POST("/fund-requests", h.CreateFundRequest)
			auth.GET("/fund-requests", h.ListFundRequests)
			auth.GET("/transactions", h.ListTransactions)
			auth.GET("/fund-summary", h.GetFundSummary)

			// 钱包
			auth.GET("/wallet", wh.GetWallet)
			auth.GET("/wallet/transactions", wh.GetTransactions)
			auth.GET("/wallet/earnings", wh.GetEarnings)
			auth.POST("/wallet/transfer-earnings", wh.TransferEarnings)

			// 游戏历史
			auth.GET("/game-history", gh.GetGameHistory)
			auth.GET("/game-history/:id", gh.GetRoundDetail)
			auth.GET("/game-stats", gh.GetGameStats)
			auth.GET("/game-rounds/:id/replay", gh.GetReplayData)
			auth.GET("/game-rounds/:id/verify", gh.VerifyRound)

			// 好友
			auth.GET("/friends", fh.GetFriendList)
			auth.POST("/friends/request", fh.SendFriendRequest)
			auth.GET("/friends/requests", fh.GetPendingRequests)
			auth.POST("/friends/accept/:id", fh.AcceptFriendRequest)
			auth.POST("/friends/reject/:id", fh.RejectFriendRequest)
			auth.DELETE("/friends/:id", fh.RemoveFriend)

			// 邀请
			auth.GET("/invitations", ih.GetPendingInvitations)
			auth.POST("/rooms/:id/invite", ih.SendInvitation)
			auth.POST("/invitations/:id/accept", ih.AcceptInvitation)
			auth.POST("/invitations/:id/decline", ih.DeclineInvitation)
			auth.POST("/rooms/:id/invite-link", ih.CreateInviteLink)
			auth.POST("/invite/:code/join", ih.JoinByInviteLink)
		}

		// 房主接口
		owner := api.Group("/owner")
		owner.Use(m.Auth(), m.RequireRole(model.RoleOwner, model.RoleAdmin))
		{
			owner.POST("/rooms", h.CreateRoom)
			owner.PUT("/rooms/:id", h.UpdateRoom)
			owner.GET("/rooms", h.ListMyRooms)
			owner.GET("/players", h.ListOwnerPlayers)
			owner.PUT("/rooms/:id/theme", th.UpdateRoomTheme)
			owner.GET("/fund-requests", h.ListOwnerFundRequests)
			owner.POST("/fund-requests/:id/process", h.ProcessOwnerFundRequest)
		}

		// 管理员接口
		admin := api.Group("/admin")
		admin.Use(m.Auth(), m.RequireRole(model.RoleAdmin))
		{
			admin.GET("/users", h.ListUsers)
			admin.POST("/owners", h.CreateOwner)
			admin.POST("/fund-requests/:id/process", h.ProcessFundRequest)
			admin.PUT("/rooms/:id/status", h.AdminUpdateRoomStatus)
			admin.GET("/platform", h.GetPlatformAccount)
			admin.GET("/conservation", h.CheckConservation)
			// 资金守恒检查报表（对账详情 + 对账类目汇总）
			admin.GET("/reports/balance-check", h.GetBalanceCheckReport)
			// 资金对账历史（全局 + 房主）
			admin.GET("/reports/balance-check/history", h.ListBalanceCheckHistory)
			// 监控指标
			admin.GET("/metrics/realtime", mh.GetRealtimeMetrics)
			admin.GET("/metrics/history", mh.GetHistoricalMetrics)
			// 风控标记
			admin.GET("/risk-flags", rh.ListRiskFlags)
			admin.GET("/risk-flags/:id", rh.GetRiskFlag)
			admin.POST("/risk-flags/:id/review", rh.ReviewRiskFlag)
			// 告警
			admin.GET("/alerts", ah.ListAlerts)
			admin.GET("/alerts/summary", ah.GetAlertSummary)
			admin.GET("/alerts/:id", ah.GetAlert)
			admin.POST("/alerts/:id/acknowledge", ah.AcknowledgeAlert)
			// 日志级别管理
			admin.GET("/log-level", llh.GetLogLevel)
			admin.PUT("/log-level", llh.SetLogLevel)
		}
	}

	// WebSocket
	r.GET("/ws", wsHandler.HandleWS)
}

func initLogger(mode string) *zap.Logger {
	var cfg zap.Config
	if mode == "release" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	logger, err := cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("Failed to init logger: %v", err))
	}
	return logger
}

// startConservationAutoCheck 启动资金守恒自动对账任务（每2小时一次）
func startConservationAutoCheck(fundService *service.FundService, logger *zap.Logger) {
	go func() {
		ticker := time.NewTicker(2 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UTC()
			periodEnd := now
			periodStart := now.Add(-2 * time.Hour)

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			check, err := fundService.CheckConservation(ctx)
			if err != nil {
				cancel()
				logger.Error("auto conservation check failed", zap.Error(err))
				continue
			}

			// 记录全局对账历史 + 房主维度 2 小时对账历史
			_ = fundService.RecordGlobalConservation(ctx, "2h", periodStart, periodEnd, check)
			_ = fundService.RecordOwnerConservation2h(ctx, periodStart, periodEnd)
			cancel()

			if !check.IsBalanced {
				logger.Warn("funds imbalance detected in auto check",
					zap.String("total_player_balance", check.TotalPlayerBalance.String()),
					zap.String("total_custody_quota", check.TotalCustodyQuota.String()),
					zap.String("total_margin", check.TotalMargin.String()),
					zap.String("platform_balance", check.PlatformBalance.String()),
					zap.String("difference", check.Difference.String()),
				)
			} else {
				logger.Info("auto conservation check passed",
					zap.String("total_player_balance", check.TotalPlayerBalance.String()),
					zap.String("total_custody_quota", check.TotalCustodyQuota.String()),
					zap.String("total_margin", check.TotalMargin.String()),
					zap.String("platform_balance", check.PlatformBalance.String()),
					zap.String("difference", check.Difference.String()),
				)
			}
		}
	}()
}

// startDailyOwnerConservationCheck 启动每日房主维度对账任务
// 简化实现: 每 24 小时执行一次, 以任务触发时间为当日区间
func startDailyOwnerConservationCheck(fundService *service.FundService, logger *zap.Logger) {
	go func() {
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for range ticker.C {
			now := time.Now().UTC()
			dayEnd := now
			dayStart := now.Add(-24 * time.Hour)

			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			if err := fundService.RecordOwnerConservationDaily(ctx, dayStart, dayEnd); err != nil {
				logger.Error("daily owner conservation check failed", zap.Error(err))
			}
			cancel()
		}
	}()
}
