// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for Chinese (`zh`).
class AppLocalizationsZh extends AppLocalizations {
  AppLocalizationsZh([String locale = 'zh']) : super(locale);

  @override
  String get appTitle => '5秒GO 管理后台';

  @override
  String get navDashboard => '概览';

  @override
  String get navUsers => '用户';

  @override
  String get navRooms => '房间';

  @override
  String get navFunds => '资金';

  @override
  String get navLogout => '退出登录';

  @override
  String get loginTitle => '5秒GO 管理后台';

  @override
  String get loginUsernameLabel => '用户名';

  @override
  String get loginUsernameRequired => '请输入用户名';

  @override
  String get loginPasswordLabel => '密码';

  @override
  String get loginPasswordTooShort => '密码至少需要 6 位';

  @override
  String get loginButton => '登录';

  @override
  String get dashboardTitle => '概览';

  @override
  String get dashboardCardTotalUsers => '用户总数';

  @override
  String get dashboardCardActiveRooms => '活跃房间';

  @override
  String get dashboardCardOnlinePlayers => '在线玩家';

  @override
  String get dashboardCardPlatformBalance => '平台总余额';

  @override
  String get dashboardFundCheckTitle => '资金平衡校验';

  @override
  String dashboardFundCheckPlayerBalance(Object amount) {
    return '玩家总余额：$amount';
  }

  @override
  String dashboardFundCheckCustodyQuota(Object amount) {
    return '房主总余额：$amount';
  }

  @override
  String dashboardFundCheckMargin(Object amount) {
    return '房主总收益：$amount';
  }

  @override
  String dashboardFundCheckPlatformProfit(Object amount) {
    return '平台总收益：$amount';
  }

  @override
  String get dashboardFundCheckBalanced => '✓ 已平衡';

  @override
  String get dashboardRecentFundRequests => '最近资金申请';

  @override
  String dashboardRecentFundTitle(Object type, Object username) {
    return '$username - $type';
  }

  @override
  String get dashboardRecentFundTypeDeposit => '充值';

  @override
  String get dashboardRecentFundTypeWithdraw => '提现';

  @override
  String get dashboardRecentFundStatusPending => '待处理';

  @override
  String get dashboardRecentFundStatusApproved => '已通过';

  @override
  String get dashboardRecentFundStatusRejected => '已拒绝';

  @override
  String get dashboardAdminInviteTitle => '管理员账号';

  @override
  String dashboardAdminInviteCodeLabel(Object code) {
    return '邀请码：$code';
  }

  @override
  String get fundsTitle => '资金';

  @override
  String get fundsTabRequests => '资金申请';

  @override
  String get fundsTabTransactions => '资金流水';

  @override
  String get fundsFilterAllStatus => '全部状态';

  @override
  String get fundsFilterStatusPending => '待处理';

  @override
  String get fundsFilterStatusApproved => '已通过';

  @override
  String get fundsFilterStatusRejected => '已拒绝';

  @override
  String get fundsFilterAllTypes => '全部类型';

  @override
  String get fundsFilterTypeDeposit => '充值';

  @override
  String get fundsFilterTypeWithdraw => '提现';

  @override
  String get fundsFilterTypeMargin => '保证金';

  @override
  String get fundsStatusPending => '待处理';

  @override
  String get fundsStatusApproved => '已通过';

  @override
  String get fundsStatusRejected => '已拒绝';

  @override
  String get fundsSearchUserHint => '按用户搜索...';

  @override
  String get fundsTxTypeGameBet => '游戏下注';

  @override
  String get fundsTxTypeGameWin => '游戏赢钱';

  @override
  String get fundsTxTypeDeposit => '充值';

  @override
  String get fundsTxTypeWithdraw => '提现';

  @override
  String get roomsTitle => '房间';

  @override
  String get roomsSearchHint => '搜索房间...';

  @override
  String get roomsFilterAllStatus => '全部状态';

  @override
  String get roomsFilterStatusActive => '进行中';

  @override
  String get roomsFilterStatusPaused => '已暂停';

  @override
  String get roomsFilterStatusLocked => '已锁定';

  @override
  String roomsItemTitle(Object roomName) {
    return '$roomName';
  }

  @override
  String roomsItemSubtitle(Object bet, Object ownerName, Object players) {
    return '房主：$ownerName · 玩家：$players/20 · 底注：$bet';
  }

  @override
  String get roomsStatusActive => '进行中';

  @override
  String get roomsStatusPaused => '已暂停';

  @override
  String get roomsStatusLocked => '已锁定';

  @override
  String get usersTitle => '用户';

  @override
  String get usersCreateOwner => '创建房主';

  @override
  String get usersSearchHint => '搜索用户...';

  @override
  String get usersFilterAllRoles => '全部角色';

  @override
  String get usersFilterRolePlayer => '玩家';

  @override
  String get usersFilterRoleOwner => '房主';

  @override
  String get usersFilterRoleAdmin => '管理员';

  @override
  String usersItemTitle(Object username) {
    return '$username';
  }

  @override
  String usersItemSubtitleOwner(Object code) {
    return '房主 · 邀请码：$code';
  }

  @override
  String usersItemSubtitlePlayer(Object ownerName) {
    return '玩家 · 邀请人：$ownerName';
  }

  @override
  String get usersDialogCreateOwnerTitle => '创建房主';

  @override
  String get usersDialogUsernameLabel => '用户名';

  @override
  String get usersDialogPasswordLabel => '密码';

  @override
  String get usersDialogCancel => '取消';

  @override
  String get usersDialogCreate => '创建';

  @override
  String get navMonitoring => '监控';

  @override
  String get navRiskFlags => '风控标记';

  @override
  String get navAlerts => '告警';

  @override
  String get monitoringTitle => '监控仪表盘';

  @override
  String get monitoringRealtimeMetrics => '实时指标';

  @override
  String get monitoringPerformanceMetrics => '性能指标';

  @override
  String get monitoringHistoricalTrends => '历史趋势';

  @override
  String get monitoringOnlinePlayers => '在线玩家';

  @override
  String get monitoringActiveRooms => '活跃房间';

  @override
  String get monitoringGamesPerMinute => '每分钟游戏数';

  @override
  String get monitoringDailyActiveUsers => '日活用户';

  @override
  String get monitoringDailyVolume => '日交易量';

  @override
  String get monitoringPlatformRevenue => '平台收入';

  @override
  String get monitoringApiLatencyP95 => 'API延迟P95';

  @override
  String get monitoringWsLatencyP95 => 'WS延迟P95';

  @override
  String get monitoringDbLatencyP95 => 'DB延迟P95';

  @override
  String get refresh => '刷新';
}
