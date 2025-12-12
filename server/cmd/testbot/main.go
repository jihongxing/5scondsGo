package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/shopspring/decimal"
)

// 配置
var (
	baseURL     = flag.String("url", "http://localhost:8080", "服务器地址")
	adminUser   = flag.String("admin", "admin", "管理员用户名")
	adminPass   = flag.String("admin-pass", "admin123", "管理员密码")
	playerCount = flag.Int("players", 3, "玩家数量")
	rounds      = flag.Int("rounds", 3, "游戏轮数")
	betAmount   = flag.String("bet", "10", "下注金额")
	initBalance = flag.String("balance", "1000", "初始余额")
)

// API 响应结构
type LoginResp struct {
	Token string    `json:"token"`
	User  *UserResp `json:"user"`
}

type UserResp struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Role     string `json:"role"`
}

type UserInfo struct {
	ID            int64  `json:"id"`
	Username      string `json:"username"`
	Role          string `json:"role"`
	Balance       string `json:"balance"`
	FrozenBalance string `json:"frozen_balance"`
	InviteCode    string `json:"invite_code,omitempty"`
}

type CreateRoomResp struct {
	ID         int64  `json:"id"`
	InviteCode string `json:"invite_code"`
	Name       string `json:"name"`
}

type FundRequest struct {
	ID     int64  `json:"id"`
	Status string `json:"status"`
}

// WebSocket 消息
type WSMessage struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type WSRoundResult struct {
	RoundID        int64    `json:"round_id"`
	Winners        []int64  `json:"winners"`
	WinnerNames    []string `json:"winner_names"`
	PrizePerWinner string   `json:"prize_per_winner"`
}

// 玩家状态
type Player struct {
	UserID   int64
	Username string
	Token    string
	Balance  decimal.Decimal
	WS       *websocket.Conn
	wsMu     sync.Mutex // 保护 WebSocket 写操作
}

// 测试机器人
type TestBot struct {
	httpClient *http.Client
	adminToken string
	owner      *Player
	players    []*Player
	roomID     int64
	roomCode   string
	
	roundsPlayed   int
	roundResults   []WSRoundResult
	seenRoundIDs   map[int64]bool // 已处理的轮次ID，避免重复计数
	mu             sync.Mutex
	done           chan struct{}
}

func main() {
	flag.Parse()
	
	log.Println("========================================")
	log.Println("5SecondsGo 测试机器人")
	log.Println("========================================")
	log.Printf("服务器: %s", *baseURL)
	log.Printf("玩家数量: %d", *playerCount)
	log.Printf("游戏轮数: %d", *rounds)
	log.Printf("下注金额: %s", *betAmount)
	log.Printf("初始余额: %s", *initBalance)
	log.Println("========================================")
	
	bot := &TestBot{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		players:      make([]*Player, 0),
		seenRoundIDs: make(map[int64]bool),
		done:         make(chan struct{}),
	}
	
	// 捕获中断信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("\n收到中断信号，正在清理...")
		close(bot.done)
	}()
	
	if err := bot.Run(); err != nil {
		log.Fatalf("测试失败: %v", err)
	}
}

func (b *TestBot) Run() error {
	// 1. 管理员登录
	log.Println("\n[步骤 1] 管理员登录...")
	if err := b.adminLogin(); err != nil {
		return fmt.Errorf("管理员登录失败: %w", err)
	}
	log.Println("✓ 管理员登录成功")
	
	// 2. 创建房主
	log.Println("\n[步骤 2] 创建房主账号...")
	if err := b.createOwner(); err != nil {
		return fmt.Errorf("创建房主失败: %w", err)
	}
	log.Printf("✓ 房主创建成功: %s (ID: %d)", b.owner.Username, b.owner.UserID)
	
	// 3. 为房主充值
	log.Println("\n[步骤 3] 为房主充值...")
	if err := b.depositOwner(); err != nil {
		return fmt.Errorf("房主充值失败: %w", err)
	}
	log.Printf("✓ 房主充值成功: %s", *initBalance)
	
	// 4. 创建玩家
	log.Println("\n[步骤 4] 创建玩家账号...")
	if err := b.createPlayers(); err != nil {
		return fmt.Errorf("创建玩家失败: %w", err)
	}
	for _, p := range b.players {
		log.Printf("✓ 玩家创建成功: %s (ID: %d)", p.Username, p.UserID)
	}
	
	// 5. 为玩家充值
	log.Println("\n[步骤 5] 为玩家充值...")
	if err := b.depositPlayers(); err != nil {
		return fmt.Errorf("玩家充值失败: %w", err)
	}
	log.Printf("✓ 所有玩家充值成功，每人: %s", *initBalance)
	
	// 6. 创建房间
	log.Println("\n[步骤 6] 创建游戏房间...")
	if err := b.createRoom(); err != nil {
		return fmt.Errorf("创建房间失败: %w", err)
	}
	log.Printf("✓ 房间创建成功: %s (ID: %d)", b.roomCode, b.roomID)
	
	// 7. 玩家加入房间并连接 WebSocket
	log.Println("\n[步骤 7] 玩家加入房间...")
	if err := b.playersJoinRoom(); err != nil {
		return fmt.Errorf("玩家加入房间失败: %w", err)
	}
	log.Println("✓ 所有玩家已加入房间")
	
	// 8. 设置自动准备并等待游戏
	log.Println("\n[步骤 8] 开始游戏...")
	if err := b.playGame(); err != nil {
		return fmt.Errorf("游戏过程出错: %w", err)
	}
	
	// 9. 分析结果
	log.Println("\n[步骤 9] 分析游戏结果...")
	b.analyzeResults()
	
	return nil
}


