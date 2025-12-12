import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'api_client.dart';

const String wsUrl = kIsWeb ? 'ws://localhost:8080/ws' : 'ws://10.0.2.2:8080/ws';

// WebSocket 消息类型
class WSMessageType {
  static const String auth = 'auth';
  static const String joinRoom = 'join_room';
  static const String leaveRoom = 'leave_room';
  static const String setAutoReady = 'set_auto_ready';
  static const String joinAsSpectator = 'join_as_spectator';
  static const String switchToParticipant = 'switch_to_participant';
  
  // 服务端推送
  static const String roomState = 'room_state';
  static const String phaseChange = 'phase_change';
  static const String phaseTick = 'phase_tick';
  static const String playerJoin = 'player_join';
  static const String playerLeave = 'player_leave';
  static const String playerUpdate = 'player_update';
  static const String bettingDone = 'betting_done';
  static const String roundResult = 'round_result';
  static const String roundFailed = 'round_failed';
  static const String balanceUpdate = 'balance_update';
  static const String error = 'error';
  static const String authSuccess = 'auth_success';
  
  // 玩家资格相关
  static const String playerDisqualified = 'player_disqualified';
  static const String roundCancelled = 'round_cancelled';
  
  // 观战者相关
  static const String spectatorJoin = 'spectator_join';
  static const String spectatorLeave = 'spectator_leave';
  static const String spectatorSwitch = 'spectator_switch';

  // 聊天相关
  static const String sendChat = 'send_chat';
  static const String chatMessage = 'chat_message';
  static const String chatHistory = 'chat_history';

  // 表情相关
  static const String sendEmoji = 'send_emoji';
  static const String emojiReaction = 'emoji_reaction';

  // 主题相关
  static const String themeChange = 'theme_change';
}

// 游戏阶段
class GamePhase {
  static const String waiting = 'waiting';
  static const String countdown = 'countdown';
  static const String betting = 'betting';
  static const String inGame = 'in_game';
  static const String settlement = 'settlement';
  static const String reset = 'reset';
}

class WSClient {
  WebSocketChannel? _channel;
  final ApiClient _apiClient;
  
  StreamController<Map<String, dynamic>>? _messageController;
  Stream<Map<String, dynamic>> get messages => _messageController!.stream;
  
  bool _isConnected = false;
  bool get isConnected => _isConnected;
  
  Timer? _heartbeatTimer;
  Timer? _heartbeatTimeoutTimer;
  Timer? _reconnectTimer;
  int _reconnectAttempts = 0;
  static const int maxReconnectAttempts = 5;
  static const Duration heartbeatInterval = Duration(seconds: 30);
  static const Duration heartbeatTimeout = Duration(seconds: 10);
  DateTime? _lastPongTime;

  WSClient(this._apiClient);

  Future<void> connect() async {
    if (_isConnected) return;
    
    final token = _apiClient.token;
    if (token == null) {
      throw Exception('Not authenticated');
    }

    try {
      _messageController?.close();
      _messageController = StreamController<Map<String, dynamic>>.broadcast();
      
      _channel = WebSocketChannel.connect(Uri.parse('$wsUrl?token=$token'));
      
      _channel!.stream.listen(
        _onMessage,
        onError: _onError,
        onDone: _onDone,
      );
      
      _isConnected = true;
      _reconnectAttempts = 0;
      _startHeartbeat();
    } catch (e) {
      _isConnected = false;
      _scheduleReconnect();
      rethrow;
    }
  }

  void _onMessage(dynamic data) {
    try {
      final message = jsonDecode(data as String) as Map<String, dynamic>;
      
      // 收到任何消息都更新最后活跃时间
      _lastPongTime = DateTime.now();
      _cancelHeartbeatTimeout();
      
      _messageController?.add(message);
    } catch (e) {
      // 忽略解析错误
    }
  }

  void _onError(dynamic error) {
    _isConnected = false;
    _stopHeartbeat();
    _scheduleReconnect();
  }

  void _onDone() {
    _isConnected = false;
    _stopHeartbeat();
    _scheduleReconnect();
  }

  void _startHeartbeat() {
    _heartbeatTimer?.cancel();
    _lastPongTime = DateTime.now();
    _heartbeatTimer = Timer.periodic(heartbeatInterval, (_) {
      if (_isConnected) {
        send({'type': 'ping'});
        _startHeartbeatTimeout();
      }
    });
  }

  void _startHeartbeatTimeout() {
    _cancelHeartbeatTimeout();
    _heartbeatTimeoutTimer = Timer(heartbeatTimeout, () {
      // 心跳超时，认为连接已断开
      if (_isConnected) {
        _isConnected = false;
        _onDone();
      }
    });
  }

  void _cancelHeartbeatTimeout() {
    _heartbeatTimeoutTimer?.cancel();
    _heartbeatTimeoutTimer = null;
  }

  void _stopHeartbeat() {
    _heartbeatTimer?.cancel();
    _heartbeatTimer = null;
    _cancelHeartbeatTimeout();
  }

  int? _lastRoomId; // 记录最后加入的房间ID，用于重连恢复

  void _scheduleReconnect() {
    if (_reconnectAttempts >= maxReconnectAttempts) {
      return;
    }
    
    _reconnectTimer?.cancel();
    final delay = Duration(seconds: 2 * (_reconnectAttempts + 1));
    _reconnectTimer = Timer(delay, () async {
      _reconnectAttempts++;
      try {
        await connect();
        // 重连成功后，自动重新加入之前的房间
        if (_lastRoomId != null && _isConnected) {
          joinRoom(_lastRoomId!);
        }
      } catch (e) {
        // 重连失败，会在 connect() 中再次调度
      }
    });
  }

  void send(Map<String, dynamic> message) {
    if (!_isConnected || _channel == null) return;
    _channel!.sink.add(jsonEncode(message));
  }

  void joinRoom(int roomId) {
    _lastRoomId = roomId;
    send({
      'type': WSMessageType.joinRoom,
      'payload': {'room_id': roomId},
    });
  }

  void leaveRoom() {
    _lastRoomId = null;
    send({'type': WSMessageType.leaveRoom});
  }

  void setAutoReady(bool autoReady) {
    send({
      'type': WSMessageType.setAutoReady,
      'payload': {'auto_ready': autoReady},
    });
  }

  void joinAsSpectator(int roomId) {
    send({
      'type': WSMessageType.joinAsSpectator,
      'payload': {'room_id': roomId},
    });
  }

  void switchToParticipant() {
    send({'type': WSMessageType.switchToParticipant});
  }

  void sendChatMessage(String content) {
    send({
      'type': WSMessageType.sendChat,
      'payload': {'content': content},
    });
  }

  void sendEmoji(String emoji) {
    send({
      'type': WSMessageType.sendEmoji,
      'payload': {'emoji': emoji},
    });
  }

  Future<void> disconnect() async {
    _stopHeartbeat();
    _reconnectTimer?.cancel();
    _reconnectTimer = null;
    _reconnectAttempts = maxReconnectAttempts; // 阻止自动重连
    _isConnected = false;
    await _channel?.sink.close();
    _channel = null;
  }

  void dispose() {
    disconnect();
    _messageController?.close();
  }
}

final wsClientProvider = Provider<WSClient>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return WSClient(apiClient);
});
