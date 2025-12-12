# Implementation Plan

## Phase 1: 数据库扩展与基础设施

- [x] 1. 创建数据库迁移脚本
  - [x] 1.1 创建聊天消息表 (chat_messages)
    - 创建表结构和索引
    - _Requirements: 2.1, 2.2_
  - [x] 1.2 创建好友相关表 (friends, friend_requests)
    - 创建好友关系表和好友请求表
    - _Requirements: 6.1, 6.2_
  - [x] 1.3 创建邀请相关表 (room_invitations, invite_links)
    - 创建房间邀请表和邀请链接表
    - _Requirements: 7.1, 7.5_
  - [x] 1.4 创建风控和监控表 (risk_flags, alerts, metrics_snapshots)
    - 创建风控标记、告警记录、指标快照表
    - _Requirements: 12.1, 13.1, 9.1_
  - [x] 1.5 创建房间主题表 (room_themes)
    - 创建主题配置表
    - _Requirements: 14.1_
  - [x] 1.6 扩展用户表字段
    - 添加 device_fingerprint, balance_version, language, consecutive_wins 字段
    - _Requirements: 10.6, 8.3, 12.1_

- [x] 2. Checkpoint - 确保数据库迁移成功
  - Ensure all tests pass, ask the user if questions arise.

## Phase 2: 观战模式实现

- [x] 3. 实现观战模式后端




  - [x] 3.1 创建 SpectatorManager 组件

    - 实现观战者状态管理
    - _Requirements: 1.1, 1.5_

  - [x] 3.2 扩展 RoomProcessor 支持观战者

    - 添加 AddSpectator, RemoveSpectator, SpectatorToParticipant 方法
    - _Requirements: 1.1, 1.4_
  - [x]* 3.3 编写观战者隔离属性测试


    - **Property 1: Spectator isolation**
    - **Validates: Requirements 1.3**

  - [x] 3.4 实现观战者 WebSocket 事件



    - 添加 spectator_join, spectator_leave, spectator_switch 事件处理

    - _Requirements: 1.1, 1.2_
  - [x]* 3.5 编写观战者接收更新属性测试
    - **Property 2: Spectator receives all updates**
    - **Validates: Requirements 1.2**


  - [x] 3.6 实现观战者 REST API
    - POST /api/rooms/:id/spectate, POST /api/rooms/:id/switch-to-participant
    - _Requirements: 1.1, 1.4_

- [x] 4. 实现观战模式前端 (Flutter App)


  - [x] 4.1 更新房间页面支持观战模式


    - 添加观战者列表显示和观战/参与切换按钮
    - _Requirements: 1.6_

  - [x] 4.2 实现观战者 UI 限制

    - 禁用自动准备开关，显示观战状态标识
    - _Requirements: 1.3_


- [x] 5. Checkpoint - 确保观战模式功能正常


  - Ensure all tests pass, ask the user if questions arise.

## Phase 3: 聊天与表情功能

- [x] 6. 实现聊天服务后端


  - [x] 6.1 创建 ChatService 和 ChatRepo

    - 实现消息存储和查询
    - _Requirements: 2.1, 2.2_

  - [x] 6.2 实现内容过滤器 (ContentFilter)
    - 实现敏感词过滤和替换

    - _Requirements: 2.5_
  - [x] 6.3 实现聊天限流器

    - 使用 Redis 实现每秒1条消息限制
    - _Requirements: 2.4_
  - [x]* 6.4 编写聊天消息截断属性测试
    - **Property 4: Chat message truncation**
    - **Validates: Requirements 2.3**
  - [x]* 6.5 编写聊天限流属性测试

    - **Property 5: Chat rate limiting**
    - **Validates: Requirements 2.4**

  - [x] 6.6 实现聊天 WebSocket 事件
    - 添加 chat_message, chat_history 事件处理
    - _Requirements: 2.1_
  - [x]* 6.7 编写聊天广播完整性属性测试
    - **Property 3: Chat message broadcast completeness**
    - **Validates: Requirements 2.1**

- [x] 7. 实现表情反应后端



  - [x] 7.1 实现表情限流器
    - 使用 Redis 实现每秒3次限制
    - _Requirements: 3.3_
  - [x]* 7.2 编写表情限流属性测试
    - **Property 6: Emoji rate limiting**

    - **Validates: Requirements 3.3**
  - [x] 7.3 实现表情 WebSocket 事件
    - 添加 emoji_reaction 事件处理和广播
    - _Requirements: 3.1_

