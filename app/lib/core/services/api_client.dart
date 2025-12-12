import 'package:dio/dio.dart';
import 'package:dio_smart_retry/dio_smart_retry.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../config/app_config.dart';
import 'secure_storage_service.dart';

class ApiClient {
  late final Dio _dio;
  String? _token;
  final SecureStorageService _secureStorage = SecureStorageService.instance;

  ApiClient() {
    _dio = Dio(BaseOptions(
      baseUrl: AppConfig.instance.apiBaseUrl,
      connectTimeout: const Duration(seconds: 15),
      receiveTimeout: const Duration(seconds: 15),
      sendTimeout: const Duration(seconds: 15),
      headers: {
        'Content-Type': 'application/json',
      },
    ));

    // 添加认证拦截器
    _dio.interceptors.add(InterceptorsWrapper(
      onRequest: (options, handler) {
        if (_token != null) {
          options.headers['Authorization'] = 'Bearer $_token';
        }
        return handler.next(options);
      },
      onError: (error, handler) {
        if (error.response?.statusCode == 401) {
          // Token 过期，清除登录状态
          clearToken();
        }
        return handler.next(error);
      },
    ));

    // 添加重试拦截器
    _dio.interceptors.add(RetryInterceptor(
      dio: _dio,
      retries: 3,
      retryDelays: const [
        Duration(seconds: 1),
        Duration(seconds: 2),
        Duration(seconds: 3),
      ],
      // 只对特定错误重试
      retryEvaluator: (error, attempt) {
        // 不重试认证错误
        if (error.response?.statusCode == 401 ||
            error.response?.statusCode == 403) {
          return false;
        }
        // 重试网络错误和服务器错误
        return error.type == DioExceptionType.connectionTimeout ||
            error.type == DioExceptionType.sendTimeout ||
            error.type == DioExceptionType.receiveTimeout ||
            error.type == DioExceptionType.connectionError ||
            (error.response?.statusCode ?? 0) >= 500;
      },
    ));
  }

  Future<void> init() async {
    // 从安全存储加载 Token
    _token = await _secureStorage.getToken();
  }

  Future<void> setToken(String token) async {
    _token = token;
    await _secureStorage.saveToken(token);
  }

  Future<void> clearToken() async {
    _token = null;
    await _secureStorage.deleteToken();
  }

  String? get token => _token;

  // ===== 认证 =====

  Future<Map<String, dynamic>> register({
    required String username,
    required String password,
    String? inviteCode,
    String? role,
  }) async {
    final response = await _dio.post('/auth/register', data: {
      'username': username,
      'password': password,
      if (inviteCode != null) 'invite_code': inviteCode,
      if (role != null) 'role': role,
    });
    return response.data;
  }

  Future<Map<String, dynamic>> login({
    required String username,
    required String password,
  }) async {
    final response = await _dio.post('/auth/login', data: {
      'username': username,
      'password': password,
    });

    if (response.data['token'] != null) {
      await setToken(response.data['token']);
    }
    return response.data;
  }

  Future<Map<String, dynamic>> getMe() async {
    final response = await _dio.get('/me');
    return response.data;
  }

  // ===== 房间 =====

  Future<List<Map<String, dynamic>>> listRooms({int? ownerId, String? inviteCode}) async {
    final response = await _dio.get('/rooms', queryParameters: {
      if (ownerId != null) 'owner_id': ownerId,
      if (inviteCode != null) 'invite_code': inviteCode,
    });
    final data = response.data;
    if (data is Map<String, dynamic>) {
      final items = data['items'];
      if (items is List) {
        return items.cast<Map<String, dynamic>>();
      }
    }
    return <Map<String, dynamic>>[];
  }

  Future<Map<String, dynamic>> getRoom(int id) async {
    final response = await _dio.get('/rooms/$id');
    return response.data;
  }

  Future<Map<String, dynamic>> joinRoom(int roomId, {String? password}) async {
    final response = await _dio.post('/rooms/$roomId/join', data: {
      if (password != null) 'password': password,
    });
    return response.data;
  }

  Future<void> leaveRoom() async {
    await _dio.post('/rooms/leave');
  }

  Future<void> setAutoReady(bool autoReady) async {
    await _dio.post('/rooms/auto-ready', data: {'auto_ready': autoReady});
  }

