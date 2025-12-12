import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../l10n/app_localizations.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';

class CreateRoomPage extends ConsumerStatefulWidget {
  const CreateRoomPage({super.key});

  @override
  ConsumerState<CreateRoomPage> createState() => _CreateRoomPageState();
}

class _CreateRoomPageState extends ConsumerState<CreateRoomPage> {
  final _formKey = GlobalKey<FormState>();
  final _nameController = TextEditingController();
  final _passwordController = TextEditingController();

  double _betAmount = 5;
  int _winnerCount = 1;
  int _maxPlayers = 10;
  double _ownerCommission = 3; // 房主费率，百分比
  bool _hasPassword = false;
  bool _isLoading = false;

  final List<double> _betOptions = [5, 10, 20, 50, 100, 200];
  static const double _platformCommission = 2; // 平台费率，固定2%
  static const double _maxTotalCommission = 10; // 最大总费率10%

  @override
  void dispose() {
    _nameController.dispose();
    _passwordController.dispose();
    super.dispose();
  }

  Future<void> _createRoom() async {
    if (!_formKey.currentState!.validate()) return;

    final l10n = AppLocalizations.of(context)!;
    
    // 验证 winnerCount < maxPlayers（赢家数必须小于最大玩家数）
    if (_winnerCount >= _maxPlayers) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(l10n.winnersMustLessThanPlayers)),
      );
      return;
    }

    setState(() => _isLoading = true);

    try {
      final api = ref.read(apiClientProvider);
      await api.createRoom(
        name: _nameController.text.trim(),
        betAmount: _betAmount,
        winnerCount: _winnerCount,
        maxPlayers: _maxPlayers,
        ownerCommission: _ownerCommission / 100, // 转换为小数
        password: _hasPassword ? _passwordController.text : null,
      );

      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(l10n.success)),
        );
        context.pop();
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('${l10n.createFailed}: $e')),
        );
      }
    } finally {
      if (mounted) {
        setState(() => _isLoading = false);
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return GradientBackground(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        appBar: AppBar(
          title: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.add_circle, size: 24.sp, color: AppColors.accent),
              SizedBox(width: 8.w),
              Text(l10n.createRoom),
            ],
          ),
        ),
        body: SingleChildScrollView(
          padding: EdgeInsets.all(16.w),
          child: Form(
            key: _formKey,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 房间基本信息
                GlassCard(
                  padding: EdgeInsets.all(20.w),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSectionTitle(l10n.basicInfo, Icons.info_outline),
                      SizedBox(height: 16.h),
                      TextFormField(
                        controller: _nameController,
                        style: TextStyle(color: AppColors.textPrimary),
                        decoration: InputDecoration(
                          labelText: l10n.roomName,
                          prefixIcon: Icon(Icons.meeting_room, color: AppColors.accent),
                        ),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return l10n.pleaseEnterRoomName;
                          }
                          return null;
                        },
                      ),
                    ],
                  ),
                ),
                SizedBox(height: 16.h),

                // 游戏设置
                GlassCard(
                  padding: EdgeInsets.all(20.w),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSectionTitle(l10n.gameSettings, Icons.settings),
                      SizedBox(height: 16.h),

                      // 下注金额
                      Text('${l10n.betAmount}:', style: TextStyle(color: AppColors.textSecondary, fontSize: 14.sp)),
                      SizedBox(height: 8.h),
                      Wrap(
                        spacing: 8.w,
                        runSpacing: 8.h,
                        children: _betOptions.map((amount) => _buildBetChip(amount)).toList(),
                      ),
                      SizedBox(height: 20.h),

                      // 获胜人数
                      _buildNumberInput(
                        label: l10n.winnerCount,
                        value: _winnerCount,
                        min: 1,
                        max: _maxPlayers - 1, // 赢家数必须小于最大玩家数
                        onChanged: (v) => setState(() => _winnerCount = v),
                        icon: Icons.emoji_events,
                        hint: l10n.minPlayersHint(_winnerCount + 1),
                      ),
                      SizedBox(height: 16.h),

                      // 最大玩家数
                      _buildNumberInput(
                        label: l10n.maxPlayers,
                        value: _maxPlayers,
                        min: 2,
                        max: 100,
                        onChanged: (v) {
                          setState(() {
                            _maxPlayers = v;
                            // 确保赢家数小于最大玩家数
                            if (_winnerCount >= v) _winnerCount = v - 1;
                            if (_winnerCount < 1) _winnerCount = 1;
                          });
                        },
                        icon: Icons.people,
                        hint: l10n.maxPlayersHint,
                      ),
                    ],
                  ),
                ),
                SizedBox(height: 16.h),

                // 费率设置
                GlassCard(
                  padding: EdgeInsets.all(20.w),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSectionTitle(l10n.feeSettings, Icons.percent),
                      SizedBox(height: 16.h),

                      // 费率说明
                      Container(
                        padding: EdgeInsets.all(12.w),
                        decoration: BoxDecoration(
                          color: AppColors.primary.withValues(alpha: 0.1),
                          borderRadius: BorderRadius.circular(12.r),
                          border: Border.all(color: AppColors.primary.withValues(alpha: 0.3)),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(l10n.platformFeeFixed, style: TextStyle(color: AppColors.textSecondary, fontSize: 13.sp)),
                                Text('$_platformCommission%', style: TextStyle(color: AppColors.textPrimary, fontSize: 14.sp, fontWeight: FontWeight.bold)),
                              ],
                            ),
                            SizedBox(height: 8.h),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(l10n.ownerFee, style: TextStyle(color: AppColors.textSecondary, fontSize: 13.sp)),
                                Text('${_ownerCommission.toStringAsFixed(1)}%', style: TextStyle(color: AppColors.accent, fontSize: 14.sp, fontWeight: FontWeight.bold)),
                              ],
                            ),
                            Divider(color: AppColors.glassBorder, height: 20.h),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(l10n.totalFee, style: TextStyle(color: AppColors.textPrimary, fontSize: 14.sp, fontWeight: FontWeight.bold)),
                                Text(
                                  '${(_platformCommission + _ownerCommission).toStringAsFixed(1)}%',
                                  style: TextStyle(
                                    color: (_platformCommission + _ownerCommission) > _maxTotalCommission ? AppColors.error : AppColors.success,
                                    fontSize: 16.sp,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      ),
                      SizedBox(height: 16.h),

                      // 房主费率滑块
                      Text(l10n.adjustOwnerFee, style: TextStyle(color: AppColors.textSecondary, fontSize: 14.sp)),
                      SizedBox(height: 8.h),
                      Row(
                        children: [
                          Text('0%', style: TextStyle(color: AppColors.textSecondary, fontSize: 12.sp)),
                          Expanded(
                            child: Slider(
                              value: _ownerCommission,
                              min: 0,
                              max: _maxTotalCommission - _platformCommission,
                              divisions: 16, // 0.5% 步进
                              activeColor: AppColors.accent,
                              inactiveColor: AppColors.glassWhite,
                              onChanged: (v) => setState(() => _ownerCommission = (v * 2).round() / 2), // 四舍五入到0.5
                            ),
                          ),
                          Text('${(_maxTotalCommission - _platformCommission).toInt()}%', style: TextStyle(color: AppColors.textSecondary, fontSize: 12.sp)),
                        ],
                      ),
                      Text(
                        l10n.feeNote,
                        style: TextStyle(color: AppColors.textSecondary.withValues(alpha: 0.7), fontSize: 11.sp),
                      ),
                    ],
                  ),
                ),
                SizedBox(height: 16.h),

                // 密码设置
                GlassCard(
                  padding: EdgeInsets.all(20.w),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildSectionTitle(l10n.securitySettings, Icons.security),
                      SizedBox(height: 12.h),
                      Row(
                        children: [
                          Expanded(
                            child: Text(
                              l10n.roomPassword,
                              style: TextStyle(color: AppColors.textPrimary, fontSize: 16.sp),
                            ),
                          ),
                          Switch(
                            value: _hasPassword,
                            onChanged: (v) => setState(() => _hasPassword = v),
                            activeColor: AppColors.accent,
                          ),
                        ],
                      ),
                      if (_hasPassword) ...[
                        SizedBox(height: 12.h),
                        TextFormField(
                          controller: _passwordController,
                          obscureText: true,
                          style: TextStyle(color: AppColors.textPrimary),
                          decoration: InputDecoration(
                            labelText: l10n.setPassword,
                            prefixIcon: Icon(Icons.lock, color: AppColors.accent),
                          ),
                          validator: (value) {
                            if (_hasPassword && (value == null || value.isEmpty)) {
                              return l10n.pleaseEnterPassword;
                            }
                            return null;
                          },
                        ),
                      ],
                    ],
                  ),
                ),
                SizedBox(height: 24.h),

                // 创建按钮
                GradientButton(
                  onPressed: _isLoading ? null : _createRoom,
                  gradient: AppColors.gradientAccent,
                  width: double.infinity,
                  height: 56.h,
                  child: _isLoading
                      ? SizedBox(
                          width: 24.w,
                          height: 24.w,
                          child: const CircularProgressIndicator(strokeWidth: 2, color: Colors.white),
                        )
                      : Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Icon(Icons.add, color: Colors.white, size: 24.sp),
                            SizedBox(width: 8.w),
                            Text(
                              l10n.createRoom,
                              style: TextStyle(color: Colors.white, fontSize: 18.sp, fontWeight: FontWeight.bold),
                            ),
                          ],
                        ),
                ),
                SizedBox(height: 16.h),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildSectionTitle(String title, IconData icon) {
    return Row(
      children: [
        Icon(icon, color: AppColors.accent, size: 20.sp),
        SizedBox(width: 8.w),
        Text(title, style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary)),
      ],
    );
  }

  Widget _buildBetChip(double amount) {
    final isSelected = _betAmount == amount;
    return GestureDetector(
      onTap: () => setState(() => _betAmount = amount),
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: 20.w, vertical: 10.h),
        decoration: BoxDecoration(
          gradient: isSelected ? AppColors.gradientAccent : null,
          color: isSelected ? null : AppColors.glassWhite,
          borderRadius: BorderRadius.circular(20.r),
          border: Border.all(color: isSelected ? Colors.transparent : AppColors.glassBorder),
          boxShadow: isSelected
              ? [BoxShadow(color: AppColors.accent.withValues(alpha: 0.3), blurRadius: 8, offset: const Offset(0, 2))]
              : null,
        ),
        child: Text(
          '¥${amount.toInt()}',
          style: TextStyle(
            fontSize: 14.sp,
            color: isSelected ? Colors.white : AppColors.textSecondary,
            fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
          ),
        ),
      ),
    );
  }

  Widget _buildNumberInput({
    required String label,
    required int value,
    required int min,
    required int max,
    required ValueChanged<int> onChanged,
    required IconData icon,
    String? hint,
  }) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(icon, color: AppColors.accent, size: 18.sp),
            SizedBox(width: 8.w),
            Text(label, style: TextStyle(color: AppColors.textSecondary, fontSize: 14.sp)),
          ],
        ),
        if (hint != null) ...[
          SizedBox(height: 4.h),
          Text(hint, style: TextStyle(color: AppColors.textSecondary.withValues(alpha: 0.6), fontSize: 11.sp)),
        ],
        SizedBox(height: 8.h),
        Container(
          decoration: BoxDecoration(
            color: AppColors.glassWhite,
            borderRadius: BorderRadius.circular(12.r),
            border: Border.all(color: AppColors.glassBorder),
          ),
          child: Row(
            children: [
              IconButton(
                onPressed: value > min ? () => onChanged(value - 1) : null,
                icon: Icon(Icons.remove, color: value > min ? AppColors.accent : AppColors.textSecondary),
              ),
              Expanded(
                child: Text(
                  '$value',
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 20.sp, fontWeight: FontWeight.bold, color: AppColors.textPrimary),
                ),
              ),
              IconButton(
                onPressed: value < max ? () => onChanged(value + 1) : null,
                icon: Icon(Icons.add, color: value < max ? AppColors.accent : AppColors.textSecondary),
              ),
            ],
          ),
        ),
      ],
    );
  }
}