- [x] 8. 实现聊天与表情前端 (Flutter App)



  - [x] 8.1 创建 ChatWidget 组件

    - 实现聊天消息列表和输入框
    - _Requirements: 2.2, 2.6_

  - [x] 8.2 创建 EmojiPicker 组件


    - 实现12个预定义表情选择器

    - _Requirements: 3.2_

  - [x] 8.3 集成聊天和表情到房间页面

    - 添加聊天面板和表情按钮
    - _Requirements: 2.1, 3.1_

- [x] 9. Checkpoint - 确保聊天和表情功能正常

  - Ensure all tests pass, ask the user if questions arise.

## Phase 4: 游戏记录与回放

- [x] 10. 实现游戏记录后端




  - [x] 10.1 创建 GameHistoryService 和扩展 GameRepo


    - 实现游戏记录查询和统计
    - _Requirements: 4.1, 4.5_
  - [x]* 10.2 编写游戏记录分页属性测试
    - **Property 7: Game history pagination**
    - **Validates: Requirements 4.1**
  - [x]* 10.3 编写游戏记录日期过滤属性测试
    - **Property 8: Game history date filtering**
    - **Validates: Requirements 4.3**



  - [x] 10.4 实现游戏记录 REST API
    - GET /api/game-history, GET /api/game-history/:id, GET /api/game-stats
    - _Requirements: 4.1, 4.6, 4.5_


- [x] 11. 实现游戏回放后端

  - [x] 11.1 实现回合验证功能

    - 使用 reveal_seed 重新计算赢家验证
    - _Requirements: 5.4_
  - [x]* 11.2 编写回合验证一致性属性测试
    - **Property 9: Round verification consistency**
    - **Validates: Requirements 5.4**


  - [x] 11.3 实现回放数据 REST API

    - GET /api/game-rounds/:id/replay
    - _Requirements: 5.1, 5.3_

- [x] 12. 实现游戏记录前端 (Flutter App)



  - [x] 12.1 创建 GameHistoryPage 页面
    - 实现游戏记录列表和筛选
    - _Requirements: 4.1, 4.3, 4.4_
  - [x] 12.2 创建 GameReplayPage 页面

    - 实现回放动画和验证功能
    - _Requirements: 5.1, 5.4_

  - [x] 12.3 创建 GameStatsWidget 组件

    - 显示胜率、总投注、净盈亏等统计
    - _Requirements: 4.5_


- [x] 13. Checkpoint - 确保游戏记录和回放功能正常

  - Ensure all tests pass, ask the user if questions arise.

## Phase 5: 好友与邀请系统


- [x] 14. 实现好友服务后端




  - [x] 14.1 创建 FriendService 和 FriendRepo

    - 实现好友请求、接受、拒绝、删除
    - _Requirements: 6.1, 6.2, 6.3, 6.5_
  - [x]* 14.2 编写好友关系双向性属性测试
    - **Property 10: Friend relationship bidirectionality**
    - **Validates: Requirements 6.2**
  - [x]* 14.3 编写好友删除双向性属性测试
    - **Property 11: Friend removal bidirectionality**
    - **Validates: Requirements 6.5**

  - [x] 14.4 实现好友在线状态追踪

    - 使用 Redis 存储在线状态
    - _Requirements: 6.4_

  - [x] 14.5 实现好友 WebSocket 事件

    - 添加 friend_request, friend_accepted, friend_online, friend_offline 事件
    - _Requirements: 6.1, 6.2, 6.4_
  - [x] 14.6 实现好友 REST API

    - POST /api/friends/request, POST /api/friends/accept, DELETE /api/friends/:id

    - _Requirements: 6.1, 6.2, 6.5_


- [x] 15. 实现邀请服务后端



  - [x] 15.1 创建 InvitationService 和 InvitationRepo

    - 实现房间邀请和邀请链接
    - _Requirements: 7.1, 7.5_
  - [x]* 15.2 编写邀请通知送达属性测试
    - **Property 12: Invitation notification delivery**
    - **Validates: Requirements 7.2**
  - [x]* 15.3 编写邀请链接有效性属性测试
    - **Property 13: Invite link validity**
    - **Validates: Requirements 7.5, 7.6**


  - [x] 15.4 实现邀请 WebSocket 事件


    - 添加 room_invitation, invite_response 事件

    - _Requirements: 7.1, 7.4_
  - [x] 15.5 实现邀请 REST API
    - POST /api/rooms/:id/invite, POST /api/rooms/:id/invite-link, POST /api/invite/:code/join
    - _Requirements: 7.1, 7.5, 7.6_

