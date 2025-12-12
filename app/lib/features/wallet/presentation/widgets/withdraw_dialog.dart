import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/services/api_client.dart';
import '../../providers/wallet_providers.dart';

class WithdrawDialog extends ConsumerStatefulWidget {
  final double availableBalance;

  const WithdrawDialog({super.key, required this.availableBalance});

  @override
  ConsumerState<WithdrawDialog> createState() => _WithdrawDialogState();
}

class _WithdrawDialogState extends ConsumerState<WithdrawDialog> {
  final _amountController = TextEditingController();
  final _remarkController = TextEditingController();
  bool _isSubmitting = false;
  String? _error;

  @override
  void dispose() {
    _amountController.dispose();
    _remarkController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('申请提现'),
      content: SingleChildScrollView(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 可用余额提示
            Container(
              padding: EdgeInsets.all(12.w),
              decoration: BoxDecoration(
                color: Colors.blue.withOpacity(0.1),
                borderRadius: BorderRadius.circular(8.r),
              ),
              child: Row(
                children: [
                  Icon(Icons.account_balance_wallet, color: Colors.blue, size: 20.sp),
                  SizedBox(width: 8.w),
                  Text(
                    '可用余额: ¥${widget.availableBalance.toStringAsFixed(2)}',
                    style: TextStyle(
                      fontSize: 14.sp,
                      fontWeight: FontWeight.bold,
                      color: Colors.blue,
                    ),
                  ),
                ],
              ),
            ),
            SizedBox(height: 16.h),

            // 金额输入
            TextField(
              controller: _amountController,
              keyboardType: const TextInputType.numberWithOptions(decimal: true),
              decoration: InputDecoration(
                labelText: '提现金额',
                prefixIcon: const Icon(Icons.currency_yuan),
                suffixIcon: TextButton(
                  onPressed: () {
                    _amountController.text = widget.availableBalance.toStringAsFixed(2);
                  },
                  child: const Text('全部'),
                ),
              ),
            ),
            SizedBox(height: 12.h),

            // 备注输入
            TextField(
              controller: _remarkController,
              decoration: const InputDecoration(
                labelText: '备注（可选）',
                prefixIcon: Icon(Icons.note_outlined),
              ),
              maxLines: 2,
            ),
            SizedBox(height: 8.h),

            // 错误提示
            if (_error != null)
              Padding(
                padding: EdgeInsets.only(top: 8.h),
                child: Text(
                  _error!,
                  style: TextStyle(color: Colors.red, fontSize: 12.sp),
                ),
              ),

            // 提示信息
            SizedBox(height: 12.h),
            Text(
              '提现申请提交后，需要管理员审核通过后才会到账。',
              style: TextStyle(
                fontSize: 12.sp,
                color: Colors.grey,
              ),
            ),
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _isSubmitting ? null : () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        ElevatedButton(
          onPressed: _isSubmitting ? null : _submit,
          child: _isSubmitting
              ? SizedBox(
                  width: 16.w,
                  height: 16.w,
                  child: const CircularProgressIndicator(strokeWidth: 2),
                )
              : const Text('提交申请'),
        ),
      ],
    );
  }

  Future<void> _submit() async {
    final amountText = _amountController.text.trim();
    if (amountText.isEmpty) {
      setState(() => _error = '请输入提现金额');
      return;
    }

    final amount = double.tryParse(amountText);
    if (amount == null || amount <= 0) {
      setState(() => _error = '金额格式不正确');
      return;
    }

    if (amount > widget.availableBalance) {
      setState(() => _error = '提现金额不能超过可用余额');
      return;
    }

    setState(() {
      _isSubmitting = true;
      _error = null;
    });

    try {
      final api = ref.read(apiClientProvider);
      await api.createFundRequest(
        type: 'withdraw',
        amount: amount,
        remark: _remarkController.text.trim().isEmpty
            ? null
            : _remarkController.text.trim(),
      );

      // 刷新数据
      ref.invalidate(walletInfoProvider);
      ref.invalidate(fundRequestsProvider);

      if (mounted) {
        Navigator.pop(context, true);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('提现申请已提交，等待审核')),
        );
      }
    } catch (e) {
      setState(() {
        _isSubmitting = false;
        _error = '提交失败: $e';
      });
    }
  }
}

/// 显示提现对话框
Future<bool?> showWithdrawDialog(BuildContext context, double availableBalance) {
  return showDialog<bool>(
    context: context,
    builder: (context) => WithdrawDialog(availableBalance: availableBalance),
  );
}