// adminLogin 管理员登录
func (b *TestBot) adminLogin() error {
	resp, err := b.post("/api/auth/login", map[string]string{
		"username": *adminUser,
		"password": *adminPass,
	}, "")
	if err != nil {
		return err
	}
	
	var loginResp LoginResp
	if err := json.Unmarshal(resp, &loginResp); err != nil {
		return err
	}
	
	b.adminToken = loginResp.Token
	return nil
}

// createOwner 创建房主（使用 admin 邀请码注册）
func (b *TestBot) createOwner() error {
	// 先获取 admin 的邀请码
	adminMeResp, err := b.get("/api/me", b.adminToken) // 路由是 /api/me 不是 /api/auth/me
	if err != nil {
		return fmt.Errorf("获取 admin 信息失败: %w", err)
	}
	
	var adminMe UserInfo
	if err := json.Unmarshal(adminMeResp, &adminMe); err != nil {
		return fmt.Errorf("解析 admin 信息失败: %w", err)
	}
	
	if adminMe.InviteCode == "" {
		return fmt.Errorf("admin 没有邀请码")
	}
	
	log.Printf("  使用 admin 邀请码: %s", adminMe.InviteCode)
	
	timestamp := time.Now().Unix()
	username := fmt.Sprintf("testowner_%d", timestamp)
	
	// 使用 admin 邀请码注册房主（需要指定 role: owner）
	_, err = b.post("/api/auth/register", map[string]interface{}{
		"username":    username,
		"password":    "test123456",
		"invite_code": adminMe.InviteCode,
		"role":        "owner",
	}, "")
	if err != nil {
		return fmt.Errorf("注册房主失败: %w", err)
	}
	
	// 注册后登录获取 token
	loginResp, err := b.post("/api/auth/login", map[string]string{
		"username": username,
		"password": "test123456",
	}, "")
	if err != nil {
		return fmt.Errorf("房主登录失败: %w", err)
	}
	
	var login LoginResp
	if err := json.Unmarshal(loginResp, &login); err != nil {
		return err
	}
	
	b.owner = &Player{
		UserID:   login.User.ID,
		Username: username,
		Token:    login.Token,
		Balance:  decimal.Zero,
	}
	
	return nil
}

