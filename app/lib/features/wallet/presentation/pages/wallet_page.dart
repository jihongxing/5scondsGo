import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';

import '../../../../l10n/app_localizations.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';
import '../../../auth/providers/auth_provider.dart';
import '../../providers/wallet_providers.dart';

class WalletPage extends ConsumerStatefulWidget {
  const WalletPage({super.key});

  @override
  ConsumerState<WalletPage> createState() => _WalletPageState();
}

class _WalletPageState extends ConsumerState<WalletPage> {
  final _amountController = TextEditingController();
  final _remarkController = TextEditingController();
  final _transferAmountController = TextEditingController();
  String? _type;
  bool _submitting = false;
  bool _transferring = false;
  bool _initialized = false;

  @override
  void dispose() {
    _amountController.dispose();
    _remarkController.dispose();
    _transferAmountController.dispose();
    super.dispose();
  }

  void _initType(bool isOwner) {
    if (!_initialized) {
      _type = isOwner ? 'owner_deposit' : 'deposit';
      _initialized = true;
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final authState = ref.watch(authProvider);
    final isOwner = authState.role == 'owner';
    _initType(isOwner);
    final walletAsync = ref.watch(walletInfoProvider);
    final requestsAsync = ref.watch(fundRequestsProvider);
    final txsAsync = ref.watch(transactionsProvider);

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.account_balance_wallet, size: 24.sp, color: AppColors.accent),
              SizedBox(width: 8.w),
              Text(l10n.wallet),
            ],
          ),
          leading: IconButton(
            icon: const Icon(Icons.arrow_back),
            onPressed: () => context.go('/home'),
          ),
        ),
        body: RefreshIndicator(
          onRefresh: () async {
            ref.invalidate(walletInfoProvider);
            ref.invalidate(fundRequestsProvider);
            ref.invalidate(transactionsProvider);
          },
          child: SingleChildScrollView(
            physics: const AlwaysScrollableScrollPhysics(),
            padding: EdgeInsets.all(16.w),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                walletAsync.when(
                  loading: () => _buildWalletShimmer(),
                  error: (error, _) => GlassCard(
                    child: Text(error.toString(), style: TextStyle(color: AppColors.error)),
                  ),
                  data: (wallet) => isOwner
                      ? _buildOwnerWalletCard(wallet)
                      : _buildPlayerWalletCard(wallet, l10n),
                ),
                SizedBox(height: 24.h),
                _buildFundRequestForm(l10n, isOwner),
                if (isOwner) ...[
                  SizedBox(height: 24.h),
                  _buildOwnerPendingRequests(l10n),
                ],
                SizedBox(height: 24.h),
                _buildFundRequestsList(l10n, requestsAsync),
                SizedBox(height: 24.h),
                _buildTransactionsList(l10n, txsAsync),
              ],
            ),
          ),
        ),
      ),
    );
  }


  Widget _buildWalletShimmer() {
    return GlassCard(
      padding: EdgeInsets.all(20.w),
      child: Column(
        children: [
          Row(
            children: [
              Expanded(child: Container(height: 80.h, decoration: BoxDecoration(color: AppColors.glassWhite, borderRadius: BorderRadius.circular(12.r)))),
              SizedBox(width: 12.w),
              Expanded(child: Container(height: 80.h, decoration: BoxDecoration(color: AppColors.glassWhite, borderRadius: BorderRadius.circular(12.r)))),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPlayerWalletCard(WalletInfo wallet, AppLocalizations l10n) {
    return GlassCard(
      padding: EdgeInsets.all(20.w),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              NeonIcon(icon: Icons.account_balance_wallet, color: AppColors.accent, size: 24.sp),
              SizedBox(width: 8.w),
              Text(l10n.wallet, style: TextStyle(fontSize: 18.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
            ],
          ),
          SizedBox(height: 20.h),
          Row(
            children: [
              Expanded(child: _buildBalanceItem(l10n.availableBalance, wallet.availableBalance, AppColors.success, Icons.account_balance)),
              SizedBox(width: 12.w),
              Expanded(child: _buildBalanceItem(l10n.frozenBalance, wallet.frozenBalance, AppColors.warning, Icons.lock)),
            ],
          ),
          SizedBox(height: 12.h),
          _buildTotalBalance(l10n.totalBalance, wallet.totalBalance),
        ],
      ),
    );
  }

  Widget _buildOwnerWalletCard(WalletInfo wallet) {
    final l10n = AppLocalizations.of(context)!;
    return GlassCard(
      padding: EdgeInsets.all(20.w),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              NeonIcon(icon: Icons.business_center, color: AppColors.warning, size: 24.sp),
              SizedBox(width: 8.w),
              Text(l10n.ownerWallet, style: TextStyle(fontSize: 18.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
            ],
          ),
          SizedBox(height: 20.h),
          // 第一行：可用余额 + 保证金
          Row(
            children: [
              Expanded(child: _buildBalanceItem(l10n.availableBalance, wallet.availableBalance, AppColors.success, Icons.account_balance_wallet, subtitle: l10n.canRechargePlayer)),
              SizedBox(width: 12.w),
              Expanded(child: _buildBalanceItem(l10n.marginBalance, wallet.ownerMarginBalance, AppColors.primary, Icons.security, subtitle: l10n.fixedGuarantee)),
            ],
          ),
          SizedBox(height: 12.h),
          // 第二行：佣金收益 + 累计佣金
          Row(
            children: [
              Expanded(child: _buildBalanceItem(l10n.commissionBalance, wallet.ownerRoomBalance, AppColors.warning, Icons.trending_up, subtitle: l10n.canTransferToBalance)),
              SizedBox(width: 12.w),
              Expanded(child: _buildBalanceItem(l10n.totalCommission, wallet.ownerTotalCommission, AppColors.accent, Icons.monetization_on)),
            ],
          ),
          // 佣金转余额功能
          if (wallet.ownerRoomBalance > 0) ...[
            SizedBox(height: 20.h),
            Divider(color: AppColors.glassBorder),
            SizedBox(height: 16.h),
            _buildTransferSection(wallet.ownerRoomBalance),
          ],
        ],
      ),
    );
  }

  Widget _buildBalanceItem(String title, double value, Color color, IconData icon, {String? subtitle}) {
    return Container(
      padding: EdgeInsets.all(16.w),
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [color.withValues(alpha: 0.15), color.withValues(alpha: 0.05)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.circular(16.r),
        border: Border.all(color: color.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 18.sp, color: color),
              SizedBox(width: 6.w),
              Expanded(child: Text(title, style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary), overflow: TextOverflow.ellipsis)),
            ],
          ),
          SizedBox(height: 8.h),
          AnimatedNumber(value: value, prefix: '¥', style: TextStyle(fontSize: 20.sp, fontWeight: FontWeight.bold, color: color)),
          if (subtitle != null) ...[
            SizedBox(height: 4.h),
            Text(subtitle, style: TextStyle(fontSize: 10.sp, color: AppColors.textSecondary)),
          ],
        ],
      ),
    );
  }

  Widget _buildTotalBalance(String title, double value) {
    return Container(
      padding: EdgeInsets.all(16.w),
      decoration: BoxDecoration(
        gradient: AppColors.gradientPrimary,
        borderRadius: BorderRadius.circular(16.r),
        boxShadow: [BoxShadow(color: AppColors.primary.withValues(alpha: 0.3), blurRadius: 15, offset: const Offset(0, 5))],
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(title, style: TextStyle(fontSize: 14.sp, color: Colors.white70)),
          AnimatedNumber(value: value, prefix: '¥', style: TextStyle(fontSize: 24.sp, fontWeight: FontWeight.bold, color: Colors.white)),
        ],
      ),
    );
  }

  Widget _buildTransferSection(double maxAmount) {
    final l10n = AppLocalizations.of(context)!;
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(l10n.transferToBalance, style: TextStyle(fontSize: 14.sp, fontWeight: FontWeight.w600, color: AppColors.textPrimary)),
        SizedBox(height: 12.h),
        Row(
          children: [
            Expanded(
              child: TextField(
                controller: _transferAmountController,
                keyboardType: const TextInputType.numberWithOptions(decimal: true),
                decoration: InputDecoration(
                  labelText: l10n.transferAmount,
                  hintText: l10n.maxTransfer(maxAmount.toStringAsFixed(2)),
                  prefixIcon: const Icon(Icons.currency_yuan),
                  isDense: true,
                ),
              ),
            ),
            SizedBox(width: 12.w),
            GradientButton(
              onPressed: _transferring ? null : () => _transferEarnings(maxAmount),
              gradient: AppColors.gradientAccent,
              height: 48.h,
              child: _transferring
                  ? SizedBox(width: 20.w, height: 20.w, child: const CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                  : Padding(padding: EdgeInsets.symmetric(horizontal: 16.w), child: Text(l10n.transfer, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600))),
            ),
          ],
        ),
        SizedBox(height: 8.h),
        Text(l10n.transferNote, style: TextStyle(fontSize: 11.sp, color: AppColors.textSecondary)),
      ],
    );
  }


  Widget _buildFundRequestForm(AppLocalizations l10n, bool isOwner) {
    return GlassCard(
      padding: EdgeInsets.all(20.w),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              NeonIcon(icon: Icons.add_card, color: AppColors.primary, size: 24.sp),
              SizedBox(width: 8.w),
              Text(l10n.submitFundRequest, style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
            ],
          ),
          SizedBox(height: 16.h),
          Wrap(
            spacing: 8.w,
            runSpacing: 8.h,
            children: isOwner
                ? [
                    _buildTypeChip(l10n.ownerDeposit, 'owner_deposit'),
                    _buildTypeChip(l10n.ownerWithdraw, 'owner_withdraw'),
                    _buildTypeChip(l10n.marginDeposit2, 'margin_deposit'),
                  ]
                : [
                    _buildTypeChip(l10n.deposit, 'deposit'),
                    _buildTypeChip(l10n.withdraw, 'withdraw'),
                  ],
          ),
          SizedBox(height: 16.h),
          TextField(
            controller: _amountController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(labelText: '${l10n.amount} (¥)', prefixIcon: const Icon(Icons.currency_yuan)),
          ),
          SizedBox(height: 12.h),
          TextField(
            controller: _remarkController,
            decoration: InputDecoration(labelText: l10n.remarkOptional, prefixIcon: const Icon(Icons.note_outlined)),
          ),
          SizedBox(height: 16.h),
          GradientButton(
            onPressed: _submitting ? null : () => _submit(l10n),
            gradient: AppColors.gradientPrimary,
            width: double.infinity,
            child: _submitting
                ? SizedBox(width: 20.w, height: 20.w, child: const CircularProgressIndicator(strokeWidth: 2, color: Colors.white))
                : Text(l10n.submitRequest, style: const TextStyle(color: Colors.white, fontWeight: FontWeight.w600)),
          ),
          SizedBox(height: 8.h),
          Text(l10n.fundRequestNote, style: TextStyle(fontSize: 11.sp, color: AppColors.textSecondary)),
        ],
      ),
    );
  }

  Widget _buildTypeChip(String label, String value) {
    final isSelected = _type == value;
    return GestureDetector(
      onTap: () => setState(() => _type = value),
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: 16.w, vertical: 8.h),
        decoration: BoxDecoration(
          gradient: isSelected ? AppColors.gradientPrimary : null,
          color: isSelected ? null : AppColors.glassWhite,
          borderRadius: BorderRadius.circular(20.r),
          border: Border.all(color: isSelected ? Colors.transparent : AppColors.glassBorder),
          boxShadow: isSelected ? [BoxShadow(color: AppColors.primary.withValues(alpha: 0.3), blurRadius: 8, offset: const Offset(0, 2))] : null,
        ),
        child: Text(label, style: TextStyle(fontSize: 13.sp, color: isSelected ? Colors.white : AppColors.textSecondary, fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal)),
      ),
    );
  }

  Widget _buildFundRequestsList(AppLocalizations l10n, AsyncValue<List<Map<String, dynamic>>> requestsAsync) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            NeonIcon(icon: Icons.receipt_long, color: AppColors.accent, size: 20.sp),
            SizedBox(width: 8.w),
            Text(l10n.fundRequestRecords, style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
          ],
        ),
        SizedBox(height: 12.h),
        requestsAsync.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, _) => Text(error.toString(), style: TextStyle(color: AppColors.error)),
          data: (items) {
            if (items.isEmpty) return GlassCard(child: Center(child: Text(l10n.noData, style: TextStyle(color: AppColors.textSecondary))));
            return GlassCard(
              padding: EdgeInsets.zero,
              child: ListView.separated(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: items.length,
                separatorBuilder: (_, __) => Divider(height: 1, color: AppColors.glassBorder),
                itemBuilder: (context, index) {
                  final item = items[index];
                  final type = item['type']?.toString() ?? '';
                  final amount = item['amount']?.toString() ?? '0';
                  final status = item['status']?.toString() ?? '';
                  final isDeposit = type.contains('deposit');
                  return ListTile(
                    leading: Container(
                      padding: EdgeInsets.all(8.w),
                      decoration: BoxDecoration(color: (isDeposit ? AppColors.success : AppColors.error).withValues(alpha: 0.2), shape: BoxShape.circle),
                      child: Icon(isDeposit ? Icons.arrow_downward : Icons.arrow_upward, color: isDeposit ? AppColors.success : AppColors.error, size: 20.sp),
                    ),
                    title: Text('¥$amount', style: TextStyle(fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
                    subtitle: Text(_getTypeDisplay(type, l10n), style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary)),
                    trailing: _buildStatusBadge(status, l10n),
                  );
                },
              ),
            );
          },
        ),
      ],
    );
  }

  Widget _buildStatusBadge(String status, AppLocalizations l10n) {
    Color color;
    String text;
    switch (status) {
      case 'pending':
        color = AppColors.warning;
        text = l10n.pending;
      case 'approved':
        color = AppColors.success;
        text = l10n.approved;
      case 'rejected':
        color = AppColors.error;
        text = l10n.rejected;
      default:
        color = AppColors.textSecondary;
        text = status;
    }
    return Container(
      padding: EdgeInsets.symmetric(horizontal: 10.w, vertical: 4.h),
      decoration: BoxDecoration(color: color.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(12.r)),
      child: Text(text, style: TextStyle(fontSize: 12.sp, color: color, fontWeight: FontWeight.w500)),
    );
  }

  String _getTypeDisplay(String type, AppLocalizations l10n) {
    switch (type) {
      case 'deposit': return l10n.deposit;
      case 'withdraw': return l10n.withdraw;
      case 'owner_deposit': return l10n.ownerDeposit;
      case 'owner_withdraw': return l10n.ownerWithdraw;
      case 'margin_deposit': return l10n.marginDeposit2;
      default: return type;
    }
  }


  Widget _buildTransactionsList(AppLocalizations l10n, AsyncValue<List<Map<String, dynamic>>> txsAsync) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            NeonIcon(icon: Icons.swap_horiz, color: AppColors.primary, size: 20.sp),
            SizedBox(width: 8.w),
            Text(l10n.transactionHistory, style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
          ],
        ),
        SizedBox(height: 12.h),
        txsAsync.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, _) => Text(error.toString(), style: TextStyle(color: AppColors.error)),
          data: (items) {
            if (items.isEmpty) return GlassCard(child: Center(child: Text(l10n.noData, style: TextStyle(color: AppColors.textSecondary))));
            return GlassCard(
              padding: EdgeInsets.zero,
              child: ListView.separated(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: items.length,
                separatorBuilder: (_, __) => Divider(height: 1, color: AppColors.glassBorder),
                itemBuilder: (context, index) {
                  final item = items[index];
                  final type = item['type']?.toString() ?? '';
                  final amount = item['amount']?.toString() ?? '0';
                  final createdAt = item['created_at']?.toString() ?? '';
                  final isPositive = type == 'game_win' || type == 'deposit' || type == 'owner_commission';
                  return ListTile(
                    leading: Container(
                      padding: EdgeInsets.all(8.w),
                      decoration: BoxDecoration(color: (isPositive ? AppColors.success : AppColors.error).withValues(alpha: 0.2), shape: BoxShape.circle),
                      child: Icon(isPositive ? Icons.add : Icons.remove, color: isPositive ? AppColors.success : AppColors.error, size: 20.sp),
                    ),
                    title: Text('${isPositive ? '+' : '-'}¥$amount', style: TextStyle(fontWeight: FontWeight.bold, color: isPositive ? AppColors.success : AppColors.error)),
                    subtitle: Text(_getTxTypeDisplay(type, l10n), style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary)),
                    trailing: Text(_formatDate(createdAt), style: TextStyle(fontSize: 11.sp, color: AppColors.textSecondary)),
                  );
                },
              ),
            );
          },
        ),
      ],
    );
  }

  String _getTxTypeDisplay(String type, AppLocalizations l10n) {
    switch (type) {
      case 'deposit': return l10n.deposit;
      case 'withdraw': return l10n.withdraw;
      case 'game_bet': return l10n.gameBet;
      case 'game_win': return l10n.gameWin;
      case 'game_refund': return l10n.gameRefund;
      case 'owner_commission': return l10n.ownerCommission;
      case 'platform_share': return l10n.platformShare;
      default: return type;
    }
  }

  String _formatDate(String dateStr) {
    try {
      final date = DateTime.parse(dateStr);
      return '${date.month}/${date.day} ${date.hour}:${date.minute.toString().padLeft(2, '0')}';
    } catch (_) {
      return dateStr;
    }
  }

  Future<void> _submit(AppLocalizations l10n) async {
    final amountText = _amountController.text.trim();
    if (amountText.isEmpty) {
      _showSnackBar(l10n.pleaseEnterAmount);
      return;
    }
    final amount = double.tryParse(amountText);
    if (amount == null || amount <= 0) {
      _showSnackBar(l10n.invalidAmountFormat);
      return;
    }
    if (_type == null) {
      _showSnackBar(l10n.pleaseSelectType);
      return;
    }

    setState(() => _submitting = true);
    try {
      await ref.read(apiClientProvider).createFundRequest(
        type: _type!,
        amount: amount,
        remark: _remarkController.text.trim().isEmpty ? null : _remarkController.text.trim(),
      );
      _amountController.clear();
      _remarkController.clear();
      ref.invalidate(fundRequestsProvider);
      ref.invalidate(fundSummaryProvider);
      _showSnackBar(l10n.fundRequestSubmitted);
    } catch (e) {
      _showSnackBar('${l10n.submitFailed}: $e');
    } finally {
      if (mounted) setState(() => _submitting = false);
    }
  }

  Future<void> _transferEarnings(double maxAmount) async {
    final l10n = AppLocalizations.of(context)!;
    final amountText = _transferAmountController.text.trim();
    if (amountText.isEmpty) {
      _showSnackBar(l10n.pleaseEnterTransferAmount);
      return;
    }
    final amount = double.tryParse(amountText);
    if (amount == null || amount <= 0) {
      _showSnackBar(l10n.invalidAmountFormat);
      return;
    }
    if (amount > maxAmount) {
      _showSnackBar(l10n.transferAmountExceeds(maxAmount.toStringAsFixed(2)));
      return;
    }

    setState(() => _transferring = true);
    try {
      await ref.read(apiClientProvider).transferEarnings(amount);
      _transferAmountController.clear();
      ref.invalidate(walletInfoProvider);
      _showSnackBar(l10n.transferSuccess);
    } catch (e) {
      _showSnackBar('${l10n.transferFailed}: $e');
    } finally {
      if (mounted) setState(() => _transferring = false);
    }
  }

  void _showSnackBar(String message) {
    if (mounted) {
      ScaffoldMessenger.of(context).showSnackBar(SnackBar(content: Text(message)));
    }
  }

  Widget _buildOwnerPendingRequests(AppLocalizations l10n) {
    final pendingAsync = ref.watch(ownerPendingFundRequestsProvider);
    
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(Icons.pending_actions, color: AppColors.warning, size: 20.sp),
            SizedBox(width: 8.w),
            Text(
              l10n.pendingPlayerRequests,
              style: TextStyle(
                fontSize: 16.sp,
                fontWeight: FontWeight.bold,
                color: AppColors.textPrimary,
              ),
            ),
            const Spacer(),
            IconButton(
              icon: Icon(Icons.refresh, size: 20.sp),
              onPressed: () => ref.invalidate(ownerPendingFundRequestsProvider),
            ),
          ],
        ),
        SizedBox(height: 12.h),
        pendingAsync.when(
          loading: () => const Center(child: CircularProgressIndicator()),
          error: (error, _) => GlassCard(
            child: Text('${l10n.loadFailed}: $error', style: TextStyle(color: AppColors.error)),
          ),
          data: (requests) {
            if (requests.isEmpty) {
              return GlassCard(
                padding: EdgeInsets.all(16.w),
                child: Center(
                  child: Text(
                    l10n.noPendingRequests,
                    style: TextStyle(color: AppColors.textSecondary, fontSize: 14.sp),
                  ),
                ),
              );
            }
            return Column(
              children: requests.map((req) => _buildPendingRequestItem(req, l10n)).toList(),
            );
          },
        ),
      ],
    );
  }

  Widget _buildPendingRequestItem(Map<String, dynamic> req, AppLocalizations l10n) {
    final type = req['type'] as String? ?? '';
    final amount = double.tryParse(req['amount']?.toString() ?? '0') ?? 0;
    final username = req['username'] as String? ?? l10n.unknownUser;
    final createdAt = req['created_at'] as String? ?? '';
    final requestId = req['id'] as int? ?? 0;

    String typeText;
    Color typeColor;
    switch (type) {
      case 'deposit':
        typeText = l10n.deposit;
        typeColor = AppColors.success;
        break;
      case 'withdraw':
        typeText = l10n.withdraw;
        typeColor = AppColors.warning;
        break;
      default:
        typeText = type;
        typeColor = AppColors.textSecondary;
    }

    return GlassCard(
      margin: EdgeInsets.only(bottom: 8.h),
      padding: EdgeInsets.all(12.w),
      child: Row(
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Container(
                      padding: EdgeInsets.symmetric(horizontal: 8.w, vertical: 2.h),
                      decoration: BoxDecoration(
                        color: typeColor.withOpacity(0.2),
                        borderRadius: BorderRadius.circular(4.r),
                      ),
                      child: Text(
                        typeText,
                        style: TextStyle(color: typeColor, fontSize: 12.sp, fontWeight: FontWeight.bold),
                      ),
                    ),
                    SizedBox(width: 8.w),
                    Text(
                      username,
                      style: TextStyle(fontSize: 14.sp, fontWeight: FontWeight.w500),
                    ),
                  ],
                ),
                SizedBox(height: 4.h),
                Text(
                  '¥${amount.toStringAsFixed(2)}',
                  style: TextStyle(
                    fontSize: 18.sp,
                    fontWeight: FontWeight.bold,
                    color: AppColors.accent,
                  ),
                ),
                if (createdAt.isNotEmpty)
                  Text(
                    _formatDate(createdAt),
                    style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary),
                  ),
              ],
            ),
          ),
          Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              IconButton(
                icon: Icon(Icons.close, color: AppColors.error, size: 24.sp),
                onPressed: () => _processRequest(requestId, false, l10n),
                tooltip: l10n.reject,
              ),
              SizedBox(width: 8.w),
              IconButton(
                icon: Icon(Icons.check, color: AppColors.success, size: 24.sp),
                onPressed: () => _processRequest(requestId, true, l10n),
                tooltip: l10n.approve,
              ),
            ],
          ),
        ],
      ),
    );
  }

  Future<void> _processRequest(int requestId, bool approved, AppLocalizations l10n) async {
    final action = approved ? l10n.approve : l10n.reject;
    final confirmTitle = approved ? l10n.confirmApprove : l10n.confirmReject;
    final confirmMessage = approved ? l10n.confirmApproveMessage : l10n.confirmRejectMessage;
    final confirm = await showDialog<bool>(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(confirmTitle),
        content: Text(confirmMessage),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context, false),
            child: Text(l10n.cancel),
          ),
          TextButton(
            onPressed: () => Navigator.pop(context, true),
            child: Text(action),
          ),
        ],
      ),
    );

    if (confirm != true) return;

    try {
      await ref.read(apiClientProvider).processOwnerFundRequest(requestId, approved: approved);
      ref.invalidate(ownerPendingFundRequestsProvider);
      ref.invalidate(walletInfoProvider);
      _showSnackBar(approved ? l10n.requestApproved : l10n.requestRejected);
    } catch (e) {
      _showSnackBar('${l10n.operationFailed}: $e');
    }
  }
}
