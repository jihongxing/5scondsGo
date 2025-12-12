import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';

/// 加入房间确认弹窗
class JoinRoomConfirmDialog extends StatelessWidget {
  final String roomName;
  final String betAmount;
  final int minPlayers;
  final int winnerCount;
  final int maxPlayers;
  final double userBalance;
  final VoidCallback onConfirm;
  final VoidCallback onCancel;

  const JoinRoomConfirmDialog({
    super.key,
    required this.roomName,
    required this.betAmount,
    required this.minPlayers,
    required this.winnerCount,
    required this.maxPlayers,
    required this.userBalance,
    required this.onConfirm,
    required this.onCancel,
  });

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final betAmountNum = double.tryParse(betAmount) ?? 0;
    final hasEnoughBalance = userBalance >= betAmountNum;

    return Dialog(
      backgroundColor: Colors.transparent,
      child: Container(
        width: 340.w,
        padding: EdgeInsets.all(24.w),
        decoration: BoxDecoration(
          color: AppColors.cardDark,
          borderRadius: BorderRadius.circular(20.r),
          border: Border.all(color: AppColors.glassBorder),
          boxShadow: [
            BoxShadow(
              color: Colors.black.withOpacity(0.3),
              blurRadius: 20,
              offset: const Offset(0, 10),
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 标题
            Row(
              children: [
                Icon(Icons.casino, color: AppColors.accent, size: 28.sp),
                SizedBox(width: 12.w),
                Expanded(
                  child: Text(
                    l10n.joinRoomTitle,
                    style: TextStyle(
                      fontSize: 20.sp,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                ),
                IconButton(
                  icon: Icon(Icons.close, color: AppColors.textSecondary, size: 24.sp),
                  onPressed: onCancel,
                  padding: EdgeInsets.zero,
                  constraints: const BoxConstraints(),
                ),
              ],
            ),
            SizedBox(height: 20.h),

            // 房间名称
            _buildInfoRow(Icons.meeting_room, l10n.roomName, roomName),
            SizedBox(height: 12.h),

            // 房间信息
            Container(
              padding: EdgeInsets.all(16.w),
              decoration: BoxDecoration(
                color: AppColors.glassWhite,
                borderRadius: BorderRadius.circular(12.r),
              ),
              child: Column(
                children: [
                  _buildDetailRow(l10n.betPerRoundLabel, '¥$betAmount', AppColors.warning),
                  SizedBox(height: 8.h),
                  _buildDetailRow(l10n.minPlayersRequired, '$minPlayers ${l10n.personUnit}', AppColors.textSecondary),
                  SizedBox(height: 8.h),
                  _buildDetailRow(l10n.winnersPerRound, '$winnerCount ${l10n.personUnit}', AppColors.success),
                  SizedBox(height: 8.h),
                  _buildDetailRow(l10n.maxPlayersAllowed, '$maxPlayers ${l10n.personUnit}', AppColors.textSecondary),
                ],
              ),
            ),
            SizedBox(height: 16.h),

            // 余额信息
            Container(
              padding: EdgeInsets.all(16.w),
              decoration: BoxDecoration(
                color: hasEnoughBalance
                    ? AppColors.success.withOpacity(0.1)
                    : AppColors.error.withOpacity(0.1),
                borderRadius: BorderRadius.circular(12.r),
                border: Border.all(
                  color: hasEnoughBalance
                      ? AppColors.success.withOpacity(0.3)
                      : AppColors.error.withOpacity(0.3),
                ),
              ),
              child: Column(
                children: [
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        l10n.currentBalance,
                        style: TextStyle(
                          fontSize: 14.sp,
                          color: AppColors.textSecondary,
                        ),
                      ),
                      Text(
                        '¥${userBalance.toStringAsFixed(2)}',
                        style: TextStyle(
                          fontSize: 16.sp,
                          fontWeight: FontWeight.bold,
                          color: hasEnoughBalance ? AppColors.success : AppColors.error,
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 8.h),
                  Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        l10n.minBalanceRequired,
                        style: TextStyle(
                          fontSize: 14.sp,
                          color: AppColors.textSecondary,
                        ),
                      ),
                      Text(
                        '¥$betAmount',
                        style: TextStyle(
                          fontSize: 14.sp,
                          color: AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                  if (!hasEnoughBalance) ...[
                    SizedBox(height: 8.h),
                    Row(
                      children: [
                        Icon(Icons.warning_amber, color: AppColors.error, size: 16.sp),
                        SizedBox(width: 4.w),
                        Text(
                          l10n.insufficientBalanceCannotJoin,
                          style: TextStyle(
                            fontSize: 12.sp,
                            color: AppColors.error,
                          ),
                        ),
                      ],
                    ),
                  ],
                ],
              ),
            ),
            SizedBox(height: 16.h),

            // 风险提示
            Container(
              padding: EdgeInsets.all(12.w),
              decoration: BoxDecoration(
                color: AppColors.warning.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8.r),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.info_outline, color: AppColors.warning, size: 16.sp),
                      SizedBox(width: 6.w),
                      Text(
                        l10n.riskWarning,
                        style: TextStyle(
                          fontSize: 13.sp,
                          fontWeight: FontWeight.bold,
                          color: AppColors.warning,
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 8.h),
                  Text(
                    '• ${l10n.riskWarningTip1}',
                    style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
                  ),
                  SizedBox(height: 4.h),
                  Text(
                    '• ${l10n.riskWarningTip2}',
                    style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
                  ),
                ],
              ),
            ),
            SizedBox(height: 24.h),

            // 按钮
            Row(
              children: [
                Expanded(
                  child: OutlinedButton(
                    onPressed: onCancel,
                    style: OutlinedButton.styleFrom(
                      padding: EdgeInsets.symmetric(vertical: 14.h),
                      side: BorderSide(color: AppColors.glassBorder),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12.r),
                      ),
                    ),
                    child: Text(
                      l10n.cancel,
                      style: TextStyle(
                        fontSize: 16.sp,
                        color: AppColors.textSecondary,
                      ),
                    ),
                  ),
                ),
                SizedBox(width: 16.w),
                Expanded(
                  child: GradientButton(
                    onPressed: hasEnoughBalance ? onConfirm : null,
                    gradient: hasEnoughBalance
                        ? AppColors.gradientAccent
                        : LinearGradient(colors: [Colors.grey, Colors.grey.shade600]),
                    height: 48.h,
                    borderRadius: 12.r,
                    child: Text(
                      l10n.confirmJoin,
                      style: TextStyle(
                        fontSize: 16.sp,
                        fontWeight: FontWeight.w600,
                        color: Colors.white,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Row(
      children: [
        Icon(icon, color: AppColors.accent, size: 20.sp),
        SizedBox(width: 8.w),
        Text(
          '$label: ',
          style: TextStyle(
            fontSize: 14.sp,
            color: AppColors.textSecondary,
          ),
        ),
        Expanded(
          child: Text(
            value,
            style: TextStyle(
              fontSize: 14.sp,
              fontWeight: FontWeight.w500,
              color: AppColors.textPrimary,
            ),
            overflow: TextOverflow.ellipsis,
          ),
        ),
      ],
    );
  }

  Widget _buildDetailRow(String label, String value, Color valueColor) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 14.sp,
            color: AppColors.textSecondary,
          ),
        ),
        Text(
          value,
          style: TextStyle(
            fontSize: 14.sp,
            fontWeight: FontWeight.w600,
            color: valueColor,
          ),
        ),
      ],
    );
  }
}
