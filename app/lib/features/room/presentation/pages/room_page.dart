import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../l10n/app_localizations.dart';

import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/ws_client.dart';
import '../../../../core/services/audio_service.dart';
import '../../../auth/providers/auth_provider.dart';
import '../widgets/chat_widget.dart';
import '../widgets/emoji_picker.dart';
import '../widgets/theme_selector.dart';
import '../widgets/win_celebration_overlay.dart';

/// 主题配置类
class ThemeConfig {
  final String name;
  final Color primaryColor;
  final Color secondaryColor;
  final Color backgroundColor;
  final Color textColor;

  const ThemeConfig({
    required this.name,
    required this.primaryColor,
    required this.secondaryColor,
    required this.backgroundColor,
    required this.textColor,
  });
}

/// 获取主题配置
ThemeConfig getThemeConfig(String themeName) {
  switch (themeName) {
    case 'neon':
      return const ThemeConfig(
        name: 'neon',
        primaryColor: Color(0xFF38F9D7),
        secondaryColor: Color(0xFFFA709A),
        backgroundColor: Color(0xFF0F0F23),
        textColor: Color(0xFFF8FAFC),
      );
    case 'ocean':
      return const ThemeConfig(
        name: 'ocean',
        primaryColor: Color(0xFF3B82F6),
        secondaryColor: Color(0xFF06B6D4),
        backgroundColor: Color(0xFF0C4A6E),
        textColor: Color(0xFFF0F9FF),
      );
    case 'forest':
      return const ThemeConfig(
        name: 'forest',
        primaryColor: Color(0xFF10B981),
        secondaryColor: Color(0xFF84CC16),
        backgroundColor: Color(0xFF14532D),
        textColor: Color(0xFFF0FDF4),
      );
    case 'luxury':
      return const ThemeConfig(
        name: 'luxury',
        primaryColor: Color(0xFFF59E0B),
        secondaryColor: Color(0xFFD97706),
        backgroundColor: Color(0xFF1C1917),
        textColor: Color(0xFFFEF3C7),
      );
    case 'classic':
    default:
      return const ThemeConfig(
        name: 'classic',
        primaryColor: Color(0xFF667EEA),
        secondaryColor: Color(0xFF764BA2),
        backgroundColor: Color(0xFF0F0F23),
        textColor: Color(0xFFF8FAFC),
      );
  }
}

class RoomPage extends ConsumerStatefulWidget {
  final int roomId;

  const RoomPage({super.key, required this.roomId});

  @override
  ConsumerState<RoomPage> createState() => _RoomPageState();
}

class _RoundEvent {
  final String type;
  final String title;
  final String subtitle;
  final DateTime time;

  const _RoundEvent({
    required this.type,
    required this.title,
    required this.subtitle,
    required this.time,
  });
}

class _RoomPageState extends ConsumerState<RoomPage> {
  String _roomName = '';
  String _phase = GamePhase.waiting;
  int _timeLeft = 5;
  int _round = 0;
  bool _autoReady = false;
  String _poolAmount = '0';
  String _betAmount = '0';
  bool _isSpectator = false;
  int _maxSpectators = 50;
  String _currentTheme = 'classic';
  ThemeConfig? _themeConfig;
  int _winnerCount = 1;
  int _maxPlayers = 10;

  int? _phaseEndTimeMs;
  Timer? _timer;
  StreamSubscription<Map<String, dynamic>>? _subscription;

  final List<_PlayerView> _players = [];
  final List<_SpectatorView> _spectators = [];
  final List<_RoundEvent> _events = [];
  final List<ChatMessage> _chatMessages = [];
  final List<_EmojiReaction> _emojiReactions = [];
  bool _showChat = false;
  bool _showEmojiPicker = false;
  
  // 获胜提示相关
  bool _showWinCelebration = false;
  String _winPrize = '0';
  bool _showOtherWinBanner = false;
  List<String> _otherWinnerNames = [];
  String _otherWinPrize = '0';

  @override
  void initState() {
    super.initState();
    _initAudio();
    _connectAndJoin();
  }

  Future<void> _initAudio() async {
    await ref.read(audioServiceProvider).init();
  }

