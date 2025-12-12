# Requirements Document

## Introduction

本文档定义了 5SecondsGo 多人在线高频押注小游戏的 P1（体验增强）和 P2（扩展优化）功能需求。这些功能将在核心 MVP 功能基础上，提升用户体验、增强社交互动、优化系统性能，并提供更完善的管理和监控能力。

## Glossary

- **System**: 5SecondsGo 游戏系统
- **Player**: 普通玩家用户
- **Owner**: 房主用户，可创建和管理房间
- **Admin**: 平台管理员
- **Room**: 游戏房间，玩家在其中进行游戏
- **Spectator**: 观战者，只能观看不能参与游戏
- **Chat Message**: 房间内的聊天消息
- **Emoji Reaction**: 表情反应，快速表达情绪
- **Game Record**: 游戏回合记录
- **Friend**: 好友关系
- **Invite Link**: 邀请链接
- **Balance Cache**: 余额缓存，用于提升读取性能
- **Risk Control**: 风控系统，检测异常行为

## Requirements

### Requirement 1: 观战模式

**User Story:** As a Player, I want to spectate a room without participating in the game, so that I can learn the game mechanics or watch friends play.

#### Acceptance Criteria

1. WHEN a Player requests to join a room as spectator THEN the System SHALL add the Player to the room's spectator list and broadcast the spectator join event to all room members
2. WHEN a Spectator is in a room THEN the System SHALL send all game state updates to the Spectator including phase changes, betting results, and settlement results
3. WHEN a Spectator is in a room THEN the System SHALL prevent the Spectator from setting auto-ready status or participating in any betting round
4. WHEN a Spectator wants to become a participant THEN the System SHALL allow the Spectator to switch to participant mode if the room is not full
5. WHEN a room reaches maximum spectator limit (50) THEN the System SHALL reject new spectator join requests with appropriate error message
6. WHEN displaying room information THEN the System SHALL show both participant count and spectator count separately

### Requirement 2: 房间聊天功能

**User Story:** As a Player, I want to send chat messages in the room, so that I can communicate with other players during the game.

#### Acceptance Criteria

1. WHEN a Player sends a chat message THEN the System SHALL broadcast the message to all room members (participants and spectators) within 100ms
2. WHEN a chat message is received THEN the System SHALL display the sender's username, message content, and timestamp
3. WHEN a Player sends a message exceeding 200 characters THEN the System SHALL truncate the message to 200 characters
4. WHEN a Player sends messages faster than 1 message per second THEN the System SHALL rate-limit and reject excess messages with a warning
5. WHEN a chat message contains prohibited content THEN the System SHALL filter the content and replace with asterisks
6. WHEN the chat history exceeds 100 messages THEN the System SHALL remove the oldest messages from the client display

### Requirement 3: 表情反应功能

**User Story:** As a Player, I want to send quick emoji reactions, so that I can express my emotions without typing.

#### Acceptance Criteria

1. WHEN a Player sends an emoji reaction THEN the System SHALL broadcast the reaction to all room members with animation display
2. WHEN displaying emoji reactions THEN the System SHALL show a predefined set of 12 emojis (happy, sad, angry, surprised, thumbs up, thumbs down, clap, fire, heart, laugh, cry, cool)
3. WHEN a Player sends emoji reactions faster than 3 per second THEN the System SHALL rate-limit and ignore excess reactions
4. WHEN an emoji reaction is displayed THEN the System SHALL show it for 3 seconds with fade-out animation
5. WHEN multiple reactions are sent simultaneously THEN the System SHALL queue and display them sequentially without overlap

### Requirement 4: 游戏记录查询

**User Story:** As a Player, I want to view my game history, so that I can track my wins, losses, and earnings over time.

#### Acceptance Criteria

1. WHEN a Player requests game history THEN the System SHALL return paginated records with 20 items per page sorted by time descending
2. WHEN displaying a game record THEN the System SHALL show room name, round number, bet amount, result (win/lose/skipped), prize amount, and timestamp
3. WHEN a Player filters game history by date range THEN the System SHALL return only records within the specified range
4. WHEN a Player filters game history by room THEN the System SHALL return only records from the specified room
5. WHEN calculating statistics THEN the System SHALL compute total rounds played, win rate, total wagered, total won, and net profit/loss
6. WHEN a Player requests detailed round information THEN the System SHALL show all participants, winners, commit hash, and reveal seed for verification

