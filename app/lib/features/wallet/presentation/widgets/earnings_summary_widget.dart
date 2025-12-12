import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';
import '../../providers/wallet_providers.dart';

class EarningsSummaryWidget extends ConsumerWidget {
  const EarningsSummaryWidget({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final earningsAsync = ref.watch(earningsSummaryProvider);

    return earningsAsync.when(
      loading: () => const Center(child: CircularProgressIndicator()),
      error: (error, stack) => Center(
        child: Text('加载失败: $error', style: TextStyle(color: AppColors.error)),
      ),
      data: (earnings) => Card(
        child: Padding(
          padding: EdgeInsets.all(16.w),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '收益统计',
                style: TextStyle(
                  fontSize: 18.sp,
                  fontWeight: FontWeight.bold,
                ),
              ),
              SizedBox(height: 16.h),
              
              // 净收益
              _buildProfitCard(
                context,
                title: '净收益',
                value: earnings.netProfit,
                isMain: true,
              ),
              SizedBox(height: 12.h),
              
              // 时间段收益
              Row(
                children: [
                  Expanded(
                    child: _buildProfitCard(
                      context,
                      title: '今日',
                      value: earnings.todayProfit,
                    ),
                  ),
                  SizedBox(width: 8.w),
                  Expanded(
                    child: _buildProfitCard(
                      context,
                      title: '本周',
                      value: earnings.weekProfit,
                    ),
                  ),
                  SizedBox(width: 8.w),
                  Expanded(
                    child: _buildProfitCard(
                      context,
                      title: '本月',
                      value: earnings.monthProfit,
                    ),
                  ),
                ],
              ),
              SizedBox(height: 16.h),
              
              // 详细统计
              Divider(),
              SizedBox(height: 12.h),
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _buildStatItem(
                    context,
                    label: '总获胜',
                    value: '¥${earnings.totalWinnings.toStringAsFixed(2)}',
                    color: Colors.green,
                  ),
                  _buildStatItem(
                    context,
                    label: '总投注',
                    value: '¥${earnings.totalLosses.toStringAsFixed(2)}',
                    color: Colors.red,
                  ),
                  _buildStatItem(
                    context,
                    label: '总局数',
                    value: '${earnings.totalRounds}',
                    color: Colors.blue,
                  ),
                  _buildStatItem(
                    context,
                    label: '胜率',
                    value: '${earnings.winRate.toStringAsFixed(1)}%',
                    color: Colors.orange,
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildProfitCard(
    BuildContext context, {
    required String title,
    required double value,
    bool isMain = false,
  }) {
    final isPositive = value >= 0;
    final color = isPositive ? Colors.green : Colors.red;
    final prefix = isPositive ? '+' : '';

    return Container(
      padding: EdgeInsets.all(isMain ? 16.w : 12.w),
      decoration: BoxDecoration(
        color: color.withOpacity(0.1),
        borderRadius: BorderRadius.circular(8.r),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            title,
            style: TextStyle(
              fontSize: isMain ? 14.sp : 12.sp,
              color: AppColors.textSecondary,
            ),
          ),
          SizedBox(height: 4.h),
          Text(
            '$prefix¥${value.toStringAsFixed(2)}',
            style: TextStyle(
              fontSize: isMain ? 24.sp : 16.sp,
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStatItem(
    BuildContext context, {
    required String label,
    required String value,
    required Color color,
  }) {
    return Column(
      children: [
        Text(
          value,
          style: TextStyle(
            fontSize: 16.sp,
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
}
