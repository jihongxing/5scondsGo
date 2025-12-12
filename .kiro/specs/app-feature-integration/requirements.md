# Requirements Document

## Introduction

本文档定义了 5SecondsGo App 端功能集成的需求规范。目标是将后端已实现但前端未集成的功能完整接入，并补充缺失的核心页面（个人中心、排行榜等），确保 App 端功能完整性和用户体验一致性。

## Glossary

- **App**: 5SecondsGo Flutter 移动客户端应用
- **GlassCard**: 应用统一的毛玻璃风格卡片组件
- **GradientButton**: 应用统一的渐变按钮组件
- **GoRouter**: Flutter 路由管理库
- **Provider**: 状态管理方案 (Riverpod)
- **WebSocket**: 实时通信协议，用于游戏状态同步

## Requirements

### Requirement 1: 路由配置完善

**User Story:** As a player, I want to navigate to all app features through proper routes, so that I can access friends, profile and other features seamlessly.

#### Acceptance Criteria

1. WHEN a user taps the friends quick action on home page, THEN the App SHALL navigate to the friend list page at route `/friends`
2. WHEN a user taps the friend requests icon on friend list page, THEN the App SHALL navigate to the friend requests page at route `/friend-requests`
3. WHEN a user taps the profile/settings entry, THEN the App SHALL navigate to the profile page at route `/profile`
4. WHEN a user taps the leaderboard quick action on home page, THEN the App SHALL navigate to the leaderboard page at route `/leaderboard`
5. WHEN a user accesses an undefined route, THEN the App SHALL display an error page with navigation back to home

### Requirement 2: 个人中心页面

**User Story:** As a player, I want to view and manage my profile information, so that I can customize my experience and manage my account.

#### Acceptance Criteria

1. WHEN a user opens the profile page, THEN the App SHALL display the user's username, role badge, and avatar
2. WHEN a user opens the profile page, THEN the App SHALL display the user's wallet balance summary (available, frozen, total)
3. WHEN a user taps the language setting, THEN the App SHALL display language options and allow switching between zh/en
4. WHEN a user changes language preference, THEN the App SHALL persist the preference and update UI immediately
5. WHEN a user taps logout button, THEN the App SHALL clear authentication state and navigate to login page

### Requirement 3: 好友功能完善

**User Story:** As a player, I want to manage my friends and invite them to game rooms, so that I can play with people I know.

#### Acceptance Criteria

1. WHEN a user views the friend list, THEN the App SHALL display each friend's online status and current room (if any) using the unified GlassCard style
2. WHEN a user taps "invite to room" on a friend, THEN the App SHALL send a room invitation via the API and display confirmation
3. WHEN a user receives a room invitation, THEN the App SHALL display a notification with accept/decline options
4. WHEN a user accepts a room invitation, THEN the App SHALL navigate the user to the invited room
5. WHEN a user searches for friends by username, THEN the App SHALL query the API and display matching users

### Requirement 4: 排行榜功能

**User Story:** As a player, I want to view leaderboards, so that I can see top players and my ranking.

#### Acceptance Criteria

1. WHEN a user opens the leaderboard page, THEN the App SHALL display a tab bar with ranking categories (win rate, total wins, profit)
2. WHEN a user selects a ranking category, THEN the App SHALL fetch and display the top 100 players for that category
3. WHEN displaying leaderboard entries, THEN the App SHALL show rank number, username, avatar, and the ranking metric value
4. WHEN the current user appears in the leaderboard, THEN the App SHALL highlight their entry with a distinct visual style
5. WHEN a user pulls to refresh the leaderboard, THEN the App SHALL fetch the latest ranking data from the server

### Requirement 5: 观战功能集成

**User Story:** As a player, I want to spectate ongoing games, so that I can watch friends play without participating.

#### Acceptance Criteria

1. WHEN a user views a room card on home page, THEN the App SHALL display a spectate button alongside the join button
2. WHEN a user taps spectate on a room, THEN the App SHALL call the spectate API and enter the room in spectator mode
3. WHILE a user is spectating, THEN the App SHALL display game state without betting controls
4. WHILE a user is spectating, THEN the App SHALL display a "switch to participant" button if room has available slots
5. WHEN a spectator taps "switch to participant", THEN the App SHALL call the switch API and enable betting controls

### Requirement 6: 房间邀请链接功能

**User Story:** As a room owner, I want to generate invite links, so that I can share my room with others easily.

#### Acceptance Criteria

1. WHEN a room owner taps "generate invite link" in room settings, THEN the App SHALL call the API and display a shareable link
2. WHEN an invite link is generated, THEN the App SHALL provide copy-to-clipboard and share functionality
3. WHEN a user opens an invite link, THEN the App SHALL validate the link and navigate to the room join flow
4. IF an invite link is expired or invalid, THEN the App SHALL display an appropriate error message

### Requirement 7: 主题功能集成

**User Story:** As a room owner, I want to customize my room's visual theme, so that I can create a unique atmosphere.

#### Acceptance Criteria

1. WHEN a room owner opens room settings, THEN the App SHALL display available themes from the API
2. WHEN a room owner selects a theme, THEN the App SHALL call the update theme API and apply the theme immediately
3. WHEN a player enters a themed room, THEN the App SHALL apply the room's theme colors to the game UI
4. WHEN the room theme changes via WebSocket event, THEN the App SHALL update the UI theme in real-time

### Requirement 8: UI 风格统一

**User Story:** As a player, I want a consistent visual experience across all pages, so that the app feels polished and professional.

#### Acceptance Criteria

1. WHEN displaying any list page (friends, leaderboard, history), THEN the App SHALL use GlassCard components for list items
2. WHEN displaying any action button, THEN the App SHALL use GradientButton with appropriate gradient styles
3. WHEN displaying any page, THEN the App SHALL use GradientBackground as the root container
4. WHEN displaying loading states, THEN the App SHALL use consistent shimmer/skeleton components
5. WHEN displaying error states, THEN the App SHALL use consistent error card styling with retry option
