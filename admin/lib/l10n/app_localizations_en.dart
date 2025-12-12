// ignore: unused_import
import 'package:intl/intl.dart' as intl;
import 'app_localizations.dart';

// ignore_for_file: type=lint

/// The translations for English (`en`).
class AppLocalizationsEn extends AppLocalizations {
  AppLocalizationsEn([String locale = 'en']) : super(locale);

  @override
  String get appTitle => '5SecondsGo Admin';

  @override
  String get navDashboard => 'Dashboard';

  @override
  String get navUsers => 'Users';

  @override
  String get navRooms => 'Rooms';

  @override
  String get navFunds => 'Funds';

  @override
  String get navLogout => 'Logout';

  @override
  String get loginTitle => '5SecondsGo Admin';

  @override
  String get loginUsernameLabel => 'Username';

  @override
  String get loginUsernameRequired => 'Please enter username';

  @override
  String get loginPasswordLabel => 'Password';

  @override
  String get loginPasswordTooShort => 'Password must be at least 6 characters';

  @override
  String get loginButton => 'Login';

  @override
  String get dashboardTitle => 'Dashboard';

  @override
  String get dashboardCardTotalUsers => 'Total Users';

  @override
  String get dashboardCardActiveRooms => 'Active Rooms';

  @override
  String get dashboardCardOnlinePlayers => 'Online Players';

  @override
  String get dashboardCardPlatformBalance => 'Platform Balance';

  @override
  String get dashboardFundCheckTitle => 'Fund Conservation Check';

  @override
  String dashboardFundCheckPlayerBalance(Object amount) {
    return 'Total player balance: $amount';
  }

  @override
  String dashboardFundCheckCustodyQuota(Object amount) {
    return 'Total owner balance: $amount';
  }

  @override
  String dashboardFundCheckMargin(Object amount) {
    return 'Total owner profit: $amount';
  }

  @override
  String dashboardFundCheckPlatformProfit(Object amount) {
    return 'Total platform profit: $amount';
  }

  @override
  String get dashboardFundCheckBalanced => '✓ Balanced';

  @override
  String get dashboardRecentFundRequests => 'Recent Fund Requests';

  @override
  String dashboardRecentFundTitle(Object type, Object username) {
    return '$username - $type';
  }

  @override
  String get dashboardRecentFundTypeDeposit => 'Deposit';

  @override
  String get dashboardRecentFundTypeWithdraw => 'Withdraw';

  @override
  String get dashboardRecentFundStatusPending => 'Pending';

  @override
  String get dashboardRecentFundStatusApproved => 'Approved';

  @override
  String get dashboardRecentFundStatusRejected => 'Rejected';

  @override
  String get dashboardAdminInviteTitle => 'Admin Account';

  @override
  String dashboardAdminInviteCodeLabel(Object code) {
    return 'Invite code: $code';
  }

  @override
  String get fundsTitle => 'Funds';

  @override
  String get fundsTabRequests => 'Fund Requests';

  @override
  String get fundsTabTransactions => 'Transactions';

  @override
  String get fundsFilterAllStatus => 'All Status';

  @override
  String get fundsFilterStatusPending => 'Pending';

  @override
  String get fundsFilterStatusApproved => 'Approved';

  @override
  String get fundsFilterStatusRejected => 'Rejected';

  @override
  String get fundsFilterAllTypes => 'All Types';

  @override
  String get fundsFilterTypeDeposit => 'Deposit';

  @override
  String get fundsFilterTypeWithdraw => 'Withdraw';

  @override
  String get fundsFilterTypeMargin => 'Margin';

  @override
  String get fundsStatusPending => 'Pending';

  @override
  String get fundsStatusApproved => 'Approved';

  @override
  String get fundsStatusRejected => 'Rejected';

  @override
  String get fundsSearchUserHint => 'Search by user...';

  @override
  String get fundsTxTypeGameBet => 'Game Bet';

  @override
  String get fundsTxTypeGameWin => 'Game Win';

  @override
  String get fundsTxTypeDeposit => 'Deposit';

  @override
  String get fundsTxTypeWithdraw => 'Withdraw';

  @override
  String get roomsTitle => 'Rooms';

  @override
  String get roomsSearchHint => 'Search rooms...';

  @override
  String get roomsFilterAllStatus => 'All Status';

  @override
  String get roomsFilterStatusActive => 'Active';

  @override
  String get roomsFilterStatusPaused => 'Paused';

  @override
  String get roomsFilterStatusLocked => 'Locked';

  @override
  String roomsItemTitle(Object roomName) {
    return '$roomName';
  }

  @override
  String roomsItemSubtitle(Object bet, Object ownerName, Object players) {
    return 'Owner: $ownerName · Players: $players/20 · Bet: $bet';
  }

  @override
  String get roomsStatusActive => 'ACTIVE';

  @override
  String get roomsStatusPaused => 'PAUSED';

  @override
  String get roomsStatusLocked => 'LOCKED';

  @override
  String get usersTitle => 'Users';

  @override
  String get usersCreateOwner => 'Create Owner';

  @override
  String get usersSearchHint => 'Search users...';

  @override
  String get usersFilterAllRoles => 'All Roles';

  @override
  String get usersFilterRolePlayer => 'Player';

  @override
  String get usersFilterRoleOwner => 'Owner';

  @override
  String get usersFilterRoleAdmin => 'Admin';

  @override
  String usersItemTitle(Object username) {
    return '$username';
  }

  @override
  String usersItemSubtitleOwner(Object code) {
    return 'Owner · Invite: $code';
  }

  @override
  String usersItemSubtitlePlayer(Object ownerName) {
    return 'Player · Invited by: $ownerName';
  }

  @override
  String get usersDialogCreateOwnerTitle => 'Create Owner';

  @override
  String get usersDialogUsernameLabel => 'Username';

  @override
  String get usersDialogPasswordLabel => 'Password';

  @override
  String get usersDialogCancel => 'Cancel';

  @override
  String get usersDialogCreate => 'Create';

  @override
  String get navMonitoring => 'Monitoring';

  @override
  String get navRiskFlags => 'Risk Flags';

  @override
  String get navAlerts => 'Alerts';

  @override
  String get monitoringTitle => 'Monitoring Dashboard';

  @override
  String get monitoringRealtimeMetrics => 'Realtime Metrics';

  @override
  String get monitoringPerformanceMetrics => 'Performance Metrics';

  @override
  String get monitoringHistoricalTrends => 'Historical Trends';

  @override
  String get monitoringOnlinePlayers => 'Online Players';

  @override
  String get monitoringActiveRooms => 'Active Rooms';

  @override
  String get monitoringGamesPerMinute => 'Games/Min';

  @override
  String get monitoringDailyActiveUsers => 'Daily Active Users';

  @override
  String get monitoringDailyVolume => 'Daily Volume';

  @override
  String get monitoringPlatformRevenue => 'Platform Revenue';

  @override
  String get monitoringApiLatencyP95 => 'API Latency P95';

  @override
  String get monitoringWsLatencyP95 => 'WS Latency P95';

  @override
  String get monitoringDbLatencyP95 => 'DB Latency P95';

  @override
  String get refresh => 'Refresh';
}
