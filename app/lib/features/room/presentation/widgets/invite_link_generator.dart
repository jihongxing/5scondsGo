import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';

class InviteLinkGenerator extends ConsumerStatefulWidget {
  final int roomId;
  final String roomName;

  const InviteLinkGenerator({
    super.key,
    required this.roomId,
    required this.roomName,
  });

  @override
  ConsumerState<InviteLinkGenerator> createState() => _InviteLinkGeneratorState();
}

class _InviteLinkGeneratorState extends ConsumerState<InviteLinkGenerator> {
  bool _isGenerating = false;
  String? _inviteLink;
  String? _inviteCode;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      backgroundColor: AppColors.cardDark,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
      title: Row(
        children: [
          Icon(Icons.link, color: AppColors.accent, size: 24.sp),
          SizedBox(width: 8.w),
          const Text('生成邀请链接'),
        ],
      ),
      content: _inviteLink == null ? _buildGenerateContent() : _buildLinkContent(),
      actions: _inviteLink == null
          ? [
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: const Text('取消'),
              ),
              GradientButton(
                onPressed: _isGenerating ? null : _generateLink,
                height: 40.h,
                child: _isGenerating
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
                        child: const Text('生成', style: TextStyle(color: Colors.white)),
                      ),
              ),
            ]
          : [
              TextButton(
                onPressed: () => Navigator.pop(context),
                child: const Text('关闭'),
              ),
            ],
    );
  }

  Widget _buildGenerateContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.qr_code, color: AppColors.textSecondary, size: 64.sp),
        SizedBox(height: 16.h),
        Text(
          '生成邀请链接分享给好友',
          style: TextStyle(fontSize: 14.sp, color: AppColors.textSecondary),
        ),
        SizedBox(height: 8.h),
        Text(
          '好友可通过链接直接加入房间',
          style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
        ),
      ],
    );
  }

  Widget _buildLinkContent() {
    return Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          padding: EdgeInsets.all(16.w),
          decoration: BoxDecoration(
            gradient: AppColors.gradientAccent,
            shape: BoxShape.circle,
          ),
          child: Icon(Icons.check, color: Colors.white, size: 32.sp),
        ),
        SizedBox(height: 16.h),
        Text(
          '邀请链接已生成',
          style: TextStyle(
            fontSize: 16.sp,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        SizedBox(height: 16.h),
        GlassCard(
          padding: EdgeInsets.all(12.w),
          child: Row(
            children: [
              Expanded(
                child: Text(
                  _inviteCode ?? '',
                  style: TextStyle(
                    fontSize: 14.sp,
                    color: AppColors.accent,
                    fontFamily: 'monospace',
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              IconButton(
                icon: Icon(Icons.copy, color: AppColors.textSecondary, size: 20.sp),
                onPressed: _copyLink,
              ),
            ],
          ),
        ),
        SizedBox(height: 16.h),
        Row(
          children: [
            Expanded(
              child: GradientButton(
                onPressed: _copyLink,
                gradient: AppColors.gradientPrimary,
                height: 44.h,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.copy, color: Colors.white, size: 18.sp),
                    SizedBox(width: 8.w),
                    const Text('复制', style: TextStyle(color: Colors.white)),
                  ],
                ),
              ),
            ),
            SizedBox(width: 12.w),
            Expanded(
              child: GradientButton(
                onPressed: _shareLink,
                gradient: AppColors.gradientAccent,
                height: 44.h,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.share, color: Colors.white, size: 18.sp),
                    SizedBox(width: 8.w),
                    const Text('分享', style: TextStyle(color: Colors.white)),
                  ],
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Future<void> _generateLink() async {
    setState(() => _isGenerating = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      final result = await apiClient.createInviteLink(widget.roomId);
      final code = result['code'] as String? ?? result['invite_code'] as String? ?? '';
      setState(() {
        _inviteCode = code;
        _inviteLink = 'app://5secondsgo/invite/$code';
      });
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('生成链接失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isGenerating = false);
    }
  }

  void _copyLink() {
    if (_inviteLink != null) {
      Clipboard.setData(ClipboardData(text: _inviteLink!));
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('链接已复制到剪贴板')),
      );
    }
  }

  void _shareLink() {
    // 分享功能 - 复制到剪贴板作为替代
    if (_inviteLink != null) {
      Clipboard.setData(ClipboardData(
        text: '邀请你加入 ${widget.roomName} 房间！\n$_inviteLink',
      ));
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('邀请信息已复制，可粘贴分享')),
      );
    }
  }
}