  Future<void> _connectAndJoin() async {
    final wsClient = ref.read(wsClientProvider);
    try {
      await wsClient.connect();
      wsClient.joinRoom(widget.roomId);
      _subscription = wsClient.messages.listen(
        _handleMessage,
        onError: (error) {
          if (mounted) {
            final l10n = AppLocalizations.of(context);
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text('${l10n?.connectionError ?? 'Connection error'}: $error')),
            );
          }
        },
        onDone: () {
          // 连接关闭，WebSocket 客户端会自动重连
          // 重连成功后会自动重新加入房间
          if (mounted) {
            final l10n = AppLocalizations.of(context);
            ScaffoldMessenger.of(context).showSnackBar(
              SnackBar(content: Text(l10n?.connectionClosed ?? 'Connection closed, reconnecting...')),
            );
          }
        },
      );
    } catch (e) {
      if (mounted) {
        final l10n = AppLocalizations.of(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('${l10n?.joinRoomFailed ?? 'Failed to join room'}: $e')),
        );
      }
    }
  }

  @override
  void dispose() {
    _timer?.cancel();
    _subscription?.cancel();
    final wsClient = ref.read(wsClientProvider);
    wsClient.leaveRoom();
    super.dispose();
  }

  void _handleMessage(Map<String, dynamic> message) {
    final type = message['type'] as String?;
    final payload = message['payload'];
    if (type == null || payload == null) return;

    if (payload is! Map<String, dynamic>) {
      return;
    }

    switch (type) {
      case WSMessageType.roomState:
        _handleRoomState(payload);
        break;
      case WSMessageType.phaseChange:
        _handlePhaseChange(payload);
        break;
      case WSMessageType.phaseTick:
        _handlePhaseTick(payload);
        break;
      case WSMessageType.playerJoin:
        _handlePlayerJoin(payload);
        break;
      case WSMessageType.playerLeave:
        _handlePlayerLeave(payload);
        break;
      case WSMessageType.playerUpdate:
        _handlePlayerUpdate(payload);
        break;
      case WSMessageType.bettingDone:
        _handleBettingDone(payload);
        break;
      case WSMessageType.roundResult:
        _handleRoundResult(payload);
        break;
      case WSMessageType.roundFailed:
        _handleRoundFailed(payload);
        break;
      case WSMessageType.playerDisqualified:
        _handlePlayerDisqualified(payload);
        break;
      case WSMessageType.roundCancelled:
        _handleRoundCancelled(payload);
        break;
      case WSMessageType.balanceUpdate:
        // 余额更新暂时不单独展示，可在钱包页使用
        break;
      case WSMessageType.spectatorJoin:
        _handleSpectatorJoin(payload);
        break;
      case WSMessageType.spectatorLeave:
        _handleSpectatorLeave(payload);
        break;
      case WSMessageType.spectatorSwitch:
        _handleSpectatorSwitch(payload);
        break;
      case WSMessageType.chatMessage:
        _handleChatMessage(payload);
        break;
      case WSMessageType.emojiReaction:
        _handleEmojiReaction(payload);
        break;
      case WSMessageType.themeChange:
        _handleThemeChange(payload);
        break;
      default:
        break;
    }
  }

  void _handleRoomState(Map<String, dynamic> payload) {
    final phase = payload['phase'] as String? ?? _phase;
    final phaseEndTime = payload['phase_end_time'] as int?;
    final currentRound = payload['current_round'] as int? ?? 0;
    final poolAmount = payload['pool_amount']?.toString() ?? '0';
    final betAmount = payload['bet_amount']?.toString() ?? '0';
    final roomName = payload['room_name'] as String? ?? 'Room ${widget.roomId}';

    final playersMap = payload['players'];
    final List<_PlayerView> players = [];
    if (playersMap is Map) {
      for (final value in playersMap.values) {
        if (value is Map) {
          final data = Map<String, dynamic>.from(value);
          players.add(
            _PlayerView(
              userId: data['user_id'] as int? ?? 0,
              username: data['username'] as String? ?? 'Player',
              autoReady: data['auto_ready'] as bool? ?? false,
              isOnline: data['is_online'] as bool? ?? false,
              disqualified: data['disqualified'] as bool? ?? false,
              disqualifyReason: data['disqualify_reason'] as String? ?? '',
            ),
          );
        }
      }
    }

    players.sort((a, b) => a.userId.compareTo(b.userId));

    // 解析观战者列表
    final spectatorsMap = payload['spectators'];
    final List<_SpectatorView> spectators = [];
    if (spectatorsMap is Map) {
      for (final value in spectatorsMap.values) {
        if (value is Map) {
          final data = Map<String, dynamic>.from(value);
          spectators.add(
            _SpectatorView(
              userId: data['user_id'] as int? ?? 0,
              username: data['username'] as String? ?? 'Spectator',
            ),
          );
        }
      }
    }
    spectators.sort((a, b) => a.userId.compareTo(b.userId));

    final isSpectator = payload['is_spectator'] as bool? ?? false;
    final maxSpectators = payload['max_spectators'] as int? ?? 50;
    final themeName = payload['theme_name'] as String? ?? 'classic';
    final winnerCount = payload['winner_count'] as int? ?? 1;
    final maxPlayers = payload['max_players'] as int? ?? 10;

    setState(() {
      _phase = phase;
      _round = currentRound;
      _poolAmount = poolAmount;
      _betAmount = betAmount;
      _roomName = roomName;
      _isSpectator = isSpectator;
      _maxSpectators = maxSpectators;
      _currentTheme = themeName;
      _themeConfig = getThemeConfig(themeName);
      _winnerCount = winnerCount;
      _maxPlayers = maxPlayers;

      _players
        ..clear()
        ..addAll(players);

      _spectators
        ..clear()
        ..addAll(spectators);

      _syncCurrentUserAutoReady();
    });

    if (phaseEndTime != null) {
      _updatePhaseEndTime(phaseEndTime);
    }
  }

  void _handlePhaseChange(Map<String, dynamic> payload) {
    final phase = payload['phase'] as String? ?? _phase;
    final phaseEndTime = payload['phase_end_time'] as int?;
    final round = payload['round'] as int? ?? _round;

    setState(() {
      _phase = phase;
      _round = round;
    });

    if (phaseEndTime != null) {
      _updatePhaseEndTime(phaseEndTime);
    }
  }

  void _handlePhaseTick(Map<String, dynamic> payload) {
    // 增量更新：只处理服务器发送的变化字段
    final phaseEndTime = payload['phase_end_time'] as int?;
    final phase = payload['phase'] as String?;
    final poolAmount = payload['pool_amount'] as String?;
    final timeRemaining = payload['time_remaining'] as int?;

    setState(() {
      if (phase != null) {
        _phase = phase;
      }
      if (poolAmount != null) {
        _poolAmount = poolAmount;
      }
      // 使用服务器提供的剩余时间进行校准
      if (timeRemaining != null) {
        _timeLeft = (timeRemaining / 1000).ceil();
        if (_timeLeft < 0) _timeLeft = 0;
      }
    });

    // 如果服务器提供了阶段结束时间，更新本地计时器
    if (phaseEndTime != null && phaseEndTime != _phaseEndTimeMs) {
      _updatePhaseEndTime(phaseEndTime);
    }
  }

  void _handlePlayerJoin(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    final username = payload['username'] as String? ?? 'Player';
    if (userId == null) return;

    setState(() {
      final index = _players.indexWhere((p) => p.userId == userId);
      if (index >= 0) {
        _players[index] = _players[index].copyWith(
          isOnline: true,
        );
      } else {
        _players.add(
          _PlayerView(
            userId: userId,
            username: username,
            autoReady: false,
            isOnline: true,
          ),
        );
      }
      _syncCurrentUserAutoReady();
    });
  }

  void _handlePlayerLeave(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    if (userId == null) return;

    setState(() {
      // 玩家真正离开房间，从列表中移除
      _players.removeWhere((p) => p.userId == userId);
      _syncCurrentUserAutoReady();
    });
  }

  void _handlePlayerUpdate(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    if (userId == null) return;

    setState(() {
      final index = _players.indexWhere((p) => p.userId == userId);
      if (index >= 0) {
        final autoReady = payload['auto_ready'] as bool?;
        final isOnline = payload['is_online'] as bool?;
        final disqualified = payload['disqualified'] as bool?;
        final disqualifyReason = payload['disqualify_reason'] as String?;
        _players[index] = _players[index].copyWith(
          autoReady: autoReady ?? _players[index].autoReady,
          isOnline: isOnline ?? _players[index].isOnline,
          disqualified: disqualified ?? _players[index].disqualified,
          disqualifyReason: disqualifyReason ?? _players[index].disqualifyReason,
        );
      }
      _syncCurrentUserAutoReady();
    });
  }

  void _updatePhaseEndTime(int phaseEndTimeMs) {
    _phaseEndTimeMs = phaseEndTimeMs;
    _timer?.cancel();
    _tickTimeLeft();
    _timer = Timer.periodic(const Duration(seconds: 1), (_) {
      _tickTimeLeft();
    });
  }

  void _tickTimeLeft() {
    final end = _phaseEndTimeMs;
    if (end == null) return;

    final nowMs = DateTime.now().millisecondsSinceEpoch;
    // 使用 round() 而不是 ceil() 避免显示比实际多1秒
    // 加上 500ms 偏移量使得在 0.5s 时显示为 1s，在 1.5s 时显示为 2s
    var secondsLeft = ((end - nowMs + 500) / 1000).floor();
    if (secondsLeft < 0) {
      secondsLeft = 0;
    }

    setState(() {
      _timeLeft = secondsLeft;
    });
  }

  void _syncCurrentUserAutoReady() {
    final auth = ref.read(authProvider);
    final userId = auth.userId;
    if (userId == null) return;

    for (final p in _players) {
      if (p.userId == userId) {
        _autoReady = p.autoReady;
        break;
      }
    }
  }

  void _pushEvent(_RoundEvent event) {
    setState(() {
      _events.insert(0, event);
      if (_events.length > 5) {
        _events.removeLast();
      }
    });
  }

  void _handleBettingDone(Map<String, dynamic> payload) {
    final pool = payload['pool_amount']?.toString() ?? '0';
    final participants = (payload['participants'] as List?)?.length ?? 0;

    // 播放金币音效
    ref.read(audioServiceProvider).playCoinSound();

    final l10n = AppLocalizations.of(context);
    final title = l10n != null
        ? l10n.bettingComplete(participants, pool)
        : 'Betting complete: $participants joined, Pool ¥$pool';

    _pushEvent(_RoundEvent(
      type: 'betting_done',
      title: title,
      subtitle: '',
      time: DateTime.now(),
    ));
  }

  void _handleRoundResult(Map<String, dynamic> payload) {
    final prize = payload['prize_per_winner']?.toString() ?? '0';
    final winnerNames =
        (payload['winner_names'] as List?)?.map((e) => e.toString()).toList() ??
            <String>[];
    final winnerIds =
        (payload['winners'] as List?)?.map((e) => e as int).toList() ??
            <int>[];

    final l10n = AppLocalizations.of(context);
    // 简化显示：只显示结果，不显示轮数
    final title = winnerNames.isEmpty
        ? (l10n?.noWinner ?? 'No winner')
        : (l10n?.wonAmount(winnerNames.join('、'), prize) ?? '${winnerNames.join('、')} won ¥$prize');

    _pushEvent(_RoundEvent(
      type: 'round_result',
      title: title,
      subtitle: '',
      time: DateTime.now(),
    ));

    // 检查当前用户是否获胜
    final auth = ref.read(authProvider);
    final currentUserId = auth.userId;
    
    if (winnerIds.isNotEmpty && currentUserId != null) {
      if (winnerIds.contains(currentUserId)) {
        // 当前用户获胜，显示全屏庆祝
        setState(() {
          _showWinCelebration = true;
          _winPrize = prize;
        });
      } else if (winnerNames.isNotEmpty) {
        // 其他玩家获胜，显示横幅提示
        setState(() {
          _showOtherWinBanner = true;
          _otherWinnerNames = winnerNames;
          _otherWinPrize = prize;
        });
      }
    }
  }

  void _handleRoundFailed(Map<String, dynamic> payload) {
    final reason = payload['reason']?.toString() ?? 'round_failed';

    final l10n = AppLocalizations.of(context);
    // 简化失败原因显示
    String displayReason;
    if (reason.contains('not_enough_players')) {
      displayReason = l10n?.notEnoughPlayers ?? 'Not enough players';
    } else {
      displayReason = l10n?.gameCancelled ?? 'Game cancelled';
    }

    _pushEvent(_RoundEvent(
      type: 'round_failed',
      title: displayReason,
      subtitle: '',
      time: DateTime.now(),
    ));
  }

  void _handlePlayerDisqualified(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    final reason = payload['reason'] as String? ?? 'unknown';
    
    // 检查是否是当前用户被取消资格
    final auth = ref.read(authProvider);
    final currentUserId = auth.userId;
    
    if (userId == currentUserId) {
      // 当前用户被取消资格，显示提示并取消准备状态
      final l10n = AppLocalizations.of(context);
      String message;
      if (reason == 'insufficient_balance') {
        message = l10n?.insufficientBalanceDisqualified ?? 'Insufficient balance, you cannot participate in this round';
      } else {
        message = l10n?.disqualifiedFromRound ?? 'You have been disqualified from this round';
      }
      
      // 显示 SnackBar 提示
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(message),
            backgroundColor: Colors.orange,
            duration: const Duration(seconds: 5),
          ),
        );
      }
      
      // 自动取消准备状态
      setState(() {
        _autoReady = false;
      });
      final wsClient = ref.read(wsClientProvider);
      wsClient.setAutoReady(false);
    }
  }

  void _handleRoundCancelled(Map<String, dynamic> payload) {
    final reason = payload['reason'] as String? ?? 'unknown';
    final minRequired = payload['min_players_required'] as int? ?? 0;
    final currentPlayers = payload['current_players'] as int? ?? 0;
    final disqualifiedList = payload['disqualified_players'] as List? ?? [];
    
    final l10n = AppLocalizations.of(context);
    
    // 构建取消原因消息
    String displayReason;
    if (reason == 'not_enough_players') {
      displayReason = l10n?.roundCancelledNotEnoughPlayers(minRequired, currentPlayers) 
          ?? 'Round cancelled: need $minRequired players, only $currentPlayers qualified';
    } else {
      displayReason = l10n?.gameCancelled ?? 'Game cancelled';
    }
    
    // 如果有被取消资格的玩家，添加到消息中
    if (disqualifiedList.isNotEmpty) {
      final names = disqualifiedList
          .map((p) => (p as Map<String, dynamic>)['username'] as String? ?? '')
          .where((name) => name.isNotEmpty)
          .toList();
      if (names.isNotEmpty) {
        displayReason += ' (${names.join(', ')} ${l10n?.insufficientBalance ?? 'insufficient balance'})';
      }
    }
    
    _pushEvent(_RoundEvent(
      type: 'round_cancelled',
      title: displayReason,
      subtitle: '',
      time: DateTime.now(),
    ));
  }

  void _handleSpectatorJoin(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    final username = payload['username'] as String? ?? 'Spectator';
    if (userId == null) return;

    setState(() {
      final index = _spectators.indexWhere((s) => s.userId == userId);
      if (index < 0) {
        _spectators.add(_SpectatorView(userId: userId, username: username));
      }
    });
  }

  void _handleSpectatorLeave(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    if (userId == null) return;

    setState(() {
      _spectators.removeWhere((s) => s.userId == userId);
    });
  }

  void _handleSpectatorSwitch(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int?;
    final username = payload['username'] as String? ?? 'Player';
    if (userId == null) return;

    setState(() {
      // 从观战者列表移除
      _spectators.removeWhere((s) => s.userId == userId);
      // 添加到玩家列表
      final index = _players.indexWhere((p) => p.userId == userId);
      if (index < 0) {
        _players.add(_PlayerView(
          userId: userId,
          username: username,
          autoReady: false,
          isOnline: true,
        ));
      }
      _syncCurrentUserAutoReady();
    });
  }

  void _switchToParticipant() {
    final wsClient = ref.read(wsClientProvider);
    wsClient.switchToParticipant();
  }

  void _handleChatMessage(Map<String, dynamic> payload) {
    final msg = ChatMessage.fromJson(payload);
    setState(() {
      _chatMessages.add(msg);
      // 保持最多100条消息
      if (_chatMessages.length > 100) {
        _chatMessages.removeAt(0);
      }
    });
  }

  static const int _maxEmojiReactions = 5; // 最大同时显示的表情数量

  void _handleEmojiReaction(Map<String, dynamic> payload) {
    final userId = payload['user_id'] as int? ?? 0;
    final username = payload['username'] as String? ?? '';
    final emoji = payload['emoji'] as String? ?? '';

    setState(() {
      // 限制同时显示的表情数量，移除最旧的
      if (_emojiReactions.length >= _maxEmojiReactions) {
        _emojiReactions.removeAt(0);
      }
      _emojiReactions.add(_EmojiReaction(
        userId: userId,
        username: username,
        emoji: emoji,
        id: DateTime.now().millisecondsSinceEpoch,
      ));
    });
  }

  void _handleThemeChange(Map<String, dynamic> payload) {
    final themeName = payload['theme_name'] as String? ?? 'classic';
    setState(() {
      _currentTheme = themeName;
      _themeConfig = getThemeConfig(themeName);
    });
  }

  void _sendChatMessage(String content) {
    final wsClient = ref.read(wsClientProvider);
    wsClient.sendChatMessage(content);
  }

  void _sendEmoji(String emoji) {
    final wsClient = ref.read(wsClientProvider);
    wsClient.sendEmoji(emoji);
    setState(() {
      _showEmojiPicker = false;
    });
  }

  void _removeEmojiReaction(int id) {
    setState(() {
      _emojiReactions.removeWhere((r) => r.id == id);
    });
  }

  /// 离开房间
  void _leaveRoom(BuildContext context) {
    // 发送 WebSocket 消息通知后端
    final wsClient = ref.read(wsClientProvider);
    wsClient.leaveRoom();
    // 跳转到首页
    context.go('/home');
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Text(_roomName.isEmpty ? 'Room ${widget.roomId}' : _roomName),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () => _leaveRoom(context),
          ),
          actions: [
            IconButton(
              icon: Icon(
                _showChat ? Icons.chat : Icons.chat_outlined,
                color: _showChat ? AppColors.accent : AppColors.textPrimary,
              ),
              onPressed: () {
                setState(() {
                  _showChat = !_showChat;
                  if (_showChat) _showEmojiPicker = false;
                });
              },
            ),
            IconButton(
              icon: Icon(
                _showEmojiPicker ? Icons.emoji_emotions : Icons.emoji_emotions_outlined,
                color: _showEmojiPicker ? AppColors.accent : AppColors.textPrimary,
              ),
              onPressed: () {
                setState(() {
                  _showEmojiPicker = !_showEmojiPicker;
                  if (_showEmojiPicker) _showChat = false;
                });
              },
            ),
            IconButton(
              icon: Icon(Icons.info_outline, color: AppColors.textPrimary),
              onPressed: () {
                // TODO: 显示房间信息
              },
            ),
          ],
        ),
        body: Stack(
        children: [
          Column(
            children: [
              // Game status bar
          Container(
            padding: EdgeInsets.all(16.w),
            decoration: BoxDecoration(
              color: _getPhaseColor(),
              boxShadow: [
                BoxShadow(
                  color: _getPhaseColor().withAlpha(76),
                  blurRadius: 8,
                  offset: const Offset(0, 4),
                ),
              ],
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceAround,
              children: [
                _buildStatusItem(l10n.betPerRound, '¥$_betAmount'),
                _buildStatusItem(_getPhaseText(l10n), '${_timeLeft}s'),
                _buildStatusItem(l10n.poolAmount, '¥$_poolAmount'),
              ],
            ),
          ),

          // Main game area
          Expanded(
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  // Timer circle
                  Container(
                    width: 200.w,
                    height: 200.w,
                    decoration: BoxDecoration(
                      shape: BoxShape.circle,
                      border: Border.all(
                        color: _getPhaseColor(),
                        width: 8,
                      ),
                    ),
                    child: Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Text(
                            '$_timeLeft',
                            style: TextStyle(
                              fontSize: 64.sp,
                              fontWeight: FontWeight.bold,
                              color: _getPhaseColor(),
                            ),
                          ),
                          Text(
                            _getPhaseText(l10n),
                            style: TextStyle(
                              fontSize: 18.sp,
                              color: AppColors.textSecondary,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                  SizedBox(height: 16.h),

                  // 游戏记录（只显示最近3条）
                  if (_events.isNotEmpty)
                    Padding(
                      padding: EdgeInsets.symmetric(horizontal: 24.w),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            l10n.gameHistory,
                            style: TextStyle(
                              fontSize: 14.sp,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textSecondary,
                            ),
                          ),
                          SizedBox(height: 8.h),
                          ..._events.take(3).map(
                            (e) => Padding(
                              padding: EdgeInsets.only(bottom: 6.h),
                              child: Row(
                                children: [
                                  Icon(
                                    e.type == 'round_result'
                                        ? Icons.emoji_events_outlined
                                        : (e.type == 'round_failed' || e.type == 'round_cancelled')
                                            ? Icons.error_outline
                                            : Icons.casino_outlined,
                                    size: 16.sp,
                                    color: (e.type == 'round_failed' || e.type == 'round_cancelled')
                                        ? AppColors.error
                                        : AppColors.textSecondary,
                                  ),
                                  SizedBox(width: 6.w),
                                  Text(
                                    '${e.time.hour.toString().padLeft(2, '0')}:${e.time.minute.toString().padLeft(2, '0')}',
                                    style: TextStyle(
                                      fontSize: 11.sp,
                                      color: AppColors.textSecondary.withAlpha(150),
                                    ),
                                  ),
                                  SizedBox(width: 6.w),
                                  Expanded(
                                    child: Text(
                                      e.title,
                                      style: TextStyle(
                                        fontSize: 12.sp,
                                        color: AppColors.textSecondary,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),

                  SizedBox(height: 16.h),

                  // Auto ready toggle or spectator switch button
                  if (_isSpectator)
                    GlassCard(
                      padding: EdgeInsets.symmetric(
                        horizontal: 24.w,
                        vertical: 12.h,
                      ),
                      child: Column(
                        children: [
                          Container(
                            padding: EdgeInsets.symmetric(
                              horizontal: 12.w,
                              vertical: 4.h,
                            ),
                            decoration: BoxDecoration(
                              color: AppColors.phaseWaiting.withAlpha(51),
                              borderRadius: BorderRadius.circular(8.r),
                            ),
                            child: Text(
                              '观战中',
                              style: TextStyle(
                                fontSize: 14.sp,
                                color: AppColors.phaseWaiting,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ),
                          SizedBox(height: 12.h),
                          GradientButton(
                            onPressed: _switchToParticipant,
                            gradient: AppColors.gradientAccent,
                            height: 40.h,
                            child: Padding(
                              padding: EdgeInsets.symmetric(horizontal: 16.w),
                              child: Text(
                                l10n.switchToParticipant,
                                style: TextStyle(
                                  color: Colors.white,
                                  fontSize: 14.sp,
                                  fontWeight: FontWeight.w500,
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    )
                  else
                    GlassCard(
                      padding: EdgeInsets.symmetric(
                        horizontal: 24.w,
                        vertical: 12.h,
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Text(
                            l10n.autoReady,
                            style: TextStyle(
                              fontSize: 16.sp,
                              fontWeight: FontWeight.w500,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          SizedBox(width: 16.w),
                          Switch(
                            value: _autoReady,
                            onChanged: (value) {
                              setState(() => _autoReady = value);
                              final wsClient = ref.read(wsClientProvider);
                              wsClient.setAutoReady(value);
                            },
                            activeColor: AppColors.accent,
                            activeTrackColor: AppColors.accent.withAlpha(100),
                            inactiveThumbColor: AppColors.textSecondary,
                            inactiveTrackColor: AppColors.glassWhite,
                          ),
                        ],
                      ),
                    ),
                ],
              ),
            ),
          ),

          // Players list
          Container(
            height: 200.h,
            decoration: BoxDecoration(
              color: AppColors.cardDark.withAlpha(230),
              borderRadius: BorderRadius.vertical(
                top: Radius.circular(24.r),
              ),
              border: Border(
                top: BorderSide(color: AppColors.glassBorder, width: 1),
              ),
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withAlpha(50),
                  blurRadius: 12,
                  offset: const Offset(0, -4),
                ),
              ],
            ),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.all(16.w),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Row(
                        children: [
                          Text(
                            '${l10n.players} (${_players.length})',
                            style: TextStyle(
                              fontSize: 16.sp,
                              fontWeight: FontWeight.bold,
                              color: AppColors.textPrimary,
                            ),
                          ),
                          SizedBox(width: 8.w),
                          // 显示最小开始人数提示
                          Container(
                            padding: EdgeInsets.symmetric(
                              horizontal: 6.w,
                              vertical: 2.h,
                            ),
                            decoration: BoxDecoration(
                              color: _players.where((p) => p.isOnline && p.autoReady).length >= _winnerCount + 1
                                  ? AppColors.success.withAlpha(51)
                                  : AppColors.warning.withAlpha(51),
                              borderRadius: BorderRadius.circular(8.r),
                            ),
                            child: Text(
                              l10n.minPlayersToStart(_winnerCount + 1),
                              style: TextStyle(
                                fontSize: 10.sp,
                                color: _players.where((p) => p.isOnline && p.autoReady).length >= _winnerCount + 1
                                    ? AppColors.success
                                    : AppColors.warning,
                              ),
                            ),
                          ),
                          SizedBox(width: 8.w),
                          Container(
                            padding: EdgeInsets.symmetric(
                              horizontal: 8.w,
                              vertical: 2.h,
                            ),
                            decoration: BoxDecoration(
                              color: AppColors.textSecondary.withAlpha(51),
                              borderRadius: BorderRadius.circular(8.r),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Icon(
                                  Icons.visibility,
                                  size: 14.sp,
                                  color: AppColors.textSecondary,
                                ),
                                SizedBox(width: 4.w),
                                Text(
                                  '${_spectators.length}/$_maxSpectators',
                                  style: TextStyle(
                                    fontSize: 12.sp,
                                    color: AppColors.textSecondary,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      TextButton(
                        onPressed: () => _leaveRoom(context),
                        child: Text(l10n.leaveRoom),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  child: ListView.builder(
                    scrollDirection: Axis.horizontal,
                    padding: EdgeInsets.symmetric(horizontal: 16.w),
                    itemCount: _players.length,
                    itemBuilder: (context, index) {
                      final player = _players[index];
                      final isReady = player.autoReady;
                      final isDisqualified = player.disqualified;

                      // 确定头像颜色
                      Color avatarColor;
                      if (isDisqualified) {
                        avatarColor = AppColors.error.withAlpha(150);
                      } else if (player.isOnline) {
                        avatarColor = AppColors.primary;
                      } else {
                        avatarColor = AppColors.textSecondary;
                      }

                      // 确定状态标签
                      String statusText;
                      Color statusColor;
                      Color statusBgColor;
                      
                      if (isDisqualified) {
                        statusText = l10n.insufficientBalance;
                        statusColor = AppColors.error;
                        statusBgColor = AppColors.error.withAlpha(51);
                      } else if (isReady) {
                        statusText = l10n.ready;
                        statusColor = AppColors.success;
                        statusBgColor = AppColors.success.withAlpha(51);
                      } else {
                        statusText = l10n.notReady;
                        statusColor = AppColors.textSecondary;
                        statusBgColor = AppColors.textSecondary.withAlpha(51);
                      }

                      return Padding(
                        padding: EdgeInsets.only(right: 12.w),
                        child: Column(
                          children: [
                            Stack(
                              children: [
                                CircleAvatar(
                                  radius: 28.r,
                                  backgroundColor: avatarColor,
                                  child: Text(
                                    player.username.isNotEmpty
                                        ? player.username.substring(0, 1).toUpperCase()
                                        : 'P',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                                // 被取消资格时显示警告图标
                                if (isDisqualified)
                                  Positioned(
                                    right: 0,
                                    bottom: 0,
                                    child: Container(
                                      padding: EdgeInsets.all(2.w),
                                      decoration: BoxDecoration(
                                        color: AppColors.error,
                                        shape: BoxShape.circle,
                                      ),
                                      child: Icon(
                                        Icons.warning,
                                        size: 12.sp,
                                        color: Colors.white,
                                      ),
                                    ),
                                  ),
                              ],
                            ),
                            SizedBox(height: 4.h),
                            Text(
                              player.username,
                              style: TextStyle(
                                fontSize: 12.sp,
                                color: isDisqualified 
                                    ? AppColors.textSecondary 
                                    : AppColors.textPrimary,
                              ),
                            ),
                            Container(
                              padding: EdgeInsets.symmetric(
                                horizontal: 8.w,
                                vertical: 2.h,
                              ),
                              decoration: BoxDecoration(
                                color: statusBgColor,
                                borderRadius: BorderRadius.circular(8.r),
                              ),
                              child: Text(
                                statusText,
                                style: TextStyle(
                                  fontSize: 10.sp,
                                  color: statusColor,
                                ),
                              ),
                            ),
                          ],
                        ),
                      );
                    },
                  ),
                ),
                ],
          ),
        ),
          ],
        ),

          // Chat panel overlay
          if (_showChat)
            Positioned(
              right: 0,
              top: 80.h,
              bottom: 200.h,
              width: 300.w,
              child: Container(
                margin: EdgeInsets.all(8.w),
                decoration: BoxDecoration(
                  color: AppColors.cardDark.withAlpha(240),
                  borderRadius: BorderRadius.circular(12.r),
                  border: Border.all(color: AppColors.glassBorder),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.black.withAlpha(50),
                      blurRadius: 12,
                    ),
                  ],
                ),
                child: ChatWidget(
                  messages: _chatMessages,
                  onSendMessage: _sendChatMessage,
                  currentUserId: ref.read(authProvider).userId,
                  enabled: !_isSpectator,
                ),
              ),
            ),

          // Emoji picker overlay
          if (_showEmojiPicker)
            Positioned(
              right: 8.w,
              top: 80.h,
              child: EmojiPicker(
                onEmojiSelected: _sendEmoji,
                enabled: !_isSpectator,
              ),
            ),

          // Emoji reactions overlay
          ..._emojiReactions.map((reaction) {
            return Positioned(
              left: MediaQuery.of(context).size.width / 2 - 40.w,
              top: MediaQuery.of(context).size.height / 3,
              child: EmojiReactionOverlay(
                key: ValueKey(reaction.id),
                emoji: reaction.emoji,
                username: reaction.username,
                onComplete: () => _removeEmojiReaction(reaction.id),
              ),
            );
          }),

          // 自己获胜的全屏庆祝
          if (_showWinCelebration)
            Positioned.fill(
              child: WinCelebrationOverlay(
                prize: _winPrize,
                onDismiss: () {
                  setState(() {
                    _showWinCelebration = false;
                  });
                },
              ),
            ),

          // 其他玩家获胜的横幅提示
          if (_showOtherWinBanner)
            Positioned(
              top: 0,
              left: 0,
              right: 0,
              child: OtherWinBanner(
                winnerNames: _otherWinnerNames,
                prize: _otherWinPrize,
                onDismiss: () {
                  setState(() {
                    _showOtherWinBanner = false;
                  });
                },
              ),
            ),
        ],
      ),
      ),
    );
  }

  Widget _buildStatusItem(String label, String value) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: 24.sp,
            fontWeight: FontWeight.bold,
            color: Colors.white,
          ),
        ),
        Text(
          label,
          style: TextStyle(
            fontSize: 12.sp,
            color: Colors.white.withAlpha(230),
          ),
        ),
      ],
    );
  }

  Color _getPhaseColor() {
    // 如果有主题配置，使用主题的主色调
    final themeColor = _themeConfig?.primaryColor;
    
    switch (_phase) {
      case 'waiting':
        return themeColor ?? AppColors.phaseWaiting;
      case 'countdown':
        return AppColors.phaseCountdown;
      case 'betting':
        return AppColors.phaseBetting;
      case 'in_game':
        return themeColor ?? AppColors.phaseInGame;
      case 'settlement':
        return _themeConfig?.secondaryColor ?? AppColors.phaseSettlement;
      default:
        return themeColor ?? AppColors.primary;
    }
  }

  /// 获取主题的主色调
  Color get _themePrimaryColor => _themeConfig?.primaryColor ?? AppColors.primary;
  
  /// 获取主题的次色调
  Color get _themeSecondaryColor => _themeConfig?.secondaryColor ?? AppColors.primaryLight;
  
  /// 获取主题的背景色
  Color get _themeBackgroundColor => _themeConfig?.backgroundColor ?? AppColors.backgroundLight;
  
  /// 获取主题的文字颜色
  Color get _themeTextColor => _themeConfig?.textColor ?? AppColors.textPrimary;

  String _getPhaseText(AppLocalizations l10n) {
    switch (_phase) {
      case GamePhase.waiting:
        return l10n.phaseWaiting;
      case GamePhase.countdown:
        return l10n.phaseCountdown;
      case GamePhase.betting:
        return l10n.phaseBetting;
      case GamePhase.inGame:
        return l10n.phaseInGame;
      case GamePhase.settlement:
        return l10n.phaseSettlement;
      case GamePhase.reset:
        return l10n.phaseReset;
      default:
        return '';
    }
  }
}

class _PlayerView {
  final int userId;
  final String username;
  final bool autoReady;
  final bool isOnline;
  final bool disqualified;
  final String disqualifyReason;

  const _PlayerView({
    required this.userId,
    required this.username,
    required this.autoReady,
    required this.isOnline,
    this.disqualified = false,
    this.disqualifyReason = '',
  });

  _PlayerView copyWith({
    bool? autoReady,
    bool? isOnline,
    bool? disqualified,
    String? disqualifyReason,
  }) {
    return _PlayerView(
      userId: userId,
      username: username,
      autoReady: autoReady ?? this.autoReady,
      isOnline: isOnline ?? this.isOnline,
      disqualified: disqualified ?? this.disqualified,
      disqualifyReason: disqualifyReason ?? this.disqualifyReason,
    );
  }
}

class _SpectatorView {
  final int userId;
  final String username;

  const _SpectatorView({
    required this.userId,
    required this.username,
  });
}


class _EmojiReaction {
  final int id;
  final int userId;
  final String username;
  final String emoji;

  const _EmojiReaction({
    required this.id,
    required this.userId,
    required this.username,
    required this.emoji,
  });
}