- [x] 16. 实现好友与邀请前端 (Flutter App)
  - [x] 16.1 创建 FriendListPage 页面
    - 显示好友列表、在线状态、当前房间
    - _Requirements: 6.4_
  - [x] 16.2 创建 FriendRequestsPage 页面
    - 显示待处理的好友请求
    - _Requirements: 6.1, 6.2, 6.3_
  - [x] 16.3 实现邀请通知弹窗
    - 显示邀请详情和接受/拒绝按钮
    - _Requirements: 7.2, 7.3, 7.4_
  - [x] 16.4 实现邀请链接分享功能

    - 生成和分享邀请链接
    - _Requirements: 7.5_


- [x] 17. Checkpoint - 确保好友和邀请功能正常
  - Ensure all tests pass, ask the user if questions arise.

## Phase 6: 钱包功能完善

- [x] 18. 完善钱包后端
  - [x] 18.1 扩展 FundService 支持钱包功能
    - 实现余额查询、交易历史、收益统计
    - _Requirements: 15.1, 15.2, 15.5_
  - [x]* 18.2 编写钱包余额准确性属性测试
    - **Property 21: Wallet balance accuracy**
    - **Validates: Requirements 15.1**
  - [x]* 18.3 编写收益计算准确性属性测试
    - **Property 22: Earnings calculation accuracy**
    - **Validates: Requirements 15.5**
  - [x] 18.4 实现钱包 REST API
    - GET /api/wallet, GET /api/wallet/transactions, GET /api/wallet/earnings
    - _Requirements: 15.1, 15.2, 15.5_

- [x] 19. 实现钱包前端 (Flutter App)
  - [x] 19.1 创建 WalletPage 页面
    - 显示可用余额、冻结余额、总余额
    - _Requirements: 15.1_
  - [x] 19.2 创建 TransactionHistoryWidget 组件
    - 显示交易历史列表
    - _Requirements: 15.2_
  - [x] 19.3 创建 EarningsSummaryWidget 组件
    - 显示收益统计图表

    - _Requirements: 15.5_
  - [x] 19.4 实现提现申请流程

    - 提现表单和状态追踪
    - _Requirements: 15.3, 15.4_

- [x] 20. Checkpoint - 确保钱包功能正常



  - Ensure all tests pass, ask the user if questions arise.

## Phase 7: 余额缓存与性能优化

- [x] 21. 实现余额缓存机制




  - [x] 21.1 创建 BalanceCache 组件

    - 实现 Redis 缓存读写
    - _Requirements: 10.1, 10.2_
  - [x]* 21.2 编写缓存一致性属性测试
    - **Property 14: Balance cache consistency**
    - **Validates: Requirements 10.2**

  - [x] 21.3 实现乐观锁机制

    - 使用 balance_version 字段防止并发冲突
    - _Requirements: 10.6_
  - [x]* 21.4 编写乐观锁属性测试
    - **Property 15: Optimistic locking prevents conflicts**


    - **Validates: Requirements 10.6**




  - [x] 21.5 集成缓存到 RoomProcessor

    - 修改余额读写使用缓存
    - _Requirements: 10.1, 10.4, 10.5_


- [x] 22. 实现 Phase Tick 优化


  - [x] 22.1 实现增量状态广播

    - 只发送变化的字段
    - _Requirements: 11.1_

  - [x] 22.2 实现空闲跳过逻辑
    - 无状态变化时跳过广播
    - _Requirements: 11.2_
  - [x] 22.3 调整 tick 间隔
    - 活跃阶段1秒，等待阶段3秒
    - _Requirements: 11.4_

- [x] 23. Checkpoint - 确保性能优化生效

  - Ensure all tests pass, ask the user if questions arise.

## Phase 8: 风控与告警系统





- [x] 24. 实现风控检测后端
  - [x] 24.1 创建 RiskControlService 和 RiskRepo



    - 实现风控检测和标记

    - _Requirements: 12.1, 12.2, 12.3_
  - [x]* 24.2 编写连续获胜检测属性测试
    - **Property 16: Consecutive win detection**
    - **Validates: Requirements 12.1**
  - [x]* 24.3 编写胜率检测属性测试
    - **Property 17: Win rate detection**
    - **Validates: Requirements 12.2**

  - [x] 24.4 实现设备指纹检测




    - 检测多账户共用设备

    - _Requirements: 12.3_
  - [x] 24.5 集成风控检测到结算流程

    - 每轮结算后检查风控条件
    - _Requirements: 12.1, 12.2_



- [x] 25. 实现告警系统后端
  - [x] 25.1 创建 AlertManager 和 AlertRepo
    - 实现告警创建和管理
    - _Requirements: 13.1, 13.2, 13.3_
  - [x]* 25.2 编写负余额告警属性测试
    - **Property 18: Negative balance alert**
    - **Validates: Requirements 13.1**
  - [x]* 25.3 编写大额交易告警属性测试
    - **Property 19: Large transaction alert**
    - **Validates: Requirements 13.3**
  - [x] 25.4 实现告警 WebSocket 推送

    - 向管理员实时推送告警
    - _Requirements: 12.4, 13.1_


  - [x] 25.5 实现告警 REST API
    - GET /api/admin/alerts, POST /api/admin/alerts/:id/acknowledge
    - _Requirements: 12.5_

