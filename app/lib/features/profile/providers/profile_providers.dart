import 'dart:ui' show Locale;
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../core/services/api_client.dart';
import '../../../core/providers/locale_provider.dart';
import '../../auth/providers/auth_provider.dart';

/// Profile 状态管理 - 主要用于语言切换和登出功能
/// 大部分状态已由 authProvider 和 localeProvider 管理
class ProfileNotifier extends StateNotifier<ProfileState> {
  final Ref _ref;

  ProfileNotifier(this._ref) : super(ProfileState());

  /// 更新语言偏好
  Future<void> updateLanguage(String languageCode) async {
    state = state.copyWith(isUpdating: true);
    try {
      // 更新本地语言设置
      _ref.read(localeProvider.notifier).setLocale(Locale(languageCode));
      state = state.copyWith(isUpdating: false);
    } catch (e) {
      state = state.copyWith(isUpdating: false, error: e.toString());
    }
  }

  /// 登出
  Future<void> logout() async {
    state = state.copyWith(isUpdating: true);
    try {
      await _ref.read(authProvider.notifier).logout();
      state = state.copyWith(isUpdating: false);
    } catch (e) {
      state = state.copyWith(isUpdating: false, error: e.toString());
    }
  }

  void clearError() {
    state = state.copyWith(error: null);
  }
}

class ProfileState {
  final bool isUpdating;
  final String? error;

  ProfileState({
    this.isUpdating = false,
    this.error,
  });

  ProfileState copyWith({
    bool? isUpdating,
    String? error,
  }) {
    return ProfileState(
      isUpdating: isUpdating ?? this.isUpdating,
      error: error,
    );
  }
}

final profileProvider = StateNotifierProvider<ProfileNotifier, ProfileState>((ref) {
  return ProfileNotifier(ref);
});