  Future<Map<String, dynamic>?> getMyRoom() async {
    try {
      final response = await _dio.get('/rooms/my');
      return response.data;
    } on DioException catch (e) {
      if (e.response?.statusCode == 404) {
        return null;
      }
      rethrow;
    }
  }

  // ===== 房主接口 =====

  Future<Map<String, dynamic>> createRoom({
    required String name,
    required double betAmount,
    required int winnerCount,
    required int maxPlayers,
    double? ownerCommission,
    String? password,
  }) async {
    try {
      final response = await _dio.post('/owner/rooms', data: {
        'name': name,
        'bet_amount': betAmount.toStringAsFixed(2),
        'winner_count': winnerCount,
        'max_players': maxPlayers,
        if (ownerCommission != null) 'owner_commission_rate': ownerCommission.toStringAsFixed(4),
        if (password != null && password.isNotEmpty) 'password': password,
      });
      return response.data;
    } on DioException catch (e) {
      final errorMsg = e.response?.data?['error'] ?? e.message ?? '未知错误';
      throw Exception(errorMsg);
    }
  }

  Future<List<Map<String, dynamic>>> listMyRooms() async {
    final response = await _dio.get('/owner/rooms');
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return <Map<String, dynamic>>[];
  }

  // ===== 资金 =====

  Future<Map<String, dynamic>> createFundRequest({
    required String type,
    required double amount,
    String? remark,
  }) async {
    final response = await _dio.post('/fund-requests', data: {
      'type': type,
      'amount': amount,
      if (remark != null) 'remark': remark,
    });
    return response.data;
  }

