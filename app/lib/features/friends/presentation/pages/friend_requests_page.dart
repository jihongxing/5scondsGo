import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_theme.dart';
import '../../providers/friend_providers.dart';

class FriendRequestsPage extends ConsumerStatefulWidget {
  const FriendRequestsPage({super.key});

  @override
  ConsumerState<FriendRequestsPage> createState() => _FriendRequestsPageState();
}

class _FriendRequestsPageState extends ConsumerState<FriendRequestsPage> {
  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(friendRequestsProvider.notifier).loadRequests();
    });
  }

  @override
  Widget build(BuildContext context) {
    final state = ref.watch(friendRequestsProvider);

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: const Text('好友请求'),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () => context.pop(),
          ),
        ),
        body: state.isLoading
            ? _buildLoadingList()
            : state.error != null
                ? _buildErrorWidget(state.error!)
                : state.requests.isEmpty
                    ? _buildEmptyWidget()
                    : RefreshIndicator(
                        onRefresh: () => ref.read(friendRequestsProvider.notifier).loadRequests(),
                        child: ListView.builder(
                          padding: EdgeInsets.all(16.w),
                          itemCount: state.requests.length,
                          itemBuilder: (context, index) => _FriendRequestCard(
                            request: state.requests[index],
                          ),
                        ),
                      ),
      ),
    );
  }

  Widget _buildLoadingList() {
    return ListView.builder(
      padding: EdgeInsets.all(16.w),
      itemCount: 3,
      itemBuilder: (context, index) => Padding(
        padding: EdgeInsets.only(bottom: 12.h),
        child: GlassCard(child: Container(height: 80.h)),
      ),
    );
  }

  Widget _buildErrorWidget(String error) {
    return Center(
      child: GlassCard(
        padding: EdgeInsets.all(24.w),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.error_outline, color: AppColors.error, size: 48.sp),
            SizedBox(height: 16.h),
            Text(
              '加载失败',
              style: TextStyle(fontSize: 16.sp, color: AppColors.textPrimary),
            ),
            SizedBox(height: 8.h),
            Text(
              error,
              style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
              textAlign: TextAlign.center,
            ),
            SizedBox(height: 16.h),
            GradientButton(
              onPressed: () => ref.read(friendRequestsProvider.notifier).loadRequests(),
              height: 40.h,
              child: Padding(
                padding: EdgeInsets.symmetric(horizontal: 24.w),
                child: const Text('重试', style: TextStyle(color: Colors.white)),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildEmptyWidget() {
    return Center(
      child: GlassCard(
        padding: EdgeInsets.all(32.w),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.mail_outline, color: AppColors.textSecondary, size: 64.sp),
            SizedBox(height: 16.h),
            Text(
              '暂无好友请求',
              style: TextStyle(fontSize: 16.sp, color: AppColors.textSecondary),
            ),
          ],
        ),
      ),
    );
  }
}

class _FriendRequestCard extends ConsumerWidget {
  final FriendRequest request;

  const _FriendRequestCard({required this.request});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Padding(
      padding: EdgeInsets.only(bottom: 12.h),
      child: GlassCard(
        padding: EdgeInsets.all(16.w),
        child: Row(
          children: [
            Container(
              padding: EdgeInsets.all(2.w),
              decoration: const BoxDecoration(
                gradient: AppColors.gradientPrimary,
                shape: BoxShape.circle,
              ),
              child: CircleAvatar(
                radius: 24.r,
                backgroundColor: AppColors.cardDark,
                child: Text(
                  (request.fromUsername ?? 'U')[0].toUpperCase(),
                  style: TextStyle(
                    fontSize: 18.sp,
                    color: AppColors.textPrimary,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ),
            SizedBox(width: 16.w),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    request.fromUsername ?? '用户 ${request.fromUserId}',
                    style: TextStyle(
                      fontSize: 16.sp,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  SizedBox(height: 4.h),
                  Text(
                    '请求添加你为好友',
                    style: TextStyle(
                      fontSize: 12.sp,
                      color: AppColors.textSecondary,
                    ),
                  ),
                ],
              ),
            ),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                _buildActionButton(
                  icon: Icons.check,
                  color: AppColors.success,
                  onTap: () => _acceptRequest(context, ref),
                ),
                SizedBox(width: 8.w),
                _buildActionButton(
                  icon: Icons.close,
                  color: AppColors.error,
                  onTap: () => _rejectRequest(context, ref),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildActionButton({
    required IconData icon,
    required Color color,
    required VoidCallback onTap,
  }) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(20.r),
      child: Container(
        padding: EdgeInsets.all(8.w),
        decoration: BoxDecoration(
          color: color.withValues(alpha: 0.2),
          shape: BoxShape.circle,
          border: Border.all(color: color.withValues(alpha: 0.5)),
        ),
        child: Icon(icon, color: color, size: 20.sp),
      ),
    );
  }

  Future<void> _acceptRequest(BuildContext context, WidgetRef ref) async {
    await ref.read(friendRequestsProvider.notifier).acceptRequest(request.id);
    ref.read(friendListProvider.notifier).loadFriends();
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('已接受好友请求')),
      );
    }
  }

  Future<void> _rejectRequest(BuildContext context, WidgetRef ref) async {
    await ref.read(friendRequestsProvider.notifier).rejectRequest(request.id);
    if (context.mounted) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('已拒绝好友请求')),
      );
    }
  }
}
