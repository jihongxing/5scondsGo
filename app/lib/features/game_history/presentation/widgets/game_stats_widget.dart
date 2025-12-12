import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';

/// 游戏统计数据模型
class GameStats {
  final int totalRounds;
  final int totalWins;
  final int totalLosses;
  final int totalSkipped;
  final double winRate;
  final double totalWagered;
  final double totalWon;
  final double netProfit;

  const GameStats({
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
      winRate: (json['win_rate'] as num?)?.toDouble() ?? 0.0,
      totalWagered: double.tryParse(json['total_wagered']?.toString() ?? '0') ?? 0.0,
      totalWon: double.tryParse(json['total_won']?.toString() ?? '0') ?? 0.0,
      netProfit: double.tryParse(json['net_profit']?.toString() ?? '0') ?? 0.0,
    );
  }
}

/// 游戏统计组件
class GameStatsWidget extends StatelessWidget {
  final GameStats stats;

  const GameStatsWidget({
    super.key,
    required this.stats,
  });

  @override
  Widget build(BuildContext context) {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16.w),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              '游戏统计',
              style: TextStyle(
                fontSize: 18.sp,
                fontWeight: FontWeight.bold,
              ),
            ),
            SizedBox(height: 16.h),
            // 胜率圆环
            Center(
              child: _buildWinRateCircle(),
            ),
            SizedBox(height: 16.h),
            // 统计数据网格
            _buildStatsGrid(),
            SizedBox(height: 16.h),
            // 盈亏统计
            _buildProfitSection(),
          ],
        ),
      ),
    );
  }

  Widget _buildWinRateCircle() {
    return SizedBox(
      width: 120.w,
      height: 120.w,
      child: Stack(
        alignment: Alignment.center,
        children: [
          SizedBox(
            width: 120.w,
            height: 120.w,
            child: CircularProgressIndicator(
              value: stats.winRate / 100,
              strokeWidth: 12.w,
              backgroundColor: Colors.grey.shade200,
              valueColor: AlwaysStoppedAnimation<Color>(
                stats.winRate >= 50 ? AppColors.success : AppColors.warning,
              ),
            ),
          ),
          Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                '${stats.winRate.toStringAsFixed(1)}%',
                style: TextStyle(
                  fontSize: 24.sp,
                  fontWeight: FontWeight.bold,
                  color: stats.winRate >= 50 ? AppColors.success : AppColors.warning,
                ),
              ),
              Text(
                '胜率',
                style: TextStyle(
                  fontSize: 12.sp,
                  color: AppColors.textSecondary,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildStatsGrid() {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceAround,
      children: [
        _buildStatItem('总场次', stats.totalRounds.toString(), AppColors.primary),
        _buildStatItem('胜利', stats.totalWins.toString(), AppColors.success),
        _buildStatItem('失败', stats.totalLosses.toString(), AppColors.error),
        _buildStatItem('跳过', stats.totalSkipped.toString(), AppColors.textSecondary),
      ],
    );
  }

  Widget _buildStatItem(String label, String value, Color color) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: 20.sp,
            fontWeight: FontWeight.bold,
            color: color,
          ),
        ),
        SizedBox(height: 4.h),
        Text(
          label,
          style: TextStyle(
            fontSize: 12.sp,
            color: AppColors.textSecondary,
          ),
        ),
      ],
    );
  }

  Widget _buildProfitSection() {
    final isProfit = stats.netProfit >= 0;
    
    return Container(
      padding: EdgeInsets.all(12.w),
      decoration: BoxDecoration(
        color: isProfit 
            ? AppColors.success.withAlpha(25) 
            : AppColors.error.withAlpha(25),
        borderRadius: BorderRadius.circular(12.r),
      ),
      child: Column(
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '总投注',
                style: TextStyle(
                  fontSize: 14.sp,
                  color: AppColors.textSecondary,
                ),
              ),
              Text(
                '¥${stats.totalWagered.toStringAsFixed(2)}',
                style: TextStyle(
                  fontSize: 14.sp,
                  fontWeight: FontWeight.w500,
                ),
              ),
            ],
          ),
          SizedBox(height: 8.h),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '总赢得',
                style: TextStyle(
                  fontSize: 14.sp,
                  color: AppColors.textSecondary,
                ),
              ),
              Text(
                '¥${stats.totalWon.toStringAsFixed(2)}',
                style: TextStyle(
                  fontSize: 14.sp,
                  fontWeight: FontWeight.w500,
                  color: AppColors.success,
                ),
              ),
            ],
          ),
          Divider(height: 16.h),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '净盈亏',
                style: TextStyle(
                  fontSize: 16.sp,
                  fontWeight: FontWeight.bold,
                ),
              ),
              Text(
                '${isProfit ? '+' : ''}¥${stats.netProfit.toStringAsFixed(2)}',
                style: TextStyle(
                  fontSize: 18.sp,
                  fontWeight: FontWeight.bold,
                  color: isProfit ? AppColors.success : AppColors.error,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
