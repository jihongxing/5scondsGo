import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';

import '../../../core/services/api_client.dart';

class AdminAuthState {
  final bool isLoggedIn;
  final String? token;
  final int? userId;
  final String? username;
  final String? role;

  const AdminAuthState({
    this.isLoggedIn = false,
    this.token,
    this.userId,
    this.username,
    this.role,
  });

  AdminAuthState copyWith({
    bool? isLoggedIn,
    String? token,
    int? userId,
    String? username,
    String? role,
  }) {
    return AdminAuthState(
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
      token: token ?? this.token,
      userId: userId ?? this.userId,
      username: username ?? this.username,
      role: role ?? this.role,
    );
  }
}

class AdminAuthNotifier extends StateNotifier<AdminAuthState> {
  final AdminApiClient _apiClient;

  AdminAuthNotifier(this._apiClient) : super(const AdminAuthState()) {
    _loadFromStorage();
  }

  Future<void> _loadFromStorage() async {
    await _apiClient.init();
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('admin_auth_token');
    final userId = prefs.getInt('admin_user_id');
    final username = prefs.getString('admin_username');
    final role = prefs.getString('admin_role');

    if (token != null && role != null) {
      state = AdminAuthState(
        isLoggedIn: true,
        token: token,
        userId: userId,
        username: username,
        role: role,
      );
    }
  }

  Future<void> login({
    required String username,
    required String password,
  }) async {
    final resp = await _apiClient.login(username: username, password: password);
    final token = resp['token'] as String;
    final user = resp['user'] as Map<String, dynamic>;
    final userId = user['id'] as int;
    final role = user['role'] as String;

    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt('admin_user_id', userId);
    await prefs.setString('admin_username', username);
    await prefs.setString('admin_role', role);

    state = AdminAuthState(
      isLoggedIn: true,
      token: token,
      userId: userId,
      username: username,
      role: role,
    );
  }

  Future<void> logout() async {
    await _apiClient.clearToken();
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('admin_user_id');
    await prefs.remove('admin_username');
    await prefs.remove('admin_role');

    state = const AdminAuthState();
  }
}

final adminAuthProvider =
    StateNotifierProvider<AdminAuthNotifier, AdminAuthState>((ref) {
  final api = ref.watch(adminApiClientProvider);
  return AdminAuthNotifier(api);
});