// depositOwner 为房主充值
func (b *TestBot) depositOwner() error {
	// 计算房主需要的余额：每个玩家的初始余额 * 玩家数量
	initBal, _ := decimal.NewFromString(*initBalance)
	totalNeeded := initBal.Mul(decimal.NewFromInt(int64(*playerCount)))
	
	// 1. 充值房主可用余额（用于给玩家充值）
	resp, err := b.post("/api/fund-requests", map[string]interface{}{
		"type":   "owner_deposit",
		"amount": totalNeeded.String(),
		"remark": "测试机器人充值-可用余额",
	}, b.owner.Token)
	if err != nil {
		return fmt.Errorf("创建房主充值申请失败: %w", err)
	}
	
	var fundReq FundRequest
	if err := json.Unmarshal(resp, &fundReq); err != nil {
		return err
	}
	
	// 管理员审批
	_, err = b.post(fmt.Sprintf("/api/admin/fund-requests/%d/process", fundReq.ID), map[string]interface{}{
		"approved": true,
		"remark":   "测试机器人审批",
	}, b.adminToken)
	if err != nil {
		return fmt.Errorf("房主充值审批失败: %w", err)
	}
	
	log.Printf("  房主可用余额充值: %s", totalNeeded.String())
	
	// 2. 充值房主保证金（创建房间需要至少 2000）
	marginAmount := "2000"
	resp, err = b.post("/api/fund-requests", map[string]interface{}{
		"type":   "margin_deposit",
		"amount": marginAmount,
		"remark": "测试机器人充值-保证金",
	}, b.owner.Token)
	if err != nil {
		return fmt.Errorf("创建保证金充值申请失败: %w", err)
	}
	
	if err := json.Unmarshal(resp, &fundReq); err != nil {
		return err
	}
	
	// 管理员审批保证金
	_, err = b.post(fmt.Sprintf("/api/admin/fund-requests/%d/process", fundReq.ID), map[string]interface{}{
		"approved": true,
		"remark":   "测试机器人审批保证金",
	}, b.adminToken)
	if err != nil {
		return fmt.Errorf("保证金充值审批失败: %w", err)
	}
	
	log.Printf("  房主保证金充值: %s", marginAmount)
	
	return nil
}

// createPlayers 创建玩家
func (b *TestBot) createPlayers() error {
	// 先获取房主的邀请码
	meResp, err := b.get("/api/me", b.owner.Token)
	if err != nil {
		return err
	}
	
	var me UserInfo
	if err := json.Unmarshal(meResp, &me); err != nil {
		return err
	}
	
	inviteCode := me.InviteCode
	if inviteCode == "" {
		return fmt.Errorf("房主没有邀请码")
	}
	
	log.Printf("  使用房主邀请码: %s", inviteCode)
	
	timestamp := time.Now().Unix()
	for i := 0; i < *playerCount; i++ {
		username := fmt.Sprintf("testplayer_%d_%d", timestamp, i+1)
		
		// 注册玩家
		_, err := b.post("/api/auth/register", map[string]string{
			"username":    username,
			"password":    "test123456",
			"invite_code": inviteCode,
		}, "")
		if err != nil {
			return fmt.Errorf("注册玩家 %d 失败: %w", i+1, err)
		}
		
		// 登录获取 token
		loginResp, err := b.post("/api/auth/login", map[string]string{
			"username": username,
			"password": "test123456",
		}, "")
		if err != nil {
			return fmt.Errorf("玩家 %d 登录失败: %w", i+1, err)
		}
		
		var login LoginResp
		if err := json.Unmarshal(loginResp, &login); err != nil {
			return err
		}
		
		b.players = append(b.players, &Player{
			UserID:   login.User.ID,
			Username: username,
			Token:    login.Token,
			Balance:  decimal.Zero,
		})
	}
	
	return nil
}

// depositPlayers 为玩家充值
func (b *TestBot) depositPlayers() error {
	for _, player := range b.players {
		// 创建玩家充值申请
		resp, err := b.post("/api/fund-requests", map[string]interface{}{
			"type":   "deposit",
			"amount": *initBalance,
			"remark": "测试机器人充值",
		}, player.Token)
		if err != nil {
			return fmt.Errorf("玩家 %s 充值申请失败: %w", player.Username, err)
		}
		
		var fundReq FundRequest
		if err := json.Unmarshal(resp, &fundReq); err != nil {
			return err
		}
		
		// 管理员审批
		_, err = b.post(fmt.Sprintf("/api/admin/fund-requests/%d/process", fundReq.ID), map[string]interface{}{
			"approved": true,
			"remark":   "测试机器人审批",
		}, b.adminToken)
		if err != nil {
			return fmt.Errorf("玩家 %s 充值审批失败: %w", player.Username, err)
		}
	}
	
	return nil
}

// createRoom 创建房间
func (b *TestBot) createRoom() error {
	resp, err := b.post("/api/owner/rooms", map[string]interface{}{
		"name":                     "测试机器人房间",
		"bet_amount":               *betAmount,
		"winner_count":             1,
		"max_players":              10,
		"owner_commission_rate":    "0.03",
		"platform_commission_rate": "0.02",
	}, b.owner.Token)
	if err != nil {
		return err
	}
	
	var room CreateRoomResp
	if err := json.Unmarshal(resp, &room); err != nil {
		return err
	}
	
	b.roomID = room.ID
	b.roomCode = room.InviteCode
	
	return nil
}


