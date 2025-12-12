# 5SecondsGo - 管理后台

5SecondsGo 游戏的 Flutter Web 管理后台，用于运营管理和监控。

## 功能特性

- 📊 实时监控仪表盘
- 👥 用户管理
- 🏠 房间管理
- 💰 资金审批
- 🛡️ 风控管理
- 🚨 告警中心

## 环境要求

- Flutter 3.16+
- Dart 3.2+
- Chrome 浏览器 (开发)

## 快速开始

### 1. 安装依赖

```bash
flutter pub get
```

### 2. 配置服务器地址

编辑 `lib/core/services/api_client.dart`:

```dart
static const String baseUrl = 'http://your-server:8080';
```

### 3. 运行开发服务器

```bash
flutter run -d chrome
```

## 项目结构

```
lib/
├── core/
│   └── services/
│       └── api_client.dart    # API 客户端
│
├── features/
│   ├── auth/                  # 登录认证
│   │   ├── presentation/
│   │   │   └── pages/
│   │   │       └── login_page.dart
│   │   └── providers/
│   │       └── auth_provider.dart
│   │
│   ├── dashboard/             # 仪表盘
│   │   └── presentation/
│   │       └── pages/
│   │           └── dashboard_page.dart
│   │
│   ├── users/                 # 用户管理
│   │   └── presentation/
│   │       └── pages/
│   │           └── users_page.dart
│   │
│   ├── rooms/                 # 房间管理
│   │   └── presentation/
│   │       └── pages/
│   │           └── rooms_page.dart
│   │
│   ├── funds/                 # 资金管理
│   │   └── presentation/
│   │       └── pages/
│   │           └── funds_page.dart
│   │
│   ├── monitoring/            # 监控中心
│   │   └── presentation/
│   │       └── pages/
│   │           └── monitoring_dashboard_page.dart
│   │
│   ├── risk/                  # 风控管理
│   │   └── presentation/
│   │       └── pages/
│   │           └── risk_flags_page.dart
│   │
│   └── alerts/                # 告警中心
│       └── presentation/
│           └── pages/
│               └── alerts_page.dart
│
├── l10n/                      # 国际化
│   ├── app_en.arb
│   └── app_zh.arb
│
└── main.dart
```

## 功能说明

### 仪表盘
- 实时在线人数
- 活跃房间数
- 今日交易量
- 系统状态概览

### 用户管理
- 用户列表查询
- 用户详情查看
- 余额调整
- 账户状态管理

### 房间管理
- 房间列表
- 房间详情
- 强制关闭房间

### 资金管理
- 充值/提现申请审批
- 交易流水查询
- 资金对账

### 风控管理
- 可疑账户标记
- 异常行为审核
- 风险等级调整

### 告警中心
- 实时告警列表
- 告警处理
- 历史告警查询

## 构建发布

```bash
# 构建 Web 版本
flutter build web --release

# 输出目录
build/web/
```

### 部署到 Nginx

```nginx
server {
    listen 80;
    server_name admin.example.com;
    
    root /var/www/5secondsgo-admin;
    index index.html;
    
    location / {
        try_files $uri $uri/ /index.html;
    }
}
```

## 权限说明

| 角色 | 权限 |
|------|------|
| admin | 全部功能 |
| operator | 用户管理、房间管理、资金审批 |
| viewer | 只读查看 |

## 默认账号

| 用户名 | 密码 | 角色 |
|--------|------|------|
| admin | admin123 | 管理员 |

> ⚠️ 生产环境请修改默认密码

## 开发说明

### 添加新页面

1. 在 `lib/features/` 创建功能目录
2. 创建 `presentation/pages/xxx_page.dart`
3. 在 `main.dart` 添加路由

### API 调用

```dart
final apiClient = ApiClient();

// GET 请求
final response = await apiClient.get('/api/users');

// POST 请求
final response = await apiClient.post('/api/users', data: {...});
```

## 浏览器兼容性

- Chrome 90+ ✅
- Firefox 88+ ✅
- Safari 14+ ✅
- Edge 90+ ✅
