import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../core/services/api_client.dart';
import '../../../../core/theme/app_theme.dart';

class GameHistoryPage extends ConsumerStatefulWidget {
  const GameHistoryPage({super.key});

  @override
  ConsumerState<GameHistoryPage> createState() => _GameHistoryPageState();
}

class _GameHistoryPageState extends ConsumerState<GameHistoryPage> {
  List<GameHistoryItem> _items = [];
  GameStats? _stats;
  bool _isLoading = true;
  int _page = 1;
  int _total = 0;
  final int _pageSize = 20;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      
      // 加载历史记录
      final historyResponse = await apiClient.get('/game-history?page=$_page&page_size=$_pageSize');
      final items = (historyResponse['items'] as List?)
          ?.map((e) => GameHistoryItem.fromJson(e as Map<String, dynamic>))
          .toList() ?? [];
      _total = historyResponse['total'] as int? ?? 0;

      // 加载统计
      final statsResponse = await apiClient.get('/game-stats');
      final stats = GameStats.fromJson(statsResponse);

      setState(() {
        _items = items;
        _stats = stats;
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        final l10n = AppLocalizations.of(context)!;
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('${l10n.loadFailed}: $e')),
        );
      }
    }
  }


  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.gameRecords),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.go('/home'),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : RefreshIndicator(
              onRefresh: _loadData,
              child: CustomScrollView(
                slivers: [
                  // 统计卡片
                  if (_stats != null)
                    SliverToBoxAdapter(
                      child: _buildStatsCard(),
                    ),
                  // 历史记录列表
                  SliverPadding(
                    padding: EdgeInsets.all(16.w),
                    sliver: SliverList(
                      delegate: SliverChildBuilderDelegate(
                        (context, index) {
                          if (index >= _items.length) return null;
                          return _buildHistoryItem(_items[index]);
                        },
                        childCount: _items.length,
                      ),
                    ),
                  ),
                  // 分页
                  if (_total > _pageSize)
                    SliverToBoxAdapter(
                      child: _buildPagination(),
                    ),
                ],
              ),
            ),
    );
  }

  Widget _buildStatsCard() {
    final stats = _stats!;
    final l10n = AppLocalizations.of(context)!;
    return Container(
      margin: EdgeInsets.all(16.w),
      padding: EdgeInsets.all(16.w),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [AppColors.primary, AppColors.primary.withAlpha(179)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16.r),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildStatItem(l10n.totalGames, '${stats.totalRounds}'),
              _buildStatItem(l10n.wins, '${stats.totalWins}'),
              _buildStatItem(l10n.winRate, '${stats.winRate.toStringAsFixed(1)}%'),
            ],
          ),
          SizedBox(height: 16.h),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              _buildStatItem(l10n.totalBet, '¥${stats.totalWagered}'),
              _buildStatItem(l10n.totalWon, '¥${stats.totalWon}'),
              _buildStatItem(
                l10n.netProfit,
                '¥${stats.netProfit}',
                color: stats.netProfit >= 0 ? Colors.greenAccent : Colors.redAccent,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem(String label, String value, {Color? color}) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: 20.sp,
            fontWeight: FontWeight.bold,
            color: color ?? Colors.white,
          ),
        ),
        Text(
          label,
          style: TextStyle(
            fontSize: 12.sp,
            color: Colors.white.withAlpha(230),
          ),
        ),
      ],
    );
  }

  Widget _buildHistoryItem(GameHistoryItem item) {
    final isWin = item.result == 'win';
    final isSkipped = item.result == 'skipped';
    final l10n = AppLocalizations.of(context)!;

    return Card(
      margin: EdgeInsets.only(bottom: 8.h),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: isWin
              ? AppColors.success.withAlpha(51)
              : isSkipped
                  ? Colors.grey.withAlpha(51)
                  : AppColors.error.withAlpha(51),
          child: Icon(
            isWin ? Icons.emoji_events : isSkipped ? Icons.skip_next : Icons.close,
            color: isWin ? AppColors.success : isSkipped ? Colors.grey : AppColors.error,
          ),
        ),
        title: Text(item.roomName),
        subtitle: Text(
          '${l10n.roundNumber(item.roundNumber)} · ${_formatDate(item.createdAt)}',
          style: TextStyle(fontSize: 12.sp),
        ),
        trailing: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          crossAxisAlignment: CrossAxisAlignment.end,
          children: [
            Text(
              isWin ? '+¥${item.prizeAmount}' : isSkipped ? l10n.skipped : '-¥${item.betAmount}',
              style: TextStyle(
                fontSize: 16.sp,
                fontWeight: FontWeight.bold,
                color: isWin ? AppColors.success : isSkipped ? Colors.grey : AppColors.error,
              ),
            ),
            Text(
              '${l10n.bet}: ¥${item.betAmount}',
              style: TextStyle(fontSize: 10.sp, color: AppColors.textSecondary),
            ),
          ],
        ),
        onTap: () => _showRoundDetail(item.id),
      ),
    );
  }

  Widget _buildPagination() {
    final totalPages = (_total / _pageSize).ceil();
    return Padding(
      padding: EdgeInsets.all(16.w),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          IconButton(
            onPressed: _page > 1 ? () => _changePage(_page - 1) : null,
            icon: const Icon(Icons.chevron_left),
          ),
          Text('$_page / $totalPages'),
          IconButton(
            onPressed: _page < totalPages ? () => _changePage(_page + 1) : null,
            icon: const Icon(Icons.chevron_right),
          ),
        ],
      ),
    );
  }

  void _changePage(int page) {
    setState(() => _page = page);
    _loadData();
  }

  void _showRoundDetail(int roundId) {
    context.push('/game-history/$roundId');
  }

  String _formatDate(DateTime date) {
    return '${date.month}/${date.day} ${date.hour}:${date.minute.toString().padLeft(2, '0')}';
  }
}

