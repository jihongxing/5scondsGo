import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';

class InviteLinkPage extends ConsumerStatefulWidget {
  final String code;

  const InviteLinkPage({super.key, required this.code});

  @override
  ConsumerState<InviteLinkPage> createState() => _InviteLinkPageState();
}

class _InviteLinkPageState extends ConsumerState<InviteLinkPage> {
  bool _isLoading = true;
  bool _isJoining = false;
  String? _error;
  Map<String, dynamic>? _roomInfo;

  @override
  void initState() {
    super.initState();
    _validateInviteCode();
  }

  Future<void> _validateInviteCode() async {
    try {
      final apiClient = ref.read(apiClientProvider);
      // 尝试获取邀请链接对应的房间信息
      final rooms = await apiClient.listRooms(inviteCode: widget.code);
      if (rooms.isNotEmpty) {
        setState(() {
          _roomInfo = rooms.first;
          _isLoading = false;
        });
      } else {
        setState(() {
          _error = '邀请链接无效或已过期';
          _isLoading = false;
        });
      }
    } catch (e) {
      setState(() {
        _error = '验证邀请链接失败: $e';
        _isLoading = false;
      });
    }
  }

  Future<void> _joinRoom() async {
    if (_roomInfo == null) return;

    setState(() => _isJoining = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.joinByInviteLink(widget.code);
      if (mounted) {
        final roomId = _roomInfo!['id'] as int;
        context.go('/room/$roomId');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加入房间失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isJoining = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: const Text('房间邀请'),
          leading: IconButton(
            icon: const Icon(Icons.close),
            onPressed: () => context.go('/home'),
          ),
        ),
        body: Center(
          child: Padding(
            padding: EdgeInsets.all(24.w),
            child: _isLoading
                ? _buildLoadingWidget()
                : _error != null
                    ? _buildErrorWidget()
                    : _buildRoomInfoWidget(),
          ),
        ),
      ),
    );
  }

  Widget _buildLoadingWidget() {
    return GlassCard(
      padding: EdgeInsets.all(32.w),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          const CircularProgressIndicator(color: AppColors.accent),
          SizedBox(height: 16.h),
          Text(
            '正在验证邀请链接...',
            style: TextStyle(fontSize: 16.sp, color: AppColors.textSecondary),
          ),
        ],
      ),
    );
  }

  Widget _buildErrorWidget() {
    return GlassCard(
      padding: EdgeInsets.all(32.w),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.error_outline, color: AppColors.error, size: 64.sp),
          SizedBox(height: 16.h),
          Text(
            '邀请无效',
            style: TextStyle(
              fontSize: 20.sp,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          SizedBox(height: 8.h),
          Text(
            _error!,
            style: TextStyle(fontSize: 14.sp, color: AppColors.textSecondary),
            textAlign: TextAlign.center,
          ),
          SizedBox(height: 24.h),
          GradientButton(
            onPressed: () => context.go('/home'),
            width: double.infinity,
            child: const Text('返回首页', style: TextStyle(color: Colors.white)),
          ),
        ],
      ),
    );
  }

  Widget _buildRoomInfoWidget() {
    final name = _roomInfo!['name'] as String? ?? '未知房间';
    final betAmount = _roomInfo!['bet_amount']?.toString() ?? '0';
    final maxPlayers = _roomInfo!['max_players'] as int? ?? 0;
    final currentPlayers = _roomInfo!['current_players'] as int? ?? 0;

    return GlassCard(
      padding: EdgeInsets.all(24.w),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Container(
            padding: EdgeInsets.all(16.w),
            decoration: BoxDecoration(
              gradient: AppColors.gradientAccent,
              shape: BoxShape.circle,
            ),
            child: Icon(Icons.casino, color: Colors.white, size: 48.sp),
          ),
          SizedBox(height: 16.h),
          Text(
            '您被邀请加入',
            style: TextStyle(fontSize: 14.sp, color: AppColors.textSecondary),
          ),
          SizedBox(height: 8.h),
          Text(
            name,
            style: TextStyle(
              fontSize: 24.sp,
              fontWeight: FontWeight.bold,
              color: AppColors.textPrimary,
            ),
          ),
          SizedBox(height: 24.h),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceEvenly,
            children: [
              _buildInfoItem(Icons.currency_yuan, '¥$betAmount', '下注金额'),
              _buildInfoItem(Icons.people, '$currentPlayers/$maxPlayers', '玩家人数'),
            ],
          ),
          SizedBox(height: 32.h),
          GradientButton(
            onPressed: _isJoining ? null : _joinRoom,
            gradient: AppColors.gradientAccent,
            width: double.infinity,
            child: _isJoining
                ? SizedBox(
                    width: 20.w,
                    height: 20.w,
                    child: const CircularProgressIndicator(
                      color: Colors.white,
                      strokeWidth: 2,
                    ),
                  )
                : Text(
                    '加入房间',
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 16.sp,
                      fontWeight: FontWeight.w600,
                    ),
                  ),
          ),
          SizedBox(height: 12.h),
          TextButton(
            onPressed: () => context.go('/home'),
            child: Text(
              '返回首页',
              style: TextStyle(color: AppColors.textSecondary, fontSize: 14.sp),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoItem(IconData icon, String value, String label) {
    return Column(
      children: [
        Icon(icon, color: AppColors.accent, size: 24.sp),
        SizedBox(height: 4.h),
        Text(
          value,
          style: TextStyle(
            fontSize: 18.sp,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
        Text(
          label,
          style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
        ),
      ],
    );
  }
}