  Future<List<Map<String, dynamic>>> listFundRequests({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/fund-requests', queryParameters: {
      'page': page,
      'page_size': pageSize,
    });
    final data = response.data;
    if (data is Map<String, dynamic>) {
      final items = data['items'];
      if (items is List) {
        return items.cast<Map<String, dynamic>>();
      }
    }
    return <Map<String, dynamic>>[];
  }

  // ===== Owner 资金审批 =====

  Future<List<Map<String, dynamic>>> listOwnerFundRequests({int page = 1, int pageSize = 20, String? status}) async {
    final response = await _dio.get('/owner/fund-requests', queryParameters: {
      'page': page,
      'page_size': pageSize,
      if (status != null) 'status': status,
    });
    final data = response.data;
    if (data is Map<String, dynamic>) {
      final items = data['items'];
      if (items is List) {
        return items.cast<Map<String, dynamic>>();
      }
    }
    return <Map<String, dynamic>>[];
  }

  Future<void> processOwnerFundRequest(int requestId, {required bool approved, String? remark}) async {
    await _dio.post('/owner/fund-requests/$requestId/process', data: {
      'approved': approved,
      if (remark != null) 'remark': remark,
    });
  }

  Future<List<Map<String, dynamic>>> listTransactions({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/transactions', queryParameters: {
      'page': page,
      'page_size': pageSize,
    });
    final data = response.data;
    if (data is Map<String, dynamic>) {
      final items = data['items'];
      if (items is List) {
        return items.cast<Map<String, dynamic>>();
      }
    }
    return <Map<String, dynamic>>[];
  }

  Future<Map<String, dynamic>> getFundSummary() async {
    final response = await _dio.get('/fund-summary');
    return response.data;
  }

  // ===== 好友 =====

  Future<List<Map<String, dynamic>>> getFriendList() async {
    final response = await _dio.get('/friends');
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return <Map<String, dynamic>>[];
  }

  Future<List<Map<String, dynamic>>> getPendingFriendRequests() async {
    final response = await _dio.get('/friends/requests');
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return <Map<String, dynamic>>[];
  }

  Future<Map<String, dynamic>> sendFriendRequest(int toUserId) async {
    final response = await _dio.post('/friends/request', data: {
      'to_user_id': toUserId,
    });
    return response.data;
  }

  Future<void> acceptFriendRequest(int requestId) async {
    await _dio.post('/friends/accept/$requestId');
  }

  Future<void> rejectFriendRequest(int requestId) async {
    await _dio.post('/friends/reject/$requestId');
  }

  Future<void> removeFriend(int friendId) async {
    await _dio.delete('/friends/$friendId');
  }

  // ===== 邀请 =====

  Future<List<Map<String, dynamic>>> getPendingInvitations() async {
    final response = await _dio.get('/invitations');
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return <Map<String, dynamic>>[];
  }

  Future<Map<String, dynamic>> sendRoomInvitation(int roomId, int toUserId) async {
    final response = await _dio.post('/rooms/$roomId/invite', data: {
      'to_user_id': toUserId,
    });
    return response.data;
  }

  Future<void> acceptInvitation(int invitationId) async {
    await _dio.post('/invitations/$invitationId/accept');
  }

  Future<void> declineInvitation(int invitationId) async {
    await _dio.post('/invitations/$invitationId/decline');
  }

  Future<Map<String, dynamic>> createInviteLink(int roomId, {int? maxUses}) async {
    final response = await _dio.post('/rooms/$roomId/invite-link', data: {
      if (maxUses != null) 'max_uses': maxUses,
    });
    return response.data;
  }

  Future<Map<String, dynamic>> joinByInviteLink(String code) async {
    final response = await _dio.post('/invite/$code/join');
    return response.data;
  }

  // ===== 钱包 =====

  Future<Map<String, dynamic>> getWallet() async {
    final response = await _dio.get('/wallet');
    return response.data;
  }

  Future<List<Map<String, dynamic>>> getWalletTransactions({int page = 1, int pageSize = 20}) async {
    final response = await _dio.get('/wallet/transactions', queryParameters: {
      'page': page,
      'page_size': pageSize,
    });
    final data = response.data;
    if (data is Map<String, dynamic>) {
      final items = data['items'];
      if (items is List) {
        return items.cast<Map<String, dynamic>>();
      }
    }
    return <Map<String, dynamic>>[];
  }

  Future<Map<String, dynamic>> getWalletEarnings() async {
    final response = await _dio.get('/wallet/earnings');
    return response.data;
  }

  Future<void> transferEarnings(double amount) async {
    await _dio.post('/wallet/transfer-earnings', data: {
      'amount': amount,
    });
  }

  // ===== 用户偏好 =====

  Future<void> updateLanguage(String language) async {
    await _dio.put('/me/language', data: {
      'language': language,
    });
  }

  // ===== 主题 =====

  Future<List<Map<String, dynamic>>> getAllThemes() async {
    final response = await _dio.get('/themes');
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return <Map<String, dynamic>>[];
  }

  Future<Map<String, dynamic>> getRoomTheme(int roomId) async {
    final response = await _dio.get('/rooms/$roomId/theme');
    return response.data;
  }

  Future<Map<String, dynamic>> updateRoomTheme(int roomId, String themeName) async {
    final response = await _dio.put('/owner/rooms/$roomId/theme', data: {
      'theme_name': themeName,
    });
    return response.data;
  }

  // ===== 排行榜 =====

  Future<List<Map<String, dynamic>>> getLeaderboard(String type) async {
    final response = await _dio.get('/leaderboard', queryParameters: {
      'type': type,
    });
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    if (data is Map<String, dynamic>) {
      final items = data['items'];
      if (items is List) {
        return items.cast<Map<String, dynamic>>();
      }
    }
    return <Map<String, dynamic>>[];
  }

  // ===== 观战 =====

  Future<Map<String, dynamic>> spectateRoom(int roomId) async {
    final response = await _dio.post('/rooms/$roomId/spectate');
    return response.data;
  }

  Future<Map<String, dynamic>> switchToParticipant(int roomId) async {
    final response = await _dio.post('/rooms/$roomId/switch-to-participant');
    return response.data;
  }

  // ===== 用户搜索 =====

  Future<List<Map<String, dynamic>>> searchUsers(String query) async {
    final response = await _dio.get('/users/search', queryParameters: {
      'q': query,
    });
    final data = response.data;
    if (data is List) {
      return data.cast<Map<String, dynamic>>();
    }
    return <Map<String, dynamic>>[];
  }

  // ===== 通用 GET 方法 =====

  Future<Map<String, dynamic>> get(String path) async {
    final response = await _dio.get(path);
    return response.data;
  }
}

// 使用单例模式确保 ApiClient 只有一个实例
class ApiClientSingleton {
  static final ApiClient instance = ApiClient();
}

final apiClientProvider = Provider<ApiClient>((ref) {
  return ApiClientSingleton.instance;
});
