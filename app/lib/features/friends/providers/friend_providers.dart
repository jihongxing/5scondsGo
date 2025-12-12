import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/services/api_client.dart';

// 好友信息模型
class FriendInfo {
  final int id;
  final String username;
  final String role;
  final bool isOnline;
  final int? currentRoom;
  final String? roomName;

  FriendInfo({
    required this.id,
    required this.username,
    required this.role,
    this.isOnline = false,
    this.currentRoom,
    this.roomName,
  });

  factory FriendInfo.fromJson(Map<String, dynamic> json) {
    return FriendInfo(
      id: json['id'] as int,
      username: json['username'] as String,
      role: json['role'] as String,
      isOnline: json['is_online'] as bool? ?? false,
      currentRoom: json['current_room'] as int?,
      roomName: json['room_name'] as String?,
    );
  }
}

// 好友请求模型
class FriendRequest {
  final int id;
  final int fromUserId;
  final int toUserId;
  final String status;
  final DateTime createdAt;
  final String? fromUsername;

  FriendRequest({
    required this.id,
    required this.fromUserId,
    required this.toUserId,
    required this.status,
    required this.createdAt,
    this.fromUsername,
  });

  factory FriendRequest.fromJson(Map<String, dynamic> json) {
    return FriendRequest(
      id: json['id'] as int,
      fromUserId: json['from_user_id'] as int,
      toUserId: json['to_user_id'] as int,
      status: json['status'] as String,
      createdAt: DateTime.parse(json['created_at'] as String),
      fromUsername: (json['from_user'] as Map<String, dynamic>?)?['username'] as String?,
    );
  }
}

// 好友列表状态
class FriendListState {
  final List<FriendInfo> friends;
  final bool isLoading;
  final String? error;

  FriendListState({
    this.friends = const [],
    this.isLoading = false,
    this.error,
  });

  FriendListState copyWith({
    List<FriendInfo>? friends,
    bool? isLoading,
    String? error,
  }) {
    return FriendListState(
      friends: friends ?? this.friends,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

// 好友列表 Notifier
class FriendListNotifier extends StateNotifier<FriendListState> {
  final ApiClient _apiClient;

  FriendListNotifier(this._apiClient) : super(FriendListState());

  Future<void> loadFriends() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final data = await _apiClient.getFriendList();
      final friends = data.map((e) => FriendInfo.fromJson(e)).toList();
      state = state.copyWith(friends: friends, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> removeFriend(int friendId) async {
    try {
      await _apiClient.removeFriend(friendId);
      state = state.copyWith(
        friends: state.friends.where((f) => f.id != friendId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  void updateFriendOnlineStatus(int friendId, bool isOnline, int? roomId) {
    final friends = state.friends.map((f) {
      if (f.id == friendId) {
        return FriendInfo(
          id: f.id,
          username: f.username,
          role: f.role,
          isOnline: isOnline,
          currentRoom: roomId,
          roomName: f.roomName,
        );
      }
      return f;
    }).toList();
    state = state.copyWith(friends: friends);
  }
}

// 好友请求状态
class FriendRequestsState {
  final List<FriendRequest> requests;
  final bool isLoading;
  final String? error;

  FriendRequestsState({
    this.requests = const [],
    this.isLoading = false,
    this.error,
  });

  FriendRequestsState copyWith({
    List<FriendRequest>? requests,
    bool? isLoading,
    String? error,
  }) {
    return FriendRequestsState(
      requests: requests ?? this.requests,
      isLoading: isLoading ?? this.isLoading,
      error: error,
    );
  }
}

// 好友请求 Notifier
class FriendRequestsNotifier extends StateNotifier<FriendRequestsState> {
  final ApiClient _apiClient;

  FriendRequestsNotifier(this._apiClient) : super(FriendRequestsState());

  Future<void> loadRequests() async {
    state = state.copyWith(isLoading: true, error: null);
    try {
      final data = await _apiClient.getPendingFriendRequests();
      final requests = data.map((e) => FriendRequest.fromJson(e)).toList();
      state = state.copyWith(requests: requests, isLoading: false);
    } catch (e) {
      state = state.copyWith(isLoading: false, error: e.toString());
    }
  }

  Future<void> acceptRequest(int requestId) async {
    try {
      await _apiClient.acceptFriendRequest(requestId);
      state = state.copyWith(
        requests: state.requests.where((r) => r.id != requestId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> rejectRequest(int requestId) async {
    try {
      await _apiClient.rejectFriendRequest(requestId);
      state = state.copyWith(
        requests: state.requests.where((r) => r.id != requestId).toList(),
      );
    } catch (e) {
      state = state.copyWith(error: e.toString());
    }
  }

  Future<void> sendRequest(int toUserId) async {
    try {
      await _apiClient.sendFriendRequest(toUserId);
    } catch (e) {
      state = state.copyWith(error: e.toString());
      rethrow;
    }
  }
}

// Providers
final friendListProvider = StateNotifierProvider<FriendListNotifier, FriendListState>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return FriendListNotifier(apiClient);
});

final friendRequestsProvider = StateNotifierProvider<FriendRequestsNotifier, FriendRequestsState>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  return FriendRequestsNotifier(apiClient);
});
