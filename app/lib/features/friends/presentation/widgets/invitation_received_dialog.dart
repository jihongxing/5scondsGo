import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';

class InvitationReceivedDialog extends ConsumerStatefulWidget {
  final int invitationId;
  final int roomId;
  final String roomName;
  final String fromUsername;

  const InvitationReceivedDialog({
    super.key,
    required this.invitationId,
    required this.roomId,
    required this.roomName,
    required this.fromUsername,
  });

  @override
  ConsumerState<InvitationReceivedDialog> createState() => _InvitationReceivedDialogState();
}

class _InvitationReceivedDialogState extends ConsumerState<InvitationReceivedDialog> {
  bool _isProcessing = false;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      backgroundColor: AppColors.cardDark,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
      title: Row(
        children: [
          Icon(Icons.mail, color: AppColors.accent, size: 24.sp),
          SizedBox(width: 8.w),
          const Text('房间邀请'),
        ],
      ),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: EdgeInsets.all(16.w),
            decoration: BoxDecoration(
              gradient: AppColors.gradientAccent,
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.casino, color: Colors.white, size: 32.sp),
          ),
          SizedBox(height: 16.h),
          Text(
            widget.fromUsername,
            style: TextStyle(
              fontSize: 18.sp,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          SizedBox(height: 4.h),
          Text(
            '邀请你加入房间',
            style: TextStyle(fontSize: 14.sp, color: AppColors.textSecondary),
          ),
          SizedBox(height: 16.h),
          GlassCard(
            padding: EdgeInsets.all(12.w),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.meeting_room, color: AppColors.primary, size: 20.sp),
                SizedBox(width: 8.w),
                Text(
                  widget.roomName,
                  style: TextStyle(
                    fontSize: 16.sp,
                    fontWeight: FontWeight.bold,
                    color: AppColors.textPrimary,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: _isProcessing ? null : () => _declineInvitation(context),
          child: Text(
            '拒绝',
            style: TextStyle(color: AppColors.error),
          ),
        ),
        GradientButton(
          onPressed: _isProcessing ? null : () => _acceptInvitation(context),
          gradient: AppColors.gradientAccent,
          height: 40.h,
          child: _isProcessing
              ? SizedBox(
                  width: 20.w,
                  height: 20.w,
                  child: const CircularProgressIndicator(
                    color: Colors.white,
                    strokeWidth: 2,
                  ),
                )
              : Padding(
                  padding: EdgeInsets.symmetric(horizontal: 16.w),
                  child: const Text('接受', style: TextStyle(color: Colors.white)),
                ),
        ),
      ],
    );
  }

  Future<void> _acceptInvitation(BuildContext context) async {
    setState(() => _isProcessing = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.acceptInvitation(widget.invitationId);
      if (mounted) {
        Navigator.pop(context);
        context.go('/room/${widget.roomId}');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('接受邀请失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isProcessing = false);
    }
  }

  Future<void> _declineInvitation(BuildContext context) async {
    setState(() => _isProcessing = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.declineInvitation(widget.invitationId);
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('已拒绝邀请')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('拒绝邀请失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isProcessing = false);
    }
  }
}
