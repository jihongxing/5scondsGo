import 'dart:ui' as ui;
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:shared_preferences/shared_preferences.dart';
import '../services/api_client.dart';

/// 支持的语言列表
const supportedLocales = [
  Locale('en'),       // English
  Locale('zh'),       // 简体中文
  Locale('zh', 'TW'), // 繁体中文
  Locale('ja'),       // 日本語
  Locale('ko'),       // 한국어
];

/// 用于在应用启动前预加载语言设置
Future<Locale> preloadLocale() async {
  final prefs = await SharedPreferences.getInstance();
  final code = prefs.getString('locale');
  final country = prefs.getString('locale_country');
  
  if (code != null) {
    final locale = country != null ? Locale(code, country) : Locale(code);
    return _findBestMatchStatic(locale);
  }
  
  // 没有保存的语言，使用系统语言
  final systemLocale = ui.PlatformDispatcher.instance.locale;
  return _findBestMatchStatic(systemLocale);
}

/// 静态方法：查找最佳匹配的语言
Locale _findBestMatchStatic(Locale locale) {
  // 精确匹配
  for (final supported in supportedLocales) {
    if (supported.languageCode == locale.languageCode &&
        supported.countryCode == locale.countryCode) {
      return supported;
    }
  }
  
  // 语言匹配（忽略国家/地区）
  for (final supported in supportedLocales) {
    if (supported.languageCode == locale.languageCode) {
      return supported;
    }
  }
  
  // 回退到英语
  return const Locale('en');
}

/// 预加载的语言设置 Provider
final preloadedLocaleProvider = FutureProvider<Locale>((ref) async {
  return preloadLocale();
});

final localeProvider = StateNotifierProvider<LocaleNotifier, Locale>((ref) {
  final apiClient = ref.watch(apiClientProvider);
  // 尝试获取预加载的语言，如果还没加载完成则使用系统默认
  final preloadedLocale = ref.watch(preloadedLocaleProvider).valueOrNull;
  return LocaleNotifier(apiClient, preloadedLocale);
});

class LocaleNotifier extends StateNotifier<Locale> {
  final ApiClient _apiClient;
  
  LocaleNotifier(this._apiClient, Locale? initialLocale) 
      : super(initialLocale ?? _getDefaultLocale()) {
    // 如果没有预加载的语言，则异步加载
    if (initialLocale == null) {
      _loadLocale();
    }
  }

  static const _key = 'locale';
  static const _countryKey = 'locale_country';

  /// 获取默认语言（基于系统语言）
  static Locale _getDefaultLocale() {
    final systemLocale = ui.PlatformDispatcher.instance.locale;
    return _findBestMatchStatic(systemLocale);
  }

  /// 查找最佳匹配的语言（实例方法，调用静态方法）
  static Locale _findBestMatch(Locale locale) {
    return _findBestMatchStatic(locale);
  }

  Future<void> _loadLocale() async {
    final prefs = await SharedPreferences.getInstance();
    final code = prefs.getString(_key);
    final country = prefs.getString(_countryKey);
    
    if (code != null) {
      final locale = country != null ? Locale(code, country) : Locale(code);
      state = _findBestMatch(locale);
    }
  }

  Future<void> setLocale(Locale locale, {bool syncToServer = true}) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString(_key, locale.languageCode);
    if (locale.countryCode != null) {
      await prefs.setString(_countryKey, locale.countryCode!);
    } else {
      await prefs.remove(_countryKey);
    }
    state = _findBestMatch(locale);
    
    // 同步到服务器（如果已登录且需要同步）
    if (syncToServer) {
      _syncToServer(locale);
    }
  }
  
  /// 同步语言偏好到服务器
  Future<void> _syncToServer(Locale locale) async {
    if (_apiClient.token == null) return;
    
    try {
      // 转换为服务器端语言代码格式
      String langCode;
      if (locale.countryCode != null) {
        langCode = '${locale.languageCode}-${locale.countryCode}';
      } else {
        langCode = locale.languageCode;
      }
      await _apiClient.updateLanguage(langCode);
    } catch (e) {
      // 同步失败不影响本地设置
    }
  }

  /// 切换到下一个语言
  void cycleLocale() {
    final currentIndex = supportedLocales.indexWhere(
      (l) => l.languageCode == state.languageCode && 
             l.countryCode == state.countryCode,
    );
    final nextIndex = (currentIndex + 1) % supportedLocales.length;
    setLocale(supportedLocales[nextIndex]); // 用户主动切换，需要同步到服务器
  }

  /// 重置为系统语言
  Future<void> resetToSystemLocale() async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.remove(_key);
    await prefs.remove(_countryKey);
    state = _getDefaultLocale();
  }

  /// 获取语言显示名称
  static String getDisplayName(Locale locale) {
    switch ('${locale.languageCode}_${locale.countryCode ?? ''}') {
      case 'en_':
        return 'English';
      case 'zh_':
        return '简体中文';
      case 'zh_TW':
        return '繁體中文';
      case 'ja_':
        return '日本語';
      case 'ko_':
        return '한국어';
      default:
        return locale.languageCode;
    }
  }
}
