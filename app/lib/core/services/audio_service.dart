import 'package:audioplayers/audioplayers.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// 音效服务
class AudioService {
  final AudioPlayer _coinPlayer = AudioPlayer();
  final AudioPlayer _winPlayer = AudioPlayer();
  
  bool _initialized = false;
  bool _enabled = true;

  /// 初始化音效
  Future<void> init() async {
    if (_initialized) return;
    
    // 预加载音效
    await _coinPlayer.setSource(AssetSource('sounds/coin.mp3'));
    await _coinPlayer.setReleaseMode(ReleaseMode.stop);
    
    _initialized = true;
  }

  /// 播放金币音效（下注时）
  Future<void> playCoinSound() async {
    if (!_enabled) return;
    
    try {
      await _coinPlayer.stop();
      await _coinPlayer.seek(Duration.zero);
      await _coinPlayer.resume();
    } catch (e) {
      // 忽略播放错误
    }
  }

  /// 播放获胜音效
  Future<void> playWinSound() async {
    if (!_enabled) return;
    
    try {
      await _winPlayer.setSource(AssetSource('sounds/win.mp3'));
      await _winPlayer.resume();
    } catch (e) {
      // 忽略播放错误
    }
  }

  /// 设置音效开关
  void setEnabled(bool enabled) {
    _enabled = enabled;
  }

  /// 获取音效开关状态
  bool get isEnabled => _enabled;

  /// 释放资源
  void dispose() {
    _coinPlayer.dispose();
    _winPlayer.dispose();
  }
}

/// 音效服务 Provider
final audioServiceProvider = Provider<AudioService>((ref) {
  final service = AudioService();
  ref.onDispose(() => service.dispose());
  return service;
});
