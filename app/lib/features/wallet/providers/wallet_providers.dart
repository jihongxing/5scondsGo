import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../core/services/api_client.dart';

/// 钱包信息模型
class WalletInfo {
  final double availableBalance;  // 可用余额
  final double frozenBalance;     // 冻结余额（游戏中）
  final double totalBalance;      // 总余额
  // 房主专属字段
  final bool isOwner;
  final double ownerMarginBalance;    // 保证金（固定不变）
  final double ownerRoomBalance;      // 佣金收益（可转入余额）
  final double ownerTotalCommission;  // 累计佣金收益

  WalletInfo({
    required this.availableBalance,
    required this.frozenBalance,
    required this.totalBalance,
    this.isOwner = false,
    this.ownerMarginBalance = 0.0,
    this.ownerRoomBalance = 0.0,
    this.ownerTotalCommission = 0.0,
  });

  factory WalletInfo.fromJson(Map<String, dynamic> json) {
    return WalletInfo(
      availableBalance: _parseDouble(json['available_balance']),
      frozenBalance: _parseDouble(json['frozen_balance']),
      totalBalance: _parseDouble(json['total_balance']),
      isOwner: json['is_owner'] == true,
      ownerMarginBalance: _parseDouble(json['owner_margin_balance']),
      ownerRoomBalance: _parseDouble(json['owner_room_balance']),
      ownerTotalCommission: _parseDouble(json['owner_total_commission']),
    );
  }

  static double _parseDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is double) return value;
    if (value is int) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }
}

/// 收益统计模型
class EarningsSummary {
  final double totalWinnings;
  final double totalLosses;
  final double netProfit;
  final int totalRounds;
  final double winRate;
  final double todayProfit;
  final double weekProfit;
  final double monthProfit;

  EarningsSummary({
    required this.totalWinnings,
    required this.totalLosses,
    required this.netProfit,
    required this.totalRounds,
    required this.winRate,
    required this.todayProfit,
    required this.weekProfit,
    required this.monthProfit,
  });

  factory EarningsSummary.fromJson(Map<String, dynamic> json) {
    return EarningsSummary(
      totalWinnings: WalletInfo._parseDouble(json['total_winnings']),
      totalLosses: WalletInfo._parseDouble(json['total_losses']),
      netProfit: WalletInfo._parseDouble(json['net_profit']),
      totalRounds: json['total_rounds'] as int? ?? 0,
      winRate: WalletInfo._parseDouble(json['win_rate']),
      todayProfit: WalletInfo._parseDouble(json['today_profit']),
      weekProfit: WalletInfo._parseDouble(json['week_profit']),
      monthProfit: WalletInfo._parseDouble(json['month_profit']),
    );
  }
}

/// 钱包信息 Provider
final walletInfoProvider = FutureProvider.autoDispose<WalletInfo>((ref) async {
  final api = ref.read(apiClientProvider);
  final data = await api.getWallet();
  return WalletInfo.fromJson(data);
});

/// 收益统计 Provider
final earningsSummaryProvider = FutureProvider.autoDispose<EarningsSummary>((ref) async {
  final api = ref.read(apiClientProvider);
  final data = await api.getWalletEarnings();
  return EarningsSummary.fromJson(data);
});

/// 钱包交易记录 Provider
final walletTransactionsProvider = FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) async {
  final api = ref.read(apiClientProvider);
  final items = await api.getWalletTransactions(page: 1, pageSize: 20);
  return items;
});

/// 钱包总览（资金统计摘要）
final fundSummaryProvider =
    FutureProvider.autoDispose<Map<String, dynamic>>((ref) async {
  final api = ref.read(apiClientProvider);
  final summary = await api.getFundSummary();
  return summary;
});

/// 当前用户的资金申请列表（简单取第一页）
final fundRequestsProvider =
    FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) async {
  final api = ref.read(apiClientProvider);
  final items = await api.listFundRequests(page: 1, pageSize: 20);
  return items;
});

/// 当前用户的余额交易记录（简单取第一页）
final transactionsProvider =
    FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) async {
  final api = ref.read(apiClientProvider);
  final items = await api.listTransactions(page: 1, pageSize: 20);
  return items;
});

/// Owner 下级玩家的待审批资金申请
final ownerPendingFundRequestsProvider =
    FutureProvider.autoDispose<List<Map<String, dynamic>>>((ref) async {
  final api = ref.read(apiClientProvider);
  final items = await api.listOwnerFundRequests(page: 1, pageSize: 50, status: 'pending');
  return items;
});