class GameHistoryItem {
  final int id;
  final int roomId;
  final String roomName;
  final int roundNumber;
  final String betAmount;
  final String result;
  final String prizeAmount;
  final DateTime createdAt;

  GameHistoryItem({
    required this.id,
    required this.roomId,
    required this.roomName,
    required this.roundNumber,
    required this.betAmount,
    required this.result,
    required this.prizeAmount,
    required this.createdAt,
  });

  factory GameHistoryItem.fromJson(Map<String, dynamic> json) {
    return GameHistoryItem(
      id: json['id'] as int? ?? 0,
      roomId: json['room_id'] as int? ?? 0,
      roomName: json['room_name'] as String? ?? '',
      roundNumber: json['round_number'] as int? ?? 0,
      betAmount: json['bet_amount']?.toString() ?? '0',
      result: json['result'] as String? ?? '',
      prizeAmount: json['prize_amount']?.toString() ?? '0',
      createdAt: DateTime.tryParse(json['created_at'] as String? ?? '') ?? DateTime.now(),
    );
  }
}

class GameStats {
  final int totalRounds;
  final int totalWins;
  final int totalLosses;
  final int totalSkipped;
  final double winRate;
  final double totalWagered;
  final double totalWon;
  final double netProfit;

  GameStats({
    required this.totalRounds,
    required this.totalWins,
    required this.totalLosses,
    required this.totalSkipped,
    required this.winRate,
    required this.totalWagered,
    required this.totalWon,
    required this.netProfit,
  });

  factory GameStats.fromJson(Map<String, dynamic> json) {
    return GameStats(
      totalRounds: json['total_rounds'] as int? ?? 0,
      totalWins: json['total_wins'] as int? ?? 0,
      totalLosses: json['total_losses'] as int? ?? 0,
      totalSkipped: json['total_skipped'] as int? ?? 0,
      winRate: (json['win_rate'] as num?)?.toDouble() ?? 0,
      totalWagered: double.tryParse(json['total_wagered']?.toString() ?? '0') ?? 0,
      totalWon: double.tryParse(json['total_won']?.toString() ?? '0') ?? 0,
      netProfit: double.tryParse(json['net_profit']?.toString() ?? '0') ?? 0,
    );
  }
}