### Requirement 5: 游戏回放功能

**User Story:** As a Player, I want to replay past game rounds, so that I can verify the fairness of the game.

#### Acceptance Criteria

1. WHEN a Player requests to replay a round THEN the System SHALL retrieve the complete round data including all state transitions
2. WHEN replaying a round THEN the System SHALL display the phase transitions with original timing (5 seconds per phase)
3. WHEN replaying a round THEN the System SHALL show the commit hash at the betting phase and reveal seed at settlement phase
4. WHEN a Player verifies a round THEN the System SHALL allow the Player to independently compute winners using the revealed seed
5. WHEN replay data is older than 30 days THEN the System SHALL archive the data but still allow retrieval with longer response time

### Requirement 6: 好友系统

**User Story:** As a Player, I want to add other players as friends, so that I can easily find and join games with them.

#### Acceptance Criteria

1. WHEN a Player sends a friend request THEN the System SHALL create a pending friend request and notify the target Player
2. WHEN a Player accepts a friend request THEN the System SHALL establish a bidirectional friend relationship
3. WHEN a Player rejects a friend request THEN the System SHALL remove the pending request and optionally notify the sender
4. WHEN a Player views their friend list THEN the System SHALL show friend username, online status, and current room (if any)
5. WHEN a Player removes a friend THEN the System SHALL remove the bidirectional relationship
6. WHEN a Player has more than 200 friends THEN the System SHALL reject new friend requests until existing friends are removed

### Requirement 7: 邀请功能

**User Story:** As a Player, I want to invite friends to join my current room, so that we can play together.

#### Acceptance Criteria

1. WHEN a Player invites a friend to a room THEN the System SHALL send a real-time notification to the friend
2. WHEN a friend receives an invitation THEN the System SHALL display the room name, bet amount, and current player count
3. WHEN a friend accepts an invitation THEN the System SHALL automatically join the friend to the room if space is available
4. WHEN a friend declines an invitation THEN the System SHALL notify the inviter of the decline
5. WHEN generating a shareable invite link THEN the System SHALL create a unique link valid for 24 hours
6. WHEN a Player uses an invite link THEN the System SHALL join the Player to the room if the link is valid and room has space

### Requirement 8: 多语言支持增强

**User Story:** As a Player, I want to use the app in my preferred language, so that I can understand all features easily.

#### Acceptance Criteria

1. WHEN a Player changes language preference THEN the System SHALL immediately update all UI text to the selected language
2. WHEN displaying system messages THEN the System SHALL use the Player's preferred language
3. WHEN a new Player registers THEN the System SHALL default to the device's system language if supported
4. THE System SHALL support Chinese (Simplified), Chinese (Traditional), English, Japanese, and Korean languages
5. WHEN language resources are missing THEN the System SHALL fall back to English

### Requirement 9: 平台监控仪表盘

**User Story:** As an Admin, I want to view real-time platform metrics, so that I can monitor system health and business performance.

#### Acceptance Criteria

1. WHEN an Admin views the dashboard THEN the System SHALL display real-time metrics including online players, active rooms, and games per minute
2. WHEN displaying performance metrics THEN the System SHALL show API response time P95, WebSocket message latency P95, and database query time P95
3. WHEN displaying business metrics THEN the System SHALL show daily active users, daily transaction volume, and platform revenue
4. WHEN a metric exceeds threshold THEN the System SHALL highlight the metric in red and show alert indicator
5. WHEN viewing historical data THEN the System SHALL allow selection of time ranges (1 hour, 24 hours, 7 days, 30 days)
6. THE System SHALL refresh dashboard metrics every 10 seconds automatically

### Requirement 10: 余额缓存机制

**User Story:** As a System, I want to cache player balances, so that I can reduce database load and improve response time.

#### Acceptance Criteria

