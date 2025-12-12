# Implementation Plan

- [x] 1. 路由配置完善



  - [x] 1.1 更新 app_router.dart 添加缺失路由

    - 添加 `/friends` 路由指向 FriendListPage
    - 添加 `/friend-requests` 路由指向 FriendRequestsPage
    - 添加 `/profile` 路由指向 ProfilePage
    - 添加 `/leaderboard` 路由指向 LeaderboardPage
    - 添加 `/invite/:code` 路由指向 InviteLinkPage
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [x] 1.2 编写路由属性测试

    - **Property 1: Invalid route handling**
    - **Validates: Requirements 1.5**

- [x] 2. 个人中心页面实现


  - [x] 2.1 创建 profile 模块目录结构


    - 创建 `features/profile/presentation/pages/` 目录
    - 创建 `features/profile/providers/` 目录
    - _Requirements: 2.1_
  - [x] 2.2 实现 ProfilePage 页面

    - 实现用户信息卡片（用户名、角色徽章、头像）
    - 实现钱包摘要卡片（可用余额、冻结余额、总余额）
    - 实现设置列表（语言切换）
    - 实现登出按钮
    - 使用 GlassCard、GradientButton 等统一组件
    - _Requirements: 2.1, 2.2, 2.3, 2.5, 8.1, 8.2, 8.3_

  - [x] 2.3 实现 ProfileProvider 状态管理

    - 实现语言切换功能
    - 实现登出功能
    - _Requirements: 2.4, 2.5_
  - [x] 2.4 编写语言偏好持久化属性测试

    - **Property 2: Language preference persistence**
    - **Validates: Requirements 2.4**


- [x] 3. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [x] 4. 排行榜功能实现
  - [x] 4.1 在 api_client.dart 添加排行榜 API 方法
    - 添加 `getLeaderboard(String type)` 方法
    - 支持 winRate、totalWins、profit 三种类型
    - _Requirements: 4.2_
  - [x] 4.2 创建 leaderboard 模块目录结构
    - 创建 `features/leaderboard/presentation/pages/` 目录
    - 创建 `features/leaderboard/providers/` 目录
    - _Requirements: 4.1_
  - [x] 4.3 实现 LeaderboardEntry 数据模型
    - 包含 userId、username、value、rank、isCurrentUser 字段
    - 实现 fromJson 工厂方法
    - _Requirements: 4.3_
  - [x] 4.4 实现 LeaderboardProvider 状态管理
    - 使用 FutureProvider.family 按类型获取排行榜
    - _Requirements: 4.2_
  - [x] 4.5 实现 LeaderboardPage 页面
    - 实现 TabBar 切换排行榜类型
    - 实现排行榜列表（使用 GlassCard）
    - 实现当前用户高亮显示
    - 实现下拉刷新
    - _Requirements: 4.1, 4.2, 4.3, 4.4, 4.5, 8.1_

  - [x] 4.6 编写排行榜属性测试
    - **Property 4: Leaderboard entry completeness**
    - **Property 5: Current user leaderboard highlighting**
    - **Validates: Requirements 4.3, 4.4**

- [x] 5. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [x] 6. 好友功能完善
  - [x] 6.1 重构 FriendListPage UI 风格
    - 使用 GradientBackground 作为页面背景
    - 使用 GlassCard 替换普通 ListTile
    - 使用 GradientButton 替换普通按钮
    - _Requirements: 3.1, 8.1, 8.2, 8.3_
  - [x] 6.2 重构 FriendRequestsPage UI 风格
    - 统一使用 GlassCard、GradientButton 组件
    - _Requirements: 8.1, 8.2, 8.3_
  - [x] 6.3 实现好友搜索功能
    - 在 api_client.dart 添加搜索用户 API（如后端支持）
    - 在 FriendListPage 添加搜索框
    - _Requirements: 3.5_
  - [x] 6.4 实现邀请好友加入房间功能
    - 创建 InviteFriendDialog 组件
    - 调用 sendRoomInvitation API
    - 显示邀请发送确认
    - _Requirements: 3.2_
  - [x] 6.5 实现房间邀请通知处理
    - 监听 WebSocket room_invitation 事件
    - 创建 InvitationReceivedDialog 组件
    - 实现接受/拒绝邀请逻辑
    - _Requirements: 3.3, 3.4_
  - [x] 6.6 编写好友搜索属性测试
    - **Property 3: Friend search results relevance**
    - **Validates: Requirements 3.5**


- [x] 7. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [x] 8. 观战功能集成

  - [x] 8.1 在 home_page.dart 房间卡片添加观战按钮
    - 在加入按钮旁添加观战按钮
    - 调用 spectate API
    - _Requirements: 5.1, 5.2_
  - [x] 8.2 创建 SpectatorControls 组件
    - 显示观战模式提示
    - 根据房间空位显示/隐藏切换按钮
    - _Requirements: 5.3, 5.4_
  - [x] 8.3 更新 RoomPage 支持观战模式
    - 根据用户角色（spectator/participant）显示不同 UI
    - 观战模式隐藏下注控件
    - 实现切换为参与者功能
    - _Requirements: 5.3, 5.4, 5.5_
  - [x] 8.4 编写观战模式属性测试
    - **Property 6: Spectator mode UI constraints**
    - **Property 7: Spectator switch button visibility**
    - **Validates: Requirements 5.3, 5.4**

- [x] 9. 房间邀请链接功能
  - [x] 9.1 创建邀请链接生成 UI
    - 在房间设置中添加生成链接按钮
    - 调用 createInviteLink API
    - 显示链接并提供复制/分享功能
    - _Requirements: 6.1, 6.2_
  - [x] 9.2 创建 InviteLinkPage 处理深度链接
    - 验证邀请码有效性
    - 显示房间信息
    - 提供加入房间按钮
    - 处理过期/无效链接错误
    - _Requirements: 6.3, 6.4_

- [x] 10. Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.

- [x] 11. 主题功能集成
  - [x] 11.1 创建 ThemeSelector 组件
    - 获取并显示可用主题列表
    - 实现主题选择和预览
    - 调用 updateRoomTheme API
    - _Requirements: 7.1, 7.2_
  - [x] 11.2 实现房间主题应用
    - 在 RoomPage 加载时获取房间主题
    - 根据主题配置应用颜色
    - _Requirements: 7.3_
  - [x] 11.3 实现主题实时更新
    - 监听 WebSocket theme_change 事件
    - 实时更新房间 UI 主题
    - _Requirements: 7.4_
  - [x] 11.4 编写主题应用属性测试
    - **Property 8: Theme color application**
    - **Validates: Requirements 7.3**

- [x] 12. UI 风格统一检查
  - [x] 12.1 检查并更新 home_page.dart
    - 确保快捷操作按钮正确导航
    - 更新排行榜按钮 onTap 回调
    - _Requirements: 1.4, 8.1, 8.2, 8.3_
  - [x] 12.2 检查并更新所有页面加载/错误状态
    - 统一使用 shimmer 加载组件
    - 统一使用错误卡片样式
    - _Requirements: 8.4, 8.5_
  - [x] 12.3 编写 UI 一致性属性测试
    - **Property 9: UI component consistency**
    - **Validates: Requirements 8.1, 8.2, 8.3, 8.4, 8.5**

- [x] 13. Final Checkpoint - 确保所有测试通过
  - Ensure all tests pass, ask the user if questions arise.
