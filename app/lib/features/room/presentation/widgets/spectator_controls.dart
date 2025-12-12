import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';
import '../../../../l10n/app_localizations.dart';

class SpectatorControls extends ConsumerStatefulWidget {
  final int roomId;
  final bool canSwitchToParticipant;
  final VoidCallback? onSwitched;

  const SpectatorControls({
    super.key,
    required this.roomId,
    required this.canSwitchToParticipant,
    this.onSwitched,
  });

  @override
  ConsumerState<SpectatorControls> createState() => _SpectatorControlsState();
}

class _SpectatorControlsState extends ConsumerState<SpectatorControls> {
  bool _isSwitching = false;

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return GlassCard(
      padding: EdgeInsets.all(16.w),
      backgroundColor: AppColors.primary.withValues(alpha: 0.1),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.visibility, color: AppColors.primary, size: 20.sp),
              SizedBox(width: 8.w),
              Text(
                '观战模式',
                style: TextStyle(
                  fontSize: 16.sp,
                  fontWeight: FontWeight.bold,
                  color: AppColors.primary,
                ),
              ),
            ],
          ),
          SizedBox(height: 8.h),
          Text(
            '您正在观看比赛，无法参与下注',
            style: TextStyle(
              fontSize: 12.sp,
              color: AppColors.textSecondary,
            ),
          ),
          if (widget.canSwitchToParticipant) ...[
            SizedBox(height: 16.h),
            GradientButton(
              onPressed: _isSwitching ? null : _switchToParticipant,
              gradient: AppColors.gradientAccent,
              height: 40.h,
              child: _isSwitching
                  ? SizedBox(
                      width: 20.w,
                      height: 20.w,
                      child: const CircularProgressIndicator(
                        color: Colors.white,
                        strokeWidth: 2,
                      ),
                    )
                  : Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        Icon(Icons.person_add, color: Colors.white, size: 18.sp),
                        SizedBox(width: 8.w),
                        Text(
                          l10n.switchToParticipant,
                          style: TextStyle(
                            color: Colors.white,
                            fontSize: 14.sp,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                      ],
                    ),
            ),
          ],
        ],
      ),
    );
  }

  Future<void> _switchToParticipant() async {
    setState(() => _isSwitching = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.switchToParticipant(widget.roomId);
      widget.onSwitched?.call();
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('已切换为参与者')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('切换失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isSwitching = false);
    }
  }
}
