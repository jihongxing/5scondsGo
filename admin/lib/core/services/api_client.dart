import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// 注意：和玩家 App 共用一套后端接口，baseUrl 保持一致
const String baseUrl = 'http://localhost:8080/api';

class AdminApiClient {
  late final Dio _dio;
  String? _token;

  AdminApiClient() {
    _dio = Dio(
      BaseOptions(
        baseUrl: baseUrl,
        connectTimeout: const Duration(seconds: 10),
        receiveTimeout: const Duration(seconds: 10),
        headers: const {
          'Content-Type': 'application/json',
        },
      ),
    );

    _dio.interceptors.add(
      InterceptorsWrapper(
        onRequest: (options, handler) {
          if (_token != null) {
            options.headers['Authorization'] = 'Bearer $_token';
          }
          return handler.next(options);
        },
        onError: (error, handler) {
          if (error.response?.statusCode == 401) {
            // Token 失效，清理本地登录状态
            clearToken();
          }
          return handler.next(error);
        },
      ),
    );
  }

  Future<void> init() async {
    final prefs = await SharedPreferences.getInstance();
    _token = prefs.getString('admin_auth_token');
  }

  Future<void> setToken(String token) async {
    _token = token;
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('admin_auth_token', token);
  }

  Future<void> clearToken() async {
    _token = null;
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('admin_auth_token');
  }

  String? get token => _token;

  // ===== 认证 =====

  Future<Map<String, dynamic>> login({
    required String username,
    required String password,
  }) async {
    final resp = await _dio.post('/auth/login', data: {
      'username': username,
      'password': password,
    });

    final data = resp.data as Map<String, dynamic>;
    final token = data['token'] as String?;
    if (token != null) {
      await setToken(token);
    }
    return data;
  }

  Future<Map<String, dynamic>> getMe() async {
    final resp = await _dio.get('/me');
    return resp.data as Map<String, dynamic>;
  }

  // ===== 管理员：资金对账 / 汇总 =====

  /// GET /api/admin/reports/balance-check
  /// 返回 {"check": ConservationCheck, "summary": FundSummary}
  Future<Map<String, dynamic>> getBalanceCheckReport() async {
    final resp = await _dio.get('/admin/reports/balance-check');
    return resp.data as Map<String, dynamic>;
  }

  /// 最近资金申请（使用通用 /fund-requests，管理员能看到所有用户）
  Future<Map<String, dynamic>> listFundRequests({
    int page = 1,
    int pageSize = 20,
    String? status, // pending / approved / rejected
    String? type, // deposit / withdraw / owner_deposit / ...
  }) async {
    final resp = await _dio.get(
      '/fund-requests',
      queryParameters: {
        'page': page,
        'page_size': pageSize,
        if (status != null) 'status': status,
        if (type != null) 'type': type,
      },
    );
    return (resp.data as Map<String, dynamic>);
  }

  /// 审批资金申请：POST /api/admin/fund-requests/:id/process
  Future<void> processFundRequest({
    required int id,
    required bool approved,
    String? remark,
  }) async {
    await _dio.post('/admin/fund-requests/$id/process', data: {
      'approved': approved,
      'remark': remark ?? '',
    });
  }

  /// 交易流水列表：GET /api/transactions
  Future<Map<String, dynamic>> listTransactions({
    int page = 1,
    int pageSize = 20,
    String? type, // 见 TransactionType
    int? userId,
  }) async {
    final resp = await _dio.get(
      '/transactions',
      queryParameters: {
        'page': page,
        'page_size': pageSize,
        if (type != null) 'type': type,
        if (userId != null) 'user_id': userId,
      },
    );
    return resp.data as Map<String, dynamic>;
  }

  // ===== 用户 / 房主 =====

  /// 用户列表：GET /api/admin/users
  Future<Map<String, dynamic>> listUsers({
    String? role, // admin / owner / player
    String? search,
    int page = 1,
    int pageSize = 50,
  }) async {
    final resp = await _dio.get(
      '/admin/users',
      queryParameters: {
        'page': page,
        'page_size': pageSize,
        if (role != null && role != 'all') 'role': role,
        if (search != null && search.isNotEmpty) 'search': search,
      },
    );
    return resp.data as Map<String, dynamic>;
  }

  /// 创建房主：POST /api/admin/owners
  Future<Map<String, dynamic>> createOwner({
    required String username,
    required String password,
  }) async {
    final resp = await _dio.post(
      '/admin/owners',
      data: {
        'username': username,
        'password': password,
      },
    );
    return resp.data as Map<String, dynamic>;
  }

