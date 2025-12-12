import 'package:flutter/foundation.dart';
import 'package:flutter_dotenv/flutter_dotenv.dart';

/// 应用配置管理
class AppConfig {
  static late final AppConfig _instance;
  static AppConfig get instance => _instance;

  final String apiBaseUrl;
  final String wsBaseUrl;
  final bool isProduction;

  AppConfig._({
    required this.apiBaseUrl,
    required this.wsBaseUrl,
    required this.isProduction,
  });

  /// 初始化配置（在 main.dart 中调用）
  static Future<void> init({String? envFile}) async {
    // 加载环境变量文件
    final file = envFile ?? (kReleaseMode ? '.env.production' : '.env');
    try {
      await dotenv.load(fileName: file);
    } catch (e) {
      // 如果文件不存在，使用默认值
      debugPrint('Warning: $file not found, using defaults');
    }

    final envMode = dotenv.env['ENV_MODE'] ?? 'development';
    final isProduction = envMode == 'production' || kReleaseMode;

    // 根据平台选择默认地址
    String defaultApiUrl;
    String defaultWsUrl;
    
    if (kIsWeb) {
      defaultApiUrl = 'http://localhost:8080/api';
      defaultWsUrl = 'ws://localhost:8080/ws';
    } else {
      // Android 模拟器使用 10.0.2.2 访问宿主机
      defaultApiUrl = 'http://10.0.2.2:8080/api';
      defaultWsUrl = 'ws://10.0.2.2:8080/ws';
    }

    _instance = AppConfig._(
      apiBaseUrl: dotenv.env['API_BASE_URL'] ?? defaultApiUrl,
      wsBaseUrl: dotenv.env['WS_BASE_URL'] ?? defaultWsUrl,
      isProduction: isProduction,
    );

    debugPrint('AppConfig initialized:');
    debugPrint('  API: ${_instance.apiBaseUrl}');
    debugPrint('  WS: ${_instance.wsBaseUrl}');
    debugPrint('  Production: ${_instance.isProduction}');
  }
}
