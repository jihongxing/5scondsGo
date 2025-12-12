import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../../../core/services/api_client.dart';
import '../../../core/providers/locale_provider.dart';

class AuthState {
  final bool isLoggedIn;
  final String? token;
  final int? userId;
  final String? username;
  final String? role;
  final String? inviteCode; // 房主/管理员的邀请码
  final String? language; // 用户语言偏好

  AuthState({
    this.isLoggedIn = false,
    this.token,
    this.userId,
    this.username,
    this.role,
    this.inviteCode,
    this.language,
  });

  AuthState copyWith({
    bool? isLoggedIn,
    String? token,
    int? userId,
    String? username,
    String? role,
    String? inviteCode,
  }) {
    return AuthState(
      isLoggedIn: isLoggedIn ?? this.isLoggedIn,
      token: token ?? this.token,
      userId: userId ?? this.userId,
      username: username ?? this.username,
      role: role ?? this.role,
      inviteCode: inviteCode ?? this.inviteCode,
    );
  }
}

class AuthNotifier extends StateNotifier<AuthState> {
  final ApiClient _apiClient;
  final LocaleNotifier? _localeNotifier;

  AuthNotifier(this._apiClient, [this._localeNotifier]) : super(AuthState()) {
    _loadAuth();
  }

  Future<void> _loadAuth() async {
    await _apiClient.init();
    final prefs = await SharedPreferences.getInstance();
    final token = prefs.getString('auth_token');
    final userId = prefs.getInt('userId');
    final username = prefs.getString('username');
    final role = prefs.getString('role');
    final inviteCode = prefs.getString('inviteCode');

    if (token != null) {
      state = AuthState(
        isLoggedIn: true,
        token: token,
        userId: userId,
        username: username,
        role: role,
        inviteCode: inviteCode,
      );
    }
  }

  Future<void> loginWithApi({
    required String username,
    required String password,
  }) async {
    final response = await _apiClient.login(
      username: username,
      password: password,
    );

    final token = response['token'] as String;
    final user = response['user'] as Map<String, dynamic>;
    final userId = user['id'] as int;
    final role = user['role'] as String;
    final inviteCode = user['invite_code'] as String?;
    final language = user['language'] as String?;

    final prefs = await SharedPreferences.getInstance();
    await prefs.setInt('userId', userId);
    await prefs.setString('username', username);
    await prefs.setString('role', role);
    if (inviteCode != null) {
      await prefs.setString('inviteCode', inviteCode);
    }

    // 应用服务器端的语言偏好
    // 只有当服务器有用户主动设置的语言时才应用（非默认值 'en'）
    // 这样可以保持用户在本地选择的语言，直到用户主动在设置中更改
    if (language != null && language.isNotEmpty && _localeNotifier != null) {
      // 服务器返回的语言优先，确保多设备同步
      _applyServerLanguage(language);
    }

    state = AuthState(
      isLoggedIn: true,
      token: token,
      userId: userId,
      username: username,
      role: role,
      inviteCode: inviteCode,
      language: language,
    );
  }

  /// 将服务器端语言代码转换为 Locale 并应用
  void _applyServerLanguage(String langCode) {
    if (_localeNotifier == null) return;
    
    Locale locale;
    // 服务器端格式: en, zh-CN, zh-TW, ja, ko
    if (langCode.contains('-')) {
      final parts = langCode.split('-');
      locale = Locale(parts[0], parts[1]);
    } else if (langCode == 'zh') {
      // zh 默认为简体中文
      locale = const Locale('zh');
    } else {
      locale = Locale(langCode);
    }
    
    // 从服务器应用语言时不需要再同步回服务器
    _localeNotifier!.setLocale(locale, syncToServer: false);
  }

  Future<void> registerWithApi({
    required String username,
    required String password,
    String? inviteCode,
    String? role,
  }) async {
    await _apiClient.register(
      username: username,
      password: password,
      inviteCode: inviteCode,
      role: role,
    );
    // 注册成功后自动登录
    await loginWithApi(username: username, password: password);
  }

  Future<void> logout() async {
    await _apiClient.clearToken();
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove('userId');
    await prefs.remove('username');
    await prefs.remove('role');
    await prefs.remove('inviteCode');

    state = AuthState();
  }
}

final authProvider = StateNotifierProvider<AuthNotifier, AuthState>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  final localeNotifier = ref.watch(localeProvider.notifier);
  return AuthNotifier(apiClient, localeNotifier);
});
