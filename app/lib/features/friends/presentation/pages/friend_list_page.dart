import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../l10n/app_localizations.dart';
import '../../providers/friend_providers.dart';
import '../widgets/invite_friend_dialog.dart';

class FriendListPage extends ConsumerStatefulWidget {
  const FriendListPage({super.key});

  @override
  ConsumerState<FriendListPage> createState() => _FriendListPageState();
}

class _FriendListPageState extends ConsumerState<FriendListPage> {
  final _searchController = TextEditingController();
  String _searchQuery = '';

  @override
  void initState() {
    super.initState();
    Future.microtask(() {
      ref.read(friendListProvider.notifier).loadFriends();
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final state = ref.watch(friendListProvider);

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Text(l10n.friends),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () => context.pop(),
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.person_add),
              onPressed: () => _showAddFriendDialog(context),
            ),
            IconButton(
              icon: const Icon(Icons.mail),
              onPressed: () => context.push('/friend-requests'),
            ),
          ],
        ),
        body: Column(
          children: [
            _buildSearchBar(l10n),
            Expanded(
              child: state.isLoading
                  ? _buildLoadingList()
                  : state.error != null
                      ? _buildErrorWidget(state.error!)
                      : state.friends.isEmpty
                          ? _buildEmptyWidget()
                          : RefreshIndicator(
                              onRefresh: () => ref.read(friendListProvider.notifier).loadFriends(),
                              child: _buildFriendList(state.friends),
                            ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildSearchBar(AppLocalizations l10n) {
    return Padding(
      padding: EdgeInsets.all(16.w),
      child: GlassCard(
        padding: EdgeInsets.symmetric(horizontal: 16.w, vertical: 4.h),
        child: TextField(
          controller: _searchController,
          decoration: InputDecoration(
            hintText: l10n.searchFriends,
            hintStyle: TextStyle(color: AppColors.textSecondary),
            prefixIcon: Icon(Icons.search, color: AppColors.textSecondary),
            border: InputBorder.none,
            enabledBorder: InputBorder.none,
            focusedBorder: InputBorder.none,
          ),
          style: TextStyle(color: AppColors.textPrimary),
          onChanged: (value) {
            setState(() => _searchQuery = value.toLowerCase());
          },
        ),
      ),
    );
  }

  Widget _buildLoadingList() {
    return ListView.builder(
      padding: EdgeInsets.symmetric(horizontal: 16.w),
      itemCount: 5,
      itemBuilder: (context, index) => Padding(
        padding: EdgeInsets.only(bottom: 12.h),
        child: GlassCard(child: Container(height: 70.h)),
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
              onPressed: () => ref.read(friendListProvider.notifier).loadFriends(),
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
            Icon(Icons.people_outline, color: AppColors.textSecondary, size: 64.sp),
            SizedBox(height: 16.h),
            Text(
              '暂无好友',
              style: TextStyle(fontSize: 16.sp, color: AppColors.textSecondary),
            ),
            SizedBox(height: 8.h),
            Text(
              '点击右上角添加好友',
              style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFriendList(List<FriendInfo> friends) {
    final filteredFriends = _searchQuery.isEmpty
        ? friends
        : friends.where((f) => f.username.toLowerCase().contains(_searchQuery)).toList();

    return ListView.builder(
      padding: EdgeInsets.symmetric(horizontal: 16.w),
      itemCount: filteredFriends.length,
      itemBuilder: (context, index) => _FriendCard(friend: filteredFriends[index]),
    );
  }

  void _showAddFriendDialog(BuildContext context) {
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: AppColors.cardDark,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
        title: Row(
          children: [
            Icon(Icons.person_add, color: AppColors.accent, size: 24.sp),
            SizedBox(width: 8.w),
            const Text('添加好友'),
          ],
        ),
        content: TextField(
          controller: controller,
          decoration: const InputDecoration(
            labelText: '用户ID',
            hintText: '请输入用户ID',
          ),
          keyboardType: TextInputType.number,
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('取消'),
          ),
          GradientButton(
            onPressed: () async {
              final userId = int.tryParse(controller.text);
              if (userId != null) {
                try {
                  await ref.read(friendRequestsProvider.notifier).sendRequest(userId);
                  if (dialogContext.mounted) {
                    Navigator.pop(dialogContext);
                    ScaffoldMessenger.of(context).showSnackBar(
                      const SnackBar(content: Text('好友请求已发送')),
                    );
                  }
                } catch (e) {
                  if (dialogContext.mounted) {
                    ScaffoldMessenger.of(context).showSnackBar(
                      SnackBar(content: Text('发送失败: $e')),
                    );
                  }
                }
              }
            },
            height: 40.h,
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: 16.w),
              child: const Text('发送', style: TextStyle(color: Colors.white)),
            ),
          ),
        ],
      ),
    );
  }
}


class _FriendCard extends ConsumerWidget {
  final FriendInfo friend;

  const _FriendCard({required this.friend});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return Padding(
      padding: EdgeInsets.only(bottom: 12.h),
      child: GlassCard(
        padding: EdgeInsets.all(16.w),
        child: Row(
          children: [
            Stack(
              children: [
                Container(
                  padding: EdgeInsets.all(2.w),
                  decoration: BoxDecoration(
                    gradient: friend.isOnline
                        ? AppColors.gradientAccent
                        : const LinearGradient(colors: [Colors.grey, Colors.grey]),
                    shape: BoxShape.circle,
                  ),
                  child: CircleAvatar(
                    radius: 24.r,
                    backgroundColor: AppColors.cardDark,
                    child: Text(
                      friend.username[0].toUpperCase(),
                      style: TextStyle(
                        fontSize: 18.sp,
                        color: AppColors.textPrimary,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
                if (friend.isOnline)
                  Positioned(
                    right: 0,
                    bottom: 0,
                    child: Container(
                      width: 14.w,
                      height: 14.w,
                      decoration: BoxDecoration(
                        color: AppColors.success,
                        shape: BoxShape.circle,
                        border: Border.all(color: AppColors.cardDark, width: 2),
                      ),
                    ),
                  ),
              ],
            ),
            SizedBox(width: 16.w),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    friend.username,
                    style: TextStyle(
                      fontSize: 16.sp,
                      fontWeight: FontWeight.bold,
                      color: AppColors.textPrimary,
                    ),
                  ),
                  SizedBox(height: 4.h),
                  Row(
                    children: [
                      Container(
                        width: 8.w,
                        height: 8.w,
                        decoration: BoxDecoration(
                          color: friend.isOnline ? AppColors.success : AppColors.textSecondary,
                          shape: BoxShape.circle,
                        ),
                      ),
                      SizedBox(width: 6.w),
                      Text(
                        friend.isOnline
                            ? (friend.currentRoom != null ? '在房间中' : '在线')
                            : '离线',
                        style: TextStyle(
                          fontSize: 12.sp,
                          color: friend.isOnline ? AppColors.success : AppColors.textSecondary,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
            PopupMenuButton<String>(
              icon: Icon(Icons.more_vert, color: AppColors.textSecondary),
              color: AppColors.cardDark,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12.r)),
              onSelected: (value) => _handleMenuAction(context, ref, value),
              itemBuilder: (context) => [
                PopupMenuItem(
                  value: 'invite',
                  child: Row(
                    children: [
                      Icon(Icons.send, color: AppColors.accent, size: 20.sp),
                      SizedBox(width: 8.w),
                      const Text('邀请加入房间'),
                    ],
                  ),
                ),
                PopupMenuItem(
                  value: 'remove',
                  child: Row(
                    children: [
                      Icon(Icons.person_remove, color: AppColors.error, size: 20.sp),
                      SizedBox(width: 8.w),
                      Text('删除好友', style: TextStyle(color: AppColors.error)),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Future<void> _handleMenuAction(BuildContext context, WidgetRef ref, String action) async {
    switch (action) {
      case 'invite':
        showDialog(
          context: context,
          builder: (_) => InviteFriendDialog(friendId: friend.id, friendName: friend.username),
        );
        break;
      case 'remove':
        final confirm = await showDialog<bool>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            backgroundColor: AppColors.cardDark,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
            title: const Text('删除好友'),
            content: Text('确定要删除好友 ${friend.username} 吗？'),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('取消'),
              ),
              GradientButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                gradient: AppColors.gradientWarm,
                height: 40.h,
                child: Padding(
                  padding: EdgeInsets.symmetric(horizontal: 16.w),
                  child: const Text('删除', style: TextStyle(color: Colors.white)),
                ),
              ),
            ],
          ),
        );
        if (confirm == true) {
          await ref.read(friendListProvider.notifier).removeFriend(friend.id);
        }
        break;
    }
  }
}
