import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';
import '../../../room/providers/room_providers.dart';

class InviteFriendDialog extends ConsumerStatefulWidget {
  final int friendId;
  final String friendName;

  const InviteFriendDialog({
    super.key,
    required this.friendId,
    required this.friendName,
  });

  @override
  ConsumerState<InviteFriendDialog> createState() => _InviteFriendDialogState();
}

class _InviteFriendDialogState extends ConsumerState<InviteFriendDialog> {
  bool _isSending = false;
  int? _selectedRoomId;

  @override
  Widget build(BuildContext context) {
    final myRoomAsync = ref.watch(myRoomProvider);

    return AlertDialog(
      backgroundColor: AppColors.cardDark,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
      title: Row(
        children: [
          Icon(Icons.send, color: AppColors.accent, size: 24.sp),
          SizedBox(width: 8.w),
          Text('邀请 ${widget.friendName}'),
        ],
      ),
      content: myRoomAsync.when(
        loading: () => SizedBox(
          height: 100.h,
          child: const Center(child: CircularProgressIndicator()),
        ),
        error: (error, _) => Text(
          '加载房间失败: $error',
          style: TextStyle(color: AppColors.error),
        ),
        data: (room) {
          if (room == null) {
            return Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Icon(Icons.info_outline, color: AppColors.warning, size: 48.sp),
                SizedBox(height: 16.h),
                Text(
                  '您当前不在任何房间中',
                  style: TextStyle(color: AppColors.textSecondary),
                ),
                SizedBox(height: 8.h),
                Text(
                  '请先加入一个房间再邀请好友',
                  style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
                ),
              ],
            );
          }

          _selectedRoomId = room['id'] as int;
          final roomName = room['name'] as String? ?? '房间';

          return Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                '邀请好友加入:',
                style: TextStyle(color: AppColors.textSecondary, fontSize: 14.sp),
              ),
              SizedBox(height: 12.h),
              GlassCard(
                padding: EdgeInsets.all(12.w),
                child: Row(
                  children: [
                    Icon(Icons.casino, color: AppColors.accent, size: 24.sp),
                    SizedBox(width: 12.w),
                    Expanded(
                      child: Text(
                        roomName,
                        style: TextStyle(
                          fontSize: 16.sp,
                          fontWeight: FontWeight.bold,
                          color: AppColors.textPrimary,
                        ),
                      ),
                    ),
                    Icon(Icons.check_circle, color: AppColors.success, size: 20.sp),
                  ],
                ),
              ),
            ],
          );
        },
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        GradientButton(
          onPressed: _selectedRoomId != null && !_isSending ? _sendInvitation : null,
          height: 40.h,
          child: _isSending
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
                  child: const Text('发送邀请', style: TextStyle(color: Colors.white)),
                ),
        ),
      ],
    );
  }

  Future<void> _sendInvitation() async {
    if (_selectedRoomId == null) return;

    setState(() => _isSending = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.sendRoomInvitation(_selectedRoomId!, widget.friendId);
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('已向 ${widget.friendName} 发送邀请')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('发送邀请失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isSending = false);
    }
  }
}
