import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/providers/locale_provider.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../auth/providers/auth_provider.dart';
import '../../../wallet/providers/wallet_providers.dart';

class ProfilePage extends ConsumerWidget {
  const ProfilePage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final authState = ref.watch(authProvider);
    final walletAsync = ref.watch(walletInfoProvider);
    final currentLocale = ref.watch(localeProvider);

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Text(l10n.profile),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () => context.pop(),
          ),
        ),
        body: SingleChildScrollView(
          padding: EdgeInsets.all(16.w),
          child: Column(
            children: [
              _buildUserInfoCard(context, authState, l10n),
              SizedBox(height: 16.h),
              _buildWalletSummary(context, walletAsync, l10n),
              SizedBox(height: 16.h),
              _buildSettingsList(context, ref, currentLocale, l10n),
              SizedBox(height: 24.h),
              _buildLogoutButton(context, ref, l10n),
            ],
          ),
        ),
      ),
    );
  }


  Widget _buildUserInfoCard(BuildContext context, dynamic authState, AppLocalizations l10n) {
    final isOwnerOrAdmin = authState.role == 'owner' || authState.role == 'admin';
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
                  radius: 40.r,
                  backgroundColor: AppColors.cardDark,
                  child: Text(
                    authState.username?.substring(0, 1).toUpperCase() ?? 'U',
                    style: TextStyle(
                      fontSize: 32.sp,
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
                      authState.username ?? 'User',
                      style: TextStyle(
                        fontSize: 22.sp,
                        fontWeight: FontWeight.bold,
                        color: AppColors.textPrimary,
                      ),
                    ),
                    SizedBox(height: 8.h),
                    _buildRoleBadge(authState.role, l10n),
                  ],
                ),
              ),
            ],
          ),
          // 房主/管理员显示邀请码
          if (isOwnerOrAdmin && authState.inviteCode != null) ...[
            SizedBox(height: 16.h),
            Divider(color: AppColors.glassBorder),
            SizedBox(height: 12.h),
            _buildInviteCodeSection(context, authState.inviteCode!, l10n),
          ],
        ],
      ),
    );
  }

  Widget _buildInviteCodeSection(BuildContext context, String inviteCode, AppLocalizations l10n) {
    return Container(
      padding: EdgeInsets.all(12.w),
      decoration: BoxDecoration(
        color: AppColors.primary.withValues(alpha: 0.1),
        borderRadius: BorderRadius.circular(12.r),
        border: Border.all(color: AppColors.primary.withValues(alpha: 0.3)),
      ),
      child: Row(
        children: [
          Icon(Icons.card_giftcard, color: AppColors.primary, size: 20.sp),
          SizedBox(width: 8.w),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  l10n.myInviteCode,
                  style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
                ),
                SizedBox(height: 4.h),
                Text(
                  inviteCode,
                  style: TextStyle(
                    fontSize: 18.sp,
                    fontWeight: FontWeight.bold,
                    color: AppColors.primary,
                    letterSpacing: 2,
                  ),
                ),
              ],
            ),
          ),
          IconButton(
            icon: Icon(Icons.copy, color: AppColors.primary, size: 20.sp),
            onPressed: () {
              _copyToClipboard(context, inviteCode, l10n);
            },
          ),
        ],
      ),
    );
  }

  void _copyToClipboard(BuildContext context, String text, AppLocalizations l10n) {
    Clipboard.setData(ClipboardData(text: text));
    ScaffoldMessenger.of(context).showSnackBar(
      SnackBar(content: Text('${l10n.inviteCodeCopied}: $text')),
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
      padding: EdgeInsets.symmetric(horizontal: 12.w, vertical: 4.h),
      decoration: BoxDecoration(
        color: color.withValues(alpha: 0.2),
        borderRadius: BorderRadius.circular(16.r),
        border: Border.all(color: color.withValues(alpha: 0.5)),
      ),
      child: Text(
        text,
        style: TextStyle(fontSize: 12.sp, color: color, fontWeight: FontWeight.w600),
      ),
    );
  }

  Widget _buildWalletSummary(
    BuildContext context,
    AsyncValue<WalletInfo> walletAsync,
    AppLocalizations l10n,
  ) {
    return GlassCard(
      padding: EdgeInsets.all(20.w),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.account_balance_wallet, color: AppColors.accent, size: 24.sp),
              SizedBox(width: 8.w),
              Text(
                l10n.wallet,
                style: TextStyle(
                  fontSize: 18.sp,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          SizedBox(height: 16.h),
          walletAsync.when(
            loading: () => _buildWalletShimmer(),
            error: (_, __) => Text(
              l10n.loadFailed,
              style: TextStyle(color: AppColors.error, fontSize: 14.sp),
            ),
            data: (wallet) => Column(
              children: [
                _buildBalanceRow(l10n.availableBalance, wallet.availableBalance, AppColors.success),
                SizedBox(height: 8.h),
                _buildBalanceRow(l10n.frozenBalance, wallet.frozenBalance, AppColors.warning),
                // 房主专属字段
                if (wallet.isOwner) ...[
                  SizedBox(height: 8.h),
                  _buildBalanceRow(l10n.marginDeposit, wallet.ownerMarginBalance, AppColors.primary),
                  SizedBox(height: 8.h),
                  _buildBalanceRow(l10n.commissionEarnings, wallet.ownerRoomBalance, AppColors.warning),
                ],
                Divider(color: AppColors.glassBorder, height: 24.h),
                _buildBalanceRow(l10n.totalBalance, wallet.totalBalance, AppColors.accent, isBold: true),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildWalletShimmer() {
    return Column(
      children: List.generate(3, (_) => Container(
        height: 20.h,
        margin: EdgeInsets.only(bottom: 8.h),
        decoration: BoxDecoration(
          color: AppColors.glassWhite,
          borderRadius: BorderRadius.circular(4.r),
        ),
      )),
    );
  }

  Widget _buildBalanceRow(String label, double value, Color color, {bool isBold = false}) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(
          label,
          style: TextStyle(
            fontSize: 14.sp,
            color: AppColors.textSecondary,
            fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
          ),
        ),
        Text(
          '¥${value.toStringAsFixed(2)}',
          style: TextStyle(
            fontSize: isBold ? 18.sp : 16.sp,
            color: color,
            fontWeight: isBold ? FontWeight.bold : FontWeight.w600,
          ),
        ),
      ],
    );
  }

  Widget _buildSettingsList(
    BuildContext context,
    WidgetRef ref,
    Locale currentLocale,
    AppLocalizations l10n,
  ) {
    return GlassCard(
      padding: EdgeInsets.all(16.w),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.settings, color: AppColors.primary, size: 24.sp),
              SizedBox(width: 8.w),
              Text(
                l10n.settings,
                style: TextStyle(
                  fontSize: 18.sp,
                  fontWeight: FontWeight.bold,
                  color: AppColors.textPrimary,
                ),
              ),
            ],
          ),
          SizedBox(height: 16.h),
          _buildLanguageSetting(context, ref, currentLocale, l10n),
        ],
      ),
    );
  }

  Widget _buildLanguageSetting(
    BuildContext context,
    WidgetRef ref,
    Locale currentLocale,
    AppLocalizations l10n,
  ) {
    return InkWell(
      onTap: () => _showLanguageDialog(context, ref, currentLocale, l10n),
      child: Padding(
        padding: EdgeInsets.symmetric(vertical: 12.h),
        child: Row(
          children: [
            Icon(Icons.language, color: AppColors.textSecondary, size: 20.sp),
            SizedBox(width: 12.w),
            Expanded(
              child: Text(
                l10n.language,
                style: TextStyle(fontSize: 16.sp, color: AppColors.textPrimary),
              ),
            ),
            Text(
              _getLanguageName(currentLocale.languageCode),
              style: TextStyle(fontSize: 14.sp, color: AppColors.textSecondary),
            ),
            SizedBox(width: 8.w),
            Icon(Icons.chevron_right, color: AppColors.textSecondary, size: 20.sp),
          ],
        ),
      ),
    );
  }

  String _getLanguageName(String code) {
    switch (code) {
      case 'zh':
        return '中文';
      case 'en':
        return 'English';
      case 'ja':
        return '日本語';
      case 'ko':
        return '한국어';
      default:
        return code;
    }
  }

  void _showLanguageDialog(BuildContext context, WidgetRef ref, Locale currentLocale, AppLocalizations l10n) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        backgroundColor: AppColors.cardDark,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
        title: Text(l10n.selectLanguage),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildLanguageOption(context, ref, 'zh', '中文', currentLocale),
            _buildLanguageOption(context, ref, 'en', 'English', currentLocale),
            _buildLanguageOption(context, ref, 'ja', '日本語', currentLocale),
            _buildLanguageOption(context, ref, 'ko', '한국어', currentLocale),
          ],
        ),
      ),
    );
  }

  Widget _buildLanguageOption(
    BuildContext context,
    WidgetRef ref,
    String code,
    String name,
    Locale currentLocale,
  ) {
    final isSelected = currentLocale.languageCode == code;
    return ListTile(
      title: Text(name),
      trailing: isSelected ? Icon(Icons.check, color: AppColors.accent) : null,
      onTap: () {
        ref.read(localeProvider.notifier).setLocale(Locale(code));
        Navigator.pop(context);
      },
    );
  }

  Widget _buildLogoutButton(BuildContext context, WidgetRef ref, AppLocalizations l10n) {
    return GradientButton(
      onPressed: () async {
        final confirm = await showDialog<bool>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            backgroundColor: AppColors.cardDark,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
            title: Text(l10n.confirmLogout),
            content: Text(l10n.confirmLogoutMessage),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: Text(l10n.cancel),
              ),
              GradientButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                gradient: AppColors.gradientWarm,
                height: 40.h,
                child: Padding(
                  padding: EdgeInsets.symmetric(horizontal: 16.w),
                  child: Text(l10n.confirm, style: const TextStyle(color: Colors.white)),
                ),
              ),
            ],
          ),
        );
        if (confirm == true && context.mounted) {
          await ref.read(authProvider.notifier).logout();
          if (context.mounted) context.go('/login');
        }
      },
      gradient: AppColors.gradientWarm,
      width: double.infinity,
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(Icons.logout, color: Colors.white, size: 20.sp),
          SizedBox(width: 8.w),
          Text(
            l10n.logout,
            style: TextStyle(
              color: Colors.white,
              fontSize: 16.sp,
              fontWeight: FontWeight.w600,
            ),
          ),
        ],
      ),
    );
  }
}