1. WHEN a Player's balance is queried within a room THEN the System SHALL read from in-memory cache instead of database
2. WHEN a balance-changing transaction completes THEN the System SHALL update both database and cache atomically
3. WHEN cache and database are inconsistent THEN the System SHALL use database as source of truth and refresh cache
4. WHEN a Player joins a room THEN the System SHALL load their balance from database into room cache
5. WHEN a Player leaves a room THEN the System SHALL remove their balance from room cache
6. THE System SHALL implement optimistic locking using balance_version field to prevent concurrent update conflicts

### Requirement 11: 性能优化 - Phase Tick

**User Story:** As a System, I want to optimize phase tick broadcasts, so that I can reduce network bandwidth and improve scalability.

#### Acceptance Criteria

1. WHEN broadcasting phase tick THEN the System SHALL send only changed fields instead of full state
2. WHEN a room has no state changes THEN the System SHALL skip the phase tick broadcast for that interval
3. WHEN a Player reconnects THEN the System SHALL send full state snapshot regardless of tick optimization
4. THE System SHALL maintain phase tick interval at 1 second for active game phases and 3 seconds for waiting phase
5. WHEN calculating time remaining THEN the System SHALL use server timestamp to allow client-side interpolation

### Requirement 12: 反作弊基础检测

**User Story:** As a System, I want to detect suspicious player behavior, so that I can maintain game fairness.

#### Acceptance Criteria

1. WHEN a Player wins more than 10 consecutive rounds THEN the System SHALL flag the account for review
2. WHEN a Player's win rate exceeds 80% over 50 rounds THEN the System SHALL flag the account for review
3. WHEN multiple accounts share the same device fingerprint THEN the System SHALL flag all accounts for review
4. WHEN a flagged account is detected THEN the System SHALL notify Admin via dashboard alert
5. WHEN an Admin reviews a flagged account THEN the System SHALL provide detailed statistics and activity logs
6. WHEN an Admin confirms cheating THEN the System SHALL allow account suspension and balance freezing

### Requirement 13: 风控告警系统

**User Story:** As an Admin, I want to receive alerts for abnormal activities, so that I can respond to issues quickly.

#### Acceptance Criteria

1. WHEN player balance becomes negative THEN the System SHALL trigger immediate alert to Admin
2. WHEN owner custody quota becomes negative THEN the System SHALL trigger immediate alert and lock owner's rooms
3. WHEN single transaction exceeds 10000 THEN the System SHALL trigger alert for manual review
4. WHEN daily transaction volume for a user exceeds 100000 THEN the System SHALL trigger alert
5. WHEN settlement transaction fails 3 consecutive times THEN the System SHALL trigger critical alert
6. WHEN fund conservation check fails THEN the System SHALL trigger critical alert with detailed discrepancy report

### Requirement 14: 房间主题皮肤

**User Story:** As an Owner, I want to customize my room's appearance, so that I can create a unique gaming experience.

#### Acceptance Criteria

1. WHEN an Owner selects a theme THEN the System SHALL apply the theme's color scheme and background to the room
2. THE System SHALL provide 5 built-in themes: Classic, Neon, Ocean, Forest, and Luxury
3. WHEN a Player joins a themed room THEN the System SHALL load and display the room's theme
4. WHEN displaying theme selection THEN the System SHALL show preview of each theme
5. WHEN an Owner changes theme THEN the System SHALL broadcast theme change to all room members immediately

### Requirement 15: 钱包功能完善

**User Story:** As a Player, I want a complete wallet interface, so that I can manage my funds easily.

#### Acceptance Criteria

1. WHEN a Player views wallet THEN the System SHALL display available balance, frozen balance, and total balance
2. WHEN a Player views transaction history THEN the System SHALL show all balance changes with type, amount, and timestamp
3. WHEN a Player initiates withdrawal THEN the System SHALL validate balance sufficiency and create withdrawal request
4. WHEN displaying withdrawal status THEN the System SHALL show pending, approved, rejected, or completed status
5. WHEN a Player views earnings summary THEN the System SHALL show total winnings, total losses, and net profit by time period