  // ===== 房间 =====

  /// 房间列表：GET /api/rooms
  Future<Map<String, dynamic>> listRooms({
    int page = 1,
    int pageSize = 50,
    String? status, // active / paused / locked
    int? ownerId,
  }) async {
    final resp = await _dio.get(
      '/rooms',
      queryParameters: {
        'page': page,
        'page_size': pageSize,
        if (status != null && status != 'all') 'status': status,
        if (ownerId != null) 'owner_id': ownerId,
      },
    );
    return resp.data as Map<String, dynamic>;
  }

  /// 更新房间状态（管理端）：PUT /api/admin/rooms/:id/status
  Future<void> updateRoomStatus({
    required int id,
    required String status, // active / paused / locked
  }) async {
    await _dio.put(
      '/admin/rooms/$id/status',
      data: {'status': status},
    );
  }

  // ===== 风控标记 =====

  /// 风控标记列表：GET /api/admin/risk-flags
  Future<Map<String, dynamic>> listRiskFlags({
    int page = 1,
    int pageSize = 20,
    String? status, // pending / confirmed / dismissed
    String? flagType, // consecutive_wins / high_win_rate / multi_account / large_transaction
    int? userId,
  }) async {
    final resp = await _dio.get(
      '/admin/risk-flags',
      queryParameters: {
        'page': page,
        'page_size': pageSize,
        if (status != null) 'status': status,
        if (flagType != null) 'flag_type': flagType,
        if (userId != null) 'user_id': userId,
      },
    );
    return resp.data as Map<String, dynamic>;
  }

  /// 获取风控标记详情：GET /api/admin/risk-flags/:id
  Future<Map<String, dynamic>> getRiskFlag(int id) async {
    final resp = await _dio.get('/admin/risk-flags/$id');
    return resp.data as Map<String, dynamic>;
  }

  /// 审核风控标记：POST /api/admin/risk-flags/:id/review
  Future<void> reviewRiskFlag(int id, String action, {String? remark}) async {
    await _dio.post('/admin/risk-flags/$id/review', data: {
      'action': action, // confirm / dismiss
      if (remark != null) 'remark': remark,
    });
  }

  // ===== 告警 =====

  /// 告警列表：GET /api/admin/alerts
  Future<Map<String, dynamic>> listAlerts({
    int page = 1,
    int pageSize = 20,
    String? status, // active / acknowledged
    String? severity, // info / warning / critical
    String? alertType,
  }) async {
    final resp = await _dio.get(
      '/admin/alerts',
      queryParameters: {
        'page': page,
        'page_size': pageSize,
        if (status != null) 'status': status,
        if (severity != null) 'severity': severity,
        if (alertType != null) 'alert_type': alertType,
      },
    );
    return resp.data as Map<String, dynamic>;
  }

  /// 获取告警详情：GET /api/admin/alerts/:id
  Future<Map<String, dynamic>> getAlert(int id) async {
    final resp = await _dio.get('/admin/alerts/$id');
    return resp.data as Map<String, dynamic>;
  }

  /// 确认告警：POST /api/admin/alerts/:id/acknowledge
  Future<void> acknowledgeAlert(int id, {String? remark}) async {
    await _dio.post('/admin/alerts/$id/acknowledge', data: {
      if (remark != null) 'remark': remark,
    });
  }

  /// 获取告警摘要：GET /api/admin/alerts/summary
  Future<Map<String, dynamic>> getAlertSummary() async {
    final resp = await _dio.get('/admin/alerts/summary');
    return resp.data as Map<String, dynamic>;
  }

  // ===== 监控指标 =====

  /// 获取实时指标：GET /api/admin/metrics/realtime
  Future<Map<String, dynamic>> getRealtimeMetrics() async {
    final resp = await _dio.get('/admin/metrics/realtime');
    return resp.data as Map<String, dynamic>;
  }

  /// 获取历史指标：GET /api/admin/metrics/history
  Future<Map<String, dynamic>> getHistoricalMetrics({
    String timeRange = '1h', // 1h, 24h, 7d, 30d
  }) async {
    final resp = await _dio.get(
      '/admin/metrics/history',
      queryParameters: {'time_range': timeRange},
    );
    return resp.data as Map<String, dynamic>;
  }
}

// 使用单例模式确保 token 不会丢失
final _adminApiClientInstance = AdminApiClient();

final adminApiClientProvider = Provider<AdminApiClient>((ref) {
  return _adminApiClientInstance;
});