- [x] 26. 实现风控管理后台 (Flutter Admin)



  - [x] 26.1 创建 RiskFlagsPage 页面

    - 显示风控标记列表和审核功能
    - _Requirements: 12.5, 12.6_

  - [x] 26.2 创建 AlertsPage 页面
    - 显示告警列表和确认功能
    - _Requirements: 13.1_

- [x] 27. Checkpoint - 确保风控和告警功能正常


  - Ensure all tests pass, ask the user if questions arise.

## Phase 9: 监控仪表盘

- [x] 28. 实现监控服务后端
  - [x] 28.1 创建 MonitoringService 和 MetricsRepo
    - 实现指标采集和存储
    - _Requirements: 9.1, 9.2, 9.3_
  - [x] 28.2 实现实时指标采集
    - 使用 Redis 存储实时指标
    - _Requirements: 9.1_
  - [x] 28.3 实现指标快照定时任务
    - 每分钟保存指标快照
    - _Requirements: 9.5_
  - [x] 28.4 实现监控 REST API
    - GET /api/admin/metrics/realtime, GET /api/admin/metrics/history
    - _Requirements: 9.1, 9.5_
  - [x] 28.5 实现监控 WebSocket 推送
    - 每10秒推送指标更新
    - _Requirements: 9.6_

- [x] 29. 实现监控仪表盘前端 (Flutter Admin)



  - [x] 29.1 创建 MonitoringDashboardPage 页面

    - 显示实时指标卡片
    - _Requirements: 9.1, 9.2, 9.3_


  - [x] 29.2 实现指标图表组件

    - 使用 fl_chart 显示历史趋势

    - _Requirements: 9.5_
  - [x] 29.3 实现告警指示器

    - 超阈值指标红色高亮
    - _Requirements: 9.4_

  - [x] 29.4 实现自动刷新


    - 每10秒自动更新数据
    - _Requirements: 9.6_

- [x] 30. Checkpoint - 确保监控仪表盘功能正常


  - Ensure all tests pass, ask the user if questions arise.



## Phase 10: 房间主题与多语言

- [x] 31. 实现房间主题后端

  - [x] 31.1 创建 ThemeService 和 ThemeRepo

    - 实现主题配置存储
    - _Requirements: 14.1_
  - [x]* 31.2 编写主题持久化属性测试
    - **Property 20: Theme persistence**
    - **Validates: Requirements 14.3**

  - [x] 31.3 实现主题 WebSocket 事件


    - 添加 theme_change 事件

    - _Requirements: 14.5_


  - [x] 31.4 实现主题 REST API
    - PUT /api/rooms/:id/theme
    - _Requirements: 14.1_



- [x] 32. 实现房间主题前端 (Flutter App)

  - [x] 32.1 创建 ThemeSelector 组件

    - 显示5个内置主题预览
    - _Requirements: 14.2, 14.4_



  - [x] 32.2 实现主题应用逻辑
    - 根据房间主题切换颜色方案
    - _Requirements: 14.1, 14.3_



- [x] 33. 增强多语言支持

  - [x] 33.1 添加新语言资源文件

    - 添加繁体中文、日语、韩语
    - _Requirements: 8.4_


  - [x] 33.2 实现语言偏好存储

    - 保存用户语言偏好到数据库

    - _Requirements: 8.1, 8.3_

  - [x] 33.3 实现语言回退逻辑

    - 缺失资源时回退到英语

    - _Requirements: 8.5_

- [x] 34. Checkpoint - 确保主题和多语言功能正常

  - Ensure all tests pass, ask the user if questions arise.

## Phase 11: 集成测试与文档





















- [x]* 35. 编写集成测试
  - [x]* 35.1 WebSocket 消息广播集成测试


    - 测试多用户场景下的消息传递
  - [x]* 35.2 并发操作集成测试
    - 测试多用户同时操作的一致性
  - [x]* 35.3 缓存失效集成测试
    - 测试 Redis 不可用时的降级

- [x]* 36. 更新项目文档
  - [x]* 36.1 更新 API 文档
    - 添加新增 API 的说明
  - [x]* 36.2 更新技术设计文档
    - 添加 P1/P2 功能的设计说明
  - [x]* 36.3 更新 README
    - 添加新功能的使用说明

- [x] 37. Final Checkpoint - 确保所有功能正常
  - Ensure all tests pass, ask the user if questions arise.