// playersJoinRoom 玩家加入房间并连接 WebSocket
func (b *TestBot) playersJoinRoom() error {
	for _, player := range b.players {
		// 捕获循环变量，避免闭包问题
		p := player
		
		// HTTP 加入房间
		_, err := b.post(fmt.Sprintf("/api/rooms/%d/join", b.roomID), nil, p.Token)
		if err != nil {
			return fmt.Errorf("玩家 %s 加入房间失败: %w", p.Username, err)
		}
		
		// 连接 WebSocket
		wsURL := b.getWSURL(p.Token)
		conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
		if err != nil {
			return fmt.Errorf("玩家 %s WebSocket 连接失败: %w", p.Username, err)
		}
		
		// 设置 ping handler - 收到服务器 ping 时自动回复 pong
		conn.SetPingHandler(func(appData string) error {
			log.Printf("  玩家 %s 收到 ping，回复 pong", p.Username)
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			// 发送 pong 响应
			p.wsMu.Lock()
			err := conn.WriteControl(websocket.PongMessage, []byte(appData), time.Now().Add(10*time.Second))
			p.wsMu.Unlock()
			if err != nil {
				log.Printf("  玩家 %s 发送 pong 失败: %v", p.Username, err)
				return err
			}
			return nil
		})
		
		// 设置 pong handler，收到 pong 时重置读取超时
		conn.SetPongHandler(func(string) error {
			conn.SetReadDeadline(time.Now().Add(60 * time.Second))
			return nil
		})
		
		p.WS = conn
		
		// 发送加入房间消息
		joinMsg := map[string]interface{}{
			"type": "join_room",
			"payload": map[string]int64{
				"room_id": b.roomID,
			},
		}
		if err := conn.WriteJSON(joinMsg); err != nil {
			return fmt.Errorf("玩家 %s 发送加入房间消息失败: %w", p.Username, err)
		}
		
		log.Printf("  玩家 %s 已连接 WebSocket", p.Username)
	}
	
	return nil
}

// playGame 进行游戏
func (b *TestBot) playGame() error {
	// 为每个玩家启动消息监听和 ping 保活
	var wg sync.WaitGroup
	errChan := make(chan error, len(b.players))
	
	for _, player := range b.players {
		wg.Add(1)
		go func(p *Player) {
			defer wg.Done()
			if err := b.listenPlayer(p); err != nil {
				errChan <- fmt.Errorf("玩家 %s 监听出错: %w", p.Username, err)
			}
		}(player)
		
		// 启动 ping 保活 goroutine（每 20 秒发送一次心跳，确保在 60 秒超时前发送）
		go func(p *Player) {
			// 立即发送第一次心跳
			if p.WS != nil {
				heartbeat := map[string]interface{}{
					"type": "heartbeat",
				}
				p.wsMu.Lock()
				p.WS.WriteJSON(heartbeat)
				p.wsMu.Unlock()
			}
			
			pingTicker := time.NewTicker(20 * time.Second)
			defer pingTicker.Stop()
			for {
				select {
				case <-b.done:
					return
				case <-pingTicker.C:
					if p.WS != nil {
						// 发送心跳消息保持连接活跃
						heartbeat := map[string]interface{}{
							"type": "heartbeat",
						}
						p.wsMu.Lock()
						err := p.WS.WriteJSON(heartbeat)
						p.wsMu.Unlock()
						if err != nil {
							log.Printf("  玩家 %s 发送心跳失败: %v", p.Username, err)
							return
						}
					}
				}
			}
		}(player)
	}
	
	// 等待一小段时间让 WebSocket 连接稳定
	time.Sleep(500 * time.Millisecond)
	
	// 设置所有玩家自动准备
	log.Println("  设置所有玩家自动准备...")
	for _, player := range b.players {
		autoReadyMsg := map[string]interface{}{
			"type": "set_auto_ready",
			"payload": map[string]bool{
				"auto_ready": true,
			},
		}
		player.wsMu.Lock()
		err := player.WS.WriteJSON(autoReadyMsg)
		player.wsMu.Unlock()
		if err != nil {
			return fmt.Errorf("玩家 %s 设置自动准备失败: %w", player.Username, err)
		}
	}
	
	log.Println("  等待游戏进行...")
	log.Printf("  目标轮数: %d", *rounds)
	
	// 等待游戏完成或超时
	timeout := time.After(time.Duration(*rounds*30+60) * time.Second)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	
	for {
		select {
		case <-b.done:
			return fmt.Errorf("收到中断信号")
		case err := <-errChan:
			return err
		case <-timeout:
			return fmt.Errorf("游戏超时")
		case <-ticker.C:
			b.mu.Lock()
			played := b.roundsPlayed
			b.mu.Unlock()
			
			log.Printf("  已完成 %d/%d 轮", played, *rounds)
			
			if played >= *rounds {
				log.Println("✓ 游戏完成!")
				// 关闭所有 WebSocket 连接
				for _, player := range b.players {
					if player.WS != nil {
						player.WS.Close()
					}
				}
				return nil
			}
		}
	}
}

