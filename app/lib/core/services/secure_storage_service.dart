import 'package:flutter/foundation.dart' show kIsWeb;
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:shared_preferences/shared_preferences.dart';

/// 安全存储服务
/// 在移动端使用 flutter_secure_storage 加密存储
/// 在 Web 端回退到 SharedPreferences（Web 不支持 secure storage）
class SecureStorageService {
  static final SecureStorageService _instance = SecureStorageService._();
  static SecureStorageService get instance => _instance;

  SecureStorageService._();

  // Android 配置：使用 AES 加密
  static const _androidOptions = AndroidOptions(
    encryptedSharedPreferences: true,
  );

  // iOS 配置
  static const _iosOptions = IOSOptions(
    accessibility: KeychainAccessibility.first_unlock_this_device,
  );

  final FlutterSecureStorage _secureStorage = const FlutterSecureStorage(
    aOptions: _androidOptions,
    iOptions: _iosOptions,
  );

  /// 存储 Token
  Future<void> saveToken(String token) async {
    if (kIsWeb) {
      // Web 端使用 SharedPreferences
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString('auth_token', token);
    } else {
      // 移动端使用加密存储
      await _secureStorage.write(key: 'auth_token', value: token);
    }
  }

  /// 读取 Token
  Future<String?> getToken() async {
    if (kIsWeb) {
      final prefs = await SharedPreferences.getInstance();
      return prefs.getString('auth_token');
    } else {
      return await _secureStorage.read(key: 'auth_token');
    }
  }

  /// 删除 Token
  Future<void> deleteToken() async {
    if (kIsWeb) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove('auth_token');
    } else {
      await _secureStorage.delete(key: 'auth_token');
    }
  }

  /// 存储通用数据
  Future<void> write(String key, String value) async {
    if (kIsWeb) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.setString(key, value);
    } else {
      await _secureStorage.write(key: key, value: value);
    }
  }

  /// 读取通用数据
  Future<String?> read(String key) async {
    if (kIsWeb) {
      final prefs = await SharedPreferences.getInstance();
      return prefs.getString(key);
    } else {
      return await _secureStorage.read(key: key);
    }
  }

  /// 删除通用数据
  Future<void> delete(String key) async {
    if (kIsWeb) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove(key);
    } else {
      await _secureStorage.delete(key: key);
    }
  }

  /// 清除所有安全存储数据
  Future<void> deleteAll() async {
    if (kIsWeb) {
      final prefs = await SharedPreferences.getInstance();
      await prefs.remove('auth_token');
    } else {
      await _secureStorage.deleteAll();
    }
  }
}
