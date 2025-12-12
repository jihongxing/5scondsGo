import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/providers/locale_provider.dart';
import '../../../../core/services/api_client.dart';
import '../../../auth/providers/auth_provider.dart';
import '../../../room/providers/room_providers.dart';
import '../../../room/presentation/widgets/join_room_confirm_dialog.dart';
import '../../../wallet/providers/wallet_providers.dart';

class HomePage extends ConsumerWidget {
  const HomePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final authState = ref.watch(authProvider);
    final roomsAsync = ref.watch(roomsProvider);
    final walletAsync = ref.watch(walletInfoProvider);

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.casino, size: 28.sp, color: AppColors.accent),
              SizedBox(width: 8.w),
              Text(l10n.appTitle),
            ],
          ),
          actions: [
            IconButton(
              icon: const Icon(Icons.language),
              onPressed: () => ref.read(localeProvider.notifier).cycleLocale(),
            ),
            IconButton(
              icon: const Icon(Icons.logout),
              onPressed: () async {
                await ref.read(authProvider.notifier).logout();
                if (context.mounted) context.go('/login');
              },
            ),
          ],
        ),
        floatingActionButton: _buildFAB(context, authState),
        body: RefreshIndicator(
          onRefresh: () async {
            ref.invalidate(roomsProvider);
            ref.invalidate(walletInfoProvider);
          },
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: EdgeInsets.all(16.w),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                _buildUserCard(context, ref, authState, walletAsync, l10n),
                SizedBox(height: 24.h),
                _buildQuickActions(context, l10n),
                SizedBox(height: 24.h),
                _buildRoomHeader(l10n),
                SizedBox(height: 12.h),
                _buildRoomList(context, ref, roomsAsync, l10n),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget? _buildFAB(BuildContext context, dynamic authState) {
    if (authState.role != 'owner' && authState.role != 'admin') return null;
    return Container(
      decoration: BoxDecoration(
        gradient: AppColors.gradientAccent,
        borderRadius: BorderRadius.circular(16.r),
        boxShadow: [
          BoxShadow(
            color: AppColors.accent.withValues(alpha: 0.4),
            blurRadius: 15,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: FloatingActionButton(
        onPressed: () => context.push('/create-room'),
        backgroundColor: Colors.transparent,
        elevation: 0,
        child: const Icon(Icons.add, color: Colors.white),
      ),
    );
  }

  Widget _buildUserCard(
    BuildContext context,
    WidgetRef ref,
    dynamic authState,
    AsyncValue<WalletInfo> walletAsync,
    AppLocalizations l10n,
  ) {
    return GlassCard(
      padding: EdgeInsets.all(20.w),
      child: Column(
        children: [
          Row(
            children: [
              Container(
                padding: EdgeInsets.all(3.w),
                decoration: const BoxDecoration(
                  gradient: AppColors.gradientPrimary,
                  shape: BoxShape.circle,
                ),
                child: CircleAvatar(
                  radius: 32.r,
                  backgroundColor: AppColors.cardDark,
                  child: Text(
                    authState.username?.substring(0, 1).toUpperCase() ?? 'U',
                    style: TextStyle(
                      fontSize: 26.sp,
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
                    Row(
                      children: [
                        Text(
                          authState.username ?? 'User',
                          style: TextStyle(
                            fontSize: 20.sp,
                            fontWeight: FontWeight.bold,
                            color: AppColors.textPrimary,
                          ),
                        ),
                        SizedBox(width: 8.w),
                        _buildRoleBadge(authState.role, l10n),
                      ],
                    ),
                    SizedBox(height: 8.h),
                    walletAsync.when(
                      loading: () => _buildBalanceShimmer(),
                      error: (_, __) => Text(
                        '${l10n.balance}: --',
                        style: TextStyle(
                          fontSize: 14.sp,
                          color: AppColors.textSecondary,
                        ),
                      ),
                      data: (wallet) => _buildBalanceDisplay(wallet, l10n),
                    ),
                  ],
                ),
              ),
            ],
          ),
          SizedBox(height: 16.h),
          GradientButton(
            onPressed: () => context.go('/wallet'),
            gradient: AppColors.gradientPrimary,
            width: double.infinity,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.account_balance_wallet, color: Colors.white, size: 20.sp),
                SizedBox(width: 8.w),
                Text(
                  l10n.wallet,
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 16.sp,
                    fontWeight: FontWeight.w600,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildRoleBadge(String? role, AppLocalizations l10n) {
    Color color;
    String text;
    switch (role) {
      case 'owner':
        color = AppColors.warning;
        text = l10n.owner;
      case 'admin':
        color = AppColors.error;
        text = l10n.admin;
      default:
        color = AppColors.accent;
        text = l10n.player;
    }
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 8.w, vertical: 2.h),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(12.r),
        border: Border.all(color: color.withValues(alpha: 0.5)),
      ),
      child: Text(
        text,
        style: TextStyle(fontSize: 10.sp, color: color, fontWeight: FontWeight.w600),
      ),
    );
  }

  Widget _buildBalanceShimmer() {
    return Container(
      height: 16.h,
      width: 120.w,
      decoration: BoxDecoration(
        color: AppColors.glassWhite,
        borderRadius: BorderRadius.circular(4.r),
      ),
    );
  }

  Widget _buildBalanceDisplay(WalletInfo wallet, AppLocalizations l10n) {
    if (wallet.isOwner) {
      // 房主显示：可用余额 + 佣金收益
      return Row(
        children: [
          _buildMiniBalance(l10n.balance, wallet.availableBalance, AppColors.success),
          SizedBox(width: 16.w),
          _buildMiniBalance(l10n.commission, wallet.ownerRoomBalance, AppColors.warning),
        ],
      );
    }
    // 玩家显示可用余额
    return _buildMiniBalance(l10n.balance, wallet.availableBalance, AppColors.success);
  }

  Widget _buildMiniBalance(String label, double value, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(
          '$label: ',
          style: TextStyle(fontSize: 13.sp, color: AppColors.textSecondary),
        ),
        Text(
          '¥${value.toStringAsFixed(2)}',
          style: TextStyle(fontSize: 14.sp, color: color, fontWeight: FontWeight.bold),
        ),
      ],
    );
  }


  Widget _buildQuickActions(BuildContext context, AppLocalizations l10n) {
    return Row(
      children: [
        Expanded(
          child: _buildQuickActionCard(
            Icons.history,
            l10n.gameRecords,
            AppColors.primary,
            () => context.push('/game-history'),
          ),
        ),
        SizedBox(width: 12.w),
        Expanded(
          child: _buildQuickActionCard(
            Icons.people,
            l10n.friends,
            AppColors.accent,
            () => context.push('/friends'),
          ),
        ),
        SizedBox(width: 12.w),
        Expanded(
          child: _buildQuickActionCard(
            Icons.person,
            l10n.myProfile,
            AppColors.warning,
            () => context.push('/profile'),
          ),
        ),
      ],
    );
  }

  Widget _buildQuickActionCard(
    IconData icon,
    String label,
    Color color,
    VoidCallback onTap,
  ) {
    return GlassCard(
      padding: EdgeInsets.symmetric(vertical: 16.h, horizontal: 8.w),
      child: InkWell(
        onTap: onTap,
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.all(10.w),
              decoration: BoxDecoration(
                color: color.withValues(alpha: 0.2),
                shape: BoxShape.circle,
              ),
              child: Icon(icon, color: color, size: 24.sp),
            ),
            SizedBox(height: 8.h),
            Text(
              label,
              style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildRoomHeader(AppLocalizations l10n) {
    return Row(
      children: [
        NeonIcon(icon: Icons.meeting_room, color: AppColors.accent, size: 24.sp),
        SizedBox(width: 8.w),
        Text(
          l10n.rooms,
          style: TextStyle(
            fontSize: 20.sp,
            fontWeight: FontWeight.bold,
            color: AppColors.textPrimary,
          ),
        ),
      ],
    );
  }

  Widget _buildRoomList(
    BuildContext context,
    WidgetRef ref,
    AsyncValue<List<Map<String, dynamic>>> roomsAsync,
    AppLocalizations l10n,
  ) {
    return roomsAsync.when(
      loading: () => Column(
        children: List.generate(
          3,
          (_) => Padding(
            padding: EdgeInsets.only(bottom: 12.h),
            child: GlassCard(child: Container(height: 80.h)),
          ),
        ),
      ),
      error: (error, _) => GlassCard(
        child: Center(
          child: Text(
            error.toString(),
            style: TextStyle(color: AppColors.error, fontSize: 14.sp),
          ),
        ),
      ),
      data: (rooms) {
        if (rooms.isEmpty) {
          return GlassCard(
            padding: EdgeInsets.all(32.w),
            child: Column(
              children: [
                Icon(
                  Icons.meeting_room_outlined,
                  size: 64.sp,
                  color: AppColors.textSecondary,
                ),
                SizedBox(height: 16.h),
                Text(
                  l10n.noRooms,
                  style: TextStyle(fontSize: 16.sp, color: AppColors.textSecondary),
                ),
                SizedBox(height: 8.h),
                Text(
                  l10n.contactOwnerToCreate,
                  style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
                ),
              ],
            ),
          );
        }
        return ListView.builder(
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemCount: rooms.length,
          itemBuilder: (context, index) => _buildRoomCard(context, ref, rooms[index], l10n),
        );
      },
    );
  }

  Widget _buildRoomCard(
    BuildContext context,
    WidgetRef ref,
    Map<String, dynamic> room,
    AppLocalizations l10n,
  ) {
    final id = room['id'] as int? ?? 0;
    final name = room['name'] as String? ?? 'Room $id';
    final betAmount = room['bet_amount']?.toString() ?? '0';
    final maxPlayers = room['max_players'] as int? ?? 0;
    final currentPlayers = room['current_players'] as int? ?? 0;
    final hasPassword = room['has_password'] as bool? ?? false;
    final playerRatio = maxPlayers > 0 ? currentPlayers / maxPlayers : 0.0;
    final isHot = playerRatio > 0.7;

    return Padding(
      padding: EdgeInsets.only(bottom: 12.h),
      child: GlassCard(
        padding: EdgeInsets.all(16.w),
        child: InkWell(
          onTap: () => hasPassword
              ? _showPasswordDialog(context, id, room, ref)
              : _showJoinConfirmDialog(context, room, null, ref),
          child: Row(
            children: [
              Container(
                width: 56.w,
                height: 56.w,
                decoration: BoxDecoration(
                  gradient: isHot ? AppColors.gradientWarm : AppColors.gradientCool,
                  borderRadius: BorderRadius.circular(16.r),
                  boxShadow: [
                    BoxShadow(
                      color: (isHot ? AppColors.warning : AppColors.primary)
                          .withValues(alpha: 0.3),
                      blurRadius: 10,
                      offset: const Offset(0, 4),
                    ),
                  ],
                ),
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    Icon(
                      hasPassword ? Icons.lock : Icons.casino,
                      color: Colors.white,
                      size: 28.sp,
                    ),
                    if (isHot)
                      Positioned(
                        top: 0,
                        right: 0,
                        child: Container(
                          padding: EdgeInsets.all(4.w),
                          decoration: const BoxDecoration(
                            color: AppColors.error,
                            shape: BoxShape.circle,
                          ),
                          child: Icon(
                            Icons.local_fire_department,
                            color: Colors.white,
                            size: 12.sp,
                          ),
                        ),
                      ),
                  ],
                ),
              ),
              SizedBox(width: 16.w),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      name,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        fontSize: 16.sp,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    SizedBox(height: 6.h),
                    Row(
                      children: [
                        _buildRoomTag(Icons.currency_yuan, betAmount, AppColors.success),
                        SizedBox(width: 12.w),
                        _buildRoomTag(
                          Icons.people,
                          '$currentPlayers/$maxPlayers',
                          playerRatio > 0.7 ? AppColors.warning : AppColors.textSecondary,
                        ),
                      ],
                    ),
                  ],
                ),
              ),
              Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  GradientButton(
                    onPressed: () => hasPassword
                        ? _showPasswordDialog(context, id, room, ref)
                        : _showJoinConfirmDialog(context, room, null, ref),
                    gradient: AppColors.gradientAccent,
                    height: 36.h,
                    borderRadius: 18.r,
                    child: Padding(
                      padding: EdgeInsets.symmetric(horizontal: 12.w),
                      child: Text(
                        l10n.joinRoom,
                        style: TextStyle(
                          color: Colors.white,
                          fontSize: 12.sp,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                  ),
                  SizedBox(height: 6.h),
                  InkWell(
                    onTap: () => _spectateRoom(context, id, ref),
                    child: Container(
                      padding: EdgeInsets.symmetric(horizontal: 12.w, vertical: 6.h),
                      decoration: BoxDecoration(
                        color: AppColors.glassWhite,
                        borderRadius: BorderRadius.circular(12.r),
                        border: Border.all(color: AppColors.glassBorder),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.visibility, color: AppColors.textSecondary, size: 14.sp),
                          SizedBox(width: 4.w),
                          Text(
                            l10n.spectate,
                            style: TextStyle(
                              color: AppColors.textSecondary,
                              fontSize: 11.sp,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildRoomTag(IconData icon, String text, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(icon, size: 14.sp, color: color),
        SizedBox(width: 4.w),
        Text(text, style: TextStyle(fontSize: 12.sp, color: color)),
      ],
    );
  }

  Future<void> _joinRoom(
    BuildContext context,
    int roomId,
    String? password,
    WidgetRef ref,
  ) async {
    try {
      await ref.read(apiClientProvider).joinRoom(roomId, password: password);
      if (context.mounted) context.go('/room/$roomId');
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.toString())));
      }
    }
  }

  /// 显示加入房间确认弹窗
  Future<void> _showJoinConfirmDialog(
    BuildContext context,
    Map<String, dynamic> room,
    String? password,
    WidgetRef ref,
  ) async {
    final walletAsync = ref.read(walletInfoProvider);
    final userBalance = walletAsync.valueOrNull?.availableBalance ?? 0.0;

    final id = room['id'] as int? ?? 0;
    final name = room['name'] as String? ?? 'Room $id';
    final betAmount = room['bet_amount']?.toString() ?? '0';
    final winnerCount = room['winner_count'] as int? ?? 1;
    final maxPlayers = room['max_players'] as int? ?? 10;
    // 最少开始人数 = 赢家数 + 1
    final minPlayers = winnerCount + 1;

    if (!context.mounted) return;

    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) => JoinRoomConfirmDialog(
        roomName: name,
        betAmount: betAmount,
        minPlayers: minPlayers,
        winnerCount: winnerCount,
        maxPlayers: maxPlayers,
        userBalance: userBalance,
        onConfirm: () {
          Navigator.pop(dialogContext);
          _joinRoom(context, id, password, ref);
        },
        onCancel: () => Navigator.pop(dialogContext),
      ),
    );
  }

  Future<void> _spectateRoom(
    BuildContext context,
    int roomId,
    WidgetRef ref,
  ) async {
    try {
      await ref.read(apiClientProvider).spectateRoom(roomId);
      if (context.mounted) context.go('/room/$roomId?spectate=true');
    } catch (e) {
      if (context.mounted) {
        ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(e.toString())));
      }
    }
  }

  void _showPasswordDialog(BuildContext context, int roomId, Map<String, dynamic> room, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final controller = TextEditingController();
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: AppColors.cardDark,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
        title: Row(
          children: [
            Icon(Icons.lock, color: AppColors.warning, size: 24.sp),
            SizedBox(width: 8.w),
            Text(l10n.enterRoomPassword),
          ],
        ),
        content: TextField(
          controller: controller,
          obscureText: true,
          decoration: InputDecoration(
            hintText: l10n.passwordHint,
            prefixIcon: const Icon(Icons.key),
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: Text(l10n.cancel),
          ),
          GradientButton(
            onPressed: () {
              Navigator.pop(dialogContext);
              _showJoinConfirmDialog(context, room, controller.text, ref);
            },
            height: 40.h,
            child: Padding(
              padding: EdgeInsets.symmetric(horizontal: 16.w),
              child: Text(l10n.confirm, style: const TextStyle(color: Colors.white)),
            ),
          ),
        ],
      ),
    );
  }
}