// listenPlayer 监听玩家 WebSocket 消息
func (b *TestBot) listenPlayer(player *Player) error {
	for {
		select {
		case <-b.done:
			return nil
		default:
		}
		
		// 设置较短的读取超时，让 ping handler 有机会执行
		player.WS.SetReadDeadline(time.Now().Add(60 * time.Second))
		
		_, message, err := player.WS.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				return nil
			}
			// 检查是否已完成
			b.mu.Lock()
			played := b.roundsPlayed
			b.mu.Unlock()
			if played >= *rounds {
				return nil
			}
			// 如果是网络错误，记录并返回
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("  玩家 %s WebSocket 意外关闭: %v", player.Username, err)
			}
			// 检查是否是 "use of closed network connection" 错误
			if err.Error() == "use of closed network connection" {
				log.Printf("  玩家 %s 连接已关闭", player.Username)
				return nil
			}
			return err
		}
		
		var msg WSMessage
		if err := json.Unmarshal(message, &msg); err != nil {
			continue
		}
		
		switch msg.Type {
		case "round_result":
			var result WSRoundResult
			if err := json.Unmarshal(msg.Payload, &result); err != nil {
				continue
			}
			
			b.mu.Lock()
			// 检查是否已处理过这个轮次（避免多个玩家重复计数）
			if !b.seenRoundIDs[result.RoundID] {
				b.seenRoundIDs[result.RoundID] = true
				b.roundsPlayed++
				b.roundResults = append(b.roundResults, result)
				
				// 打印结果
				winnerStr := ""
				for i, name := range result.WinnerNames {
					if i > 0 {
						winnerStr += ", "
					}
					winnerStr += name
				}
				log.Printf("  第 %d 轮结束 - 赢家: %s, 奖金: %s", b.roundsPlayed, winnerStr, result.PrizePerWinner)
			}
			b.mu.Unlock()
			
		case "balance_update":
			var update struct {
				Balance       string `json:"balance"`
				FrozenBalance string `json:"frozen_balance"`
			}
			if err := json.Unmarshal(msg.Payload, &update); err != nil {
				continue
			}
			
			balance, _ := decimal.NewFromString(update.Balance)
			player.Balance = balance
			
		case "error":
			var wsErr struct {
				Code    int    `json:"code"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(msg.Payload, &wsErr); err != nil {
				continue
			}
			log.Printf("  玩家 %s 收到错误: [%d] %s", player.Username, wsErr.Code, wsErr.Message)
			
		case "game_state":
			// 游戏状态更新
			if player == b.players[0] {
				var state struct {
					State string `json:"state"`
				}
				if err := json.Unmarshal(msg.Payload, &state); err == nil {
					log.Printf("  游戏状态: %s", state.State)
				}
			}
			
		case "phase_change":
			// 阶段变化
			if player == b.players[0] {
				var phase struct {
					Phase string `json:"phase"`
					Round int    `json:"round"`
				}
				if err := json.Unmarshal(msg.Payload, &phase); err == nil {
					log.Printf("  阶段变化: %s (轮次: %d)", phase.Phase, phase.Round)
				}
			}
			
		case "round_failed":
			// 轮次失败
			if player == b.players[0] {
				var failed struct {
					Reason string `json:"reason"`
				}
				if err := json.Unmarshal(msg.Payload, &failed); err == nil {
					log.Printf("  轮次失败: %s", failed.Reason)
				}
			}
			
		case "countdown", "phase_tick":
			// 倒计时消息，忽略
			
		case "player_ready", "player_joined", "player_left", "player_update", "room_state", "player_join", "betting_done":
			// 玩家状态消息，忽略
			
		default:
			// 打印未知消息类型（仅第一个玩家）
			if player == b.players[0] {
				log.Printf("  收到消息: %s", msg.Type)
			}
		}
	}
}

// analyzeResults 分析游戏结果
func (b *TestBot) analyzeResults() {
	log.Println("\n========================================")
	log.Println("游戏结果分析")
	log.Println("========================================")
	
	// 获取最新余额
	log.Println("\n[玩家余额]")
	initBal, _ := decimal.NewFromString(*initBalance)
	bet, _ := decimal.NewFromString(*betAmount)
	
	totalPlayerBalance := decimal.Zero
	for _, player := range b.players {
		// 获取最新余额
		meResp, err := b.get("/api/me", player.Token)
		if err != nil {
			log.Printf("  获取玩家 %s 余额失败: %v", player.Username, err)
			continue
		}
		
		var me UserInfo
		if err := json.Unmarshal(meResp, &me); err != nil {
			continue
		}
		
		balance, _ := decimal.NewFromString(me.Balance)
		diff := balance.Sub(initBal)
		sign := ""
		if diff.IsPositive() {
			sign = "+"
		}
		
		log.Printf("  %s: %s (初始: %s, 变化: %s%s)", 
			player.Username, me.Balance, *initBalance, sign, diff.String())
		
		totalPlayerBalance = totalPlayerBalance.Add(balance)
	}
	
	// 获取房主余额
	log.Println("\n[房主余额]")
	ownerResp, err := b.get("/api/me", b.owner.Token)
	if err != nil {
		log.Printf("  获取房主余额失败: %v", err)
	} else {
		var owner UserInfo
		if err := json.Unmarshal(ownerResp, &owner); err == nil {
			log.Printf("  %s: %s", b.owner.Username, owner.Balance)
		}
	}
	
	// 统计胜负
	log.Println("\n[游戏统计]")
	log.Printf("  总轮数: %d", len(b.roundResults))
	
	winCount := make(map[int64]int)
	for _, result := range b.roundResults {
		for _, winnerID := range result.Winners {
			winCount[winnerID]++
		}
	}
	
	for _, player := range b.players {
		wins := winCount[player.UserID]
		log.Printf("  %s: 获胜 %d 次", player.Username, wins)
	}
	
	// 资金分析
	log.Println("\n[资金分析]")
	totalBet := bet.Mul(decimal.NewFromInt(int64(len(b.players)))).Mul(decimal.NewFromInt(int64(len(b.roundResults))))
	log.Printf("  总下注金额: %s", totalBet.String())
	
	// 计算抽成
	ownerCommission := totalBet.Mul(decimal.NewFromFloat(0.03))
	platformCommission := totalBet.Mul(decimal.NewFromFloat(0.02))
	log.Printf("  房主抽成 (3%%): %s", ownerCommission.String())
	log.Printf("  平台抽成 (2%%): %s", platformCommission.String())
	
	// 玩家总余额变化
	totalInitBalance := initBal.Mul(decimal.NewFromInt(int64(len(b.players))))
	playerBalanceChange := totalPlayerBalance.Sub(totalInitBalance)
	log.Printf("  玩家总余额变化: %s", playerBalanceChange.String())
	
	// 验证资金守恒
	log.Println("\n[资金守恒验证]")
	expectedLoss := ownerCommission.Add(platformCommission)
	log.Printf("  预期玩家总损失 (抽成): %s", expectedLoss.Neg().String())
	log.Printf("  实际玩家总损失: %s", playerBalanceChange.String())
	
	diff := playerBalanceChange.Add(expectedLoss).Abs()
	if diff.LessThan(decimal.NewFromFloat(0.01)) {
		log.Println("  ✓ 资金守恒验证通过!")
	} else {
		log.Printf("  ✗ 资金守恒验证失败，差异: %s", diff.String())
	}
	
	log.Println("\n========================================")
	log.Println("测试完成")
	log.Println("========================================")
}


// HTTP 辅助方法

func (b *TestBot) get(path, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", *baseURL+path, nil)
	if err != nil {
		return nil, err
	}
	
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(body))
	}
	
	return body, nil
}

func (b *TestBot) post(path string, data interface{}, token string) ([]byte, error) {
	var body io.Reader
	if data != nil {
		jsonData, err := json.Marshal(data)
		if err != nil {
			return nil, err
		}
		body = bytes.NewBuffer(jsonData)
	}
	
	req, err := http.NewRequest("POST", *baseURL+path, body)
	if err != nil {
		return nil, err
	}
	
	req.Header.Set("Content-Type", "application/json")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	
	resp, err := b.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, string(respBody))
	}
	
	return respBody, nil
}

func (b *TestBot) getWSURL(token string) string {
	u, _ := url.Parse(*baseURL)
	scheme := "ws"
	if u.Scheme == "https" {
		scheme = "wss"
	}
	return fmt.Sprintf("%s://%s/ws?token=%s", scheme, u.Host, token)
}
