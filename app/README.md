# 5SecondsGo - Flutter 移动端

5SecondsGo 游戏的 Flutter 移动客户端，支持 iOS 和 Android 平台。

## 功能特性

- 🎮 实时多人押注游戏
- 👀 观战模式
- 💬 房间聊天和表情
- 👥 好友系统
- 📨 邀请功能
- 💼 钱包管理
- 📜 游戏记录
- 🎨 多主题支持
- 🌐 多语言 (中/英/日/韩)

## 环境要求

- Flutter 3.16+
- Dart 3.2+
- Android Studio / Xcode

## 快速开始

### 1. 安装依赖

```bash
flutter pub get
```

### 2. 配置服务器地址

编辑 `lib/core/services/api_client.dart`:

```dart
static const String baseUrl = 'http://your-server:8080';
static const String wsUrl = 'ws://your-server:8080/ws';
```

### 3. 运行应用

```bash
# 开发模式
flutter run

# 指定设备
flutter run -d <device_id>
```

## 项目结构

```
lib/
├── core/                    # 核心模块
│   ├── providers/          # 状态管理 (Riverpod)
│   │   └── locale_provider.dart
│   ├── services/           # 服务层
│   │   ├── api_client.dart    # HTTP 客户端
│   │   ├── ws_client.dart     # WebSocket 客户端
│   │   └── audio_service.dart # 音效服务
│   ├── router/             # 路由配置 (GoRouter)
│   │   └── app_router.dart
│   └── theme/              # 主题配置
│       └── app_theme.dart
│
├── features/                # 功能模块
│   ├── auth/               # 认证模块
│   │   ├── presentation/
│   │   │   └── pages/
│   │   │       ├── login_page.dart
│   │   │       └── register_page.dart
│   │   └── providers/
│   │       └── auth_provider.dart
│   │
│   ├── room/               # 房间模块
│   │   ├── presentation/
│   │   │   ├── pages/
│   │   │   │   ├── room_page.dart
│   │   │   │   └── create_room_page.dart
│   │   │   └── widgets/
│   │   │       ├── chat_widget.dart
│   │   │       ├── emoji_picker.dart
│   │   │       └── ...
│   │   └── providers/
│   │
│   ├── wallet/             # 钱包模块
│   ├── friends/            # 好友模块
│   ├── game_history/       # 游戏记录
│   ├── profile/            # 个人中心
│   ├── home/               # 首页
│   └── invite/             # 邀请功能
│
├── l10n/                    # 国际化
│   ├── app_en.arb          # 英文
│   ├── app_zh.arb          # 简体中文
│   ├── app_zh_TW.arb       # 繁体中文
│   ├── app_ja.arb          # 日文
│   └── app_ko.arb          # 韩文
│
└── main.dart               # 入口文件
```

## 主要依赖

| 包名 | 用途 |
|------|------|
| flutter_riverpod | 状态管理 |
| go_router | 路由管理 |
| dio | HTTP 客户端 |
| web_socket_channel | WebSocket |
| flutter_screenutil | 屏幕适配 |
| shared_preferences | 本地存储 |
| audioplayers | 音效播放 |

## 构建发布

### Android

```bash
# 构建 APK
flutter build apk --release

# 构建 App Bundle
flutter build appbundle --release
```

输出位置: `build/app/outputs/`

### iOS

```bash
# 需要 macOS 和 Xcode
flutter build ios --release
```

## 多语言支持

支持的语言:
- English (en)
- 简体中文 (zh)
- 繁體中文 (zh_TW)
- 日本語 (ja)
- 한국어 (ko)

添加新语言:
1. 在 `lib/l10n/` 创建 `app_<locale>.arb`
2. 在 `lib/core/providers/locale_provider.dart` 添加 Locale
3. 运行 `flutter gen-l10n`

## 主题定制

应用支持 5 种内置主题:
- Classic (经典)
- Neon (霓虹)
- Ocean (海洋)
- Forest (森林)
- Luxury (奢华)

主题配置在 `lib/core/theme/app_theme.dart`

## 调试

```bash
# 查看日志
flutter logs

# 性能分析
flutter run --profile
```

## 常见问题

### WebSocket 连接失败
- 检查服务器地址配置
- 确保服务器已启动
- 检查网络连接

### 音效不播放
- 确保 `assets/sounds/` 目录存在音效文件
- 检查 `pubspec.yaml` 中的资源配置

### 国际化不生效
- 运行 `flutter gen-l10n`
- 重启应用
