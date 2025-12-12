import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import '../../../../core/theme/app_theme.dart';
import '../../../../core/services/api_client.dart';
import '../../../../l10n/app_localizations.dart';

/// 主题信息模型
class ThemeInfo {
  final String name;
  final String displayName;
  final Color primaryColor;
  final Color backgroundColor;
  final Color accentColor;

  ThemeInfo({
    required this.name,
    required this.displayName,
    required this.primaryColor,
    required this.backgroundColor,
    required this.accentColor,
  });

  factory ThemeInfo.fromJson(Map<String, dynamic> json) {
    return ThemeInfo(
      name: json['name'] as String? ?? '',
      displayName: json['display_name'] as String? ?? json['name'] as String? ?? '',
      primaryColor: _parseColor(json['primary_color']),
      backgroundColor: _parseColor(json['background_color']),
      accentColor: _parseColor(json['accent_color']),
    );
  }

  static Color _parseColor(dynamic value) {
    if (value == null) return AppColors.primary;
    if (value is String && value.startsWith('#')) {
      return Color(int.parse(value.substring(1), radix: 16) + 0xFF000000);
    }
    return AppColors.primary;
  }
}

/// 主题列表 Provider
final themesProvider = FutureProvider.autoDispose<List<ThemeInfo>>((ref) async {
  final apiClient = ref.read(apiClientProvider);
  final data = await apiClient.getAllThemes();
  return data.map((json) => ThemeInfo.fromJson(json)).toList();
});

class ThemeSelector extends ConsumerStatefulWidget {
  final int roomId;
  final String? currentTheme;
  final ValueChanged<String>? onThemeChanged;

  const ThemeSelector({
    super.key,
    required this.roomId,
    this.currentTheme,
    this.onThemeChanged,
  });

  @override
  ConsumerState<ThemeSelector> createState() => _ThemeSelectorState();
}

class _ThemeSelectorState extends ConsumerState<ThemeSelector> {
  String? _selectedTheme;
  bool _isUpdating = false;

  @override
  void initState() {
    super.initState();
    _selectedTheme = widget.currentTheme;
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final themesAsync = ref.watch(themesProvider);

    return AlertDialog(
      backgroundColor: AppColors.cardDark,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20.r)),
      title: Row(
        children: [
          Icon(Icons.palette, color: AppColors.accent, size: 24.sp),
          SizedBox(width: 8.w),
          Text(l10n.roomThemeTitle),
        ],
      ),
      content: SizedBox(
        width: 300.w,
        child: themesAsync.when(
          loading: () => SizedBox(
            height: 200.h,
            child: const Center(child: CircularProgressIndicator()),
          ),
          error: (error, _) => Text(
            '加载主题失败: $error',
            style: TextStyle(color: AppColors.error),
          ),
          data: (themes) => _buildThemeGrid(themes),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('取消'),
        ),
        GradientButton(
          onPressed: _isUpdating || _selectedTheme == null ? null : _applyTheme,
          height: 40.h,
          child: _isUpdating
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
                  child: const Text('应用', style: TextStyle(color: Colors.white)),
                ),
        ),
      ],
    );
  }

  Widget _buildThemeGrid(List<ThemeInfo> themes) {
    return GridView.builder(
      shrinkWrap: true,
      gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
        crossAxisCount: 2,
        crossAxisSpacing: 12.w,
        mainAxisSpacing: 12.h,
        childAspectRatio: 1.2,
      ),
      itemCount: themes.length,
      itemBuilder: (context, index) => _buildThemeCard(themes[index]),
    );
  }

  Widget _buildThemeCard(ThemeInfo theme) {
    final isSelected = _selectedTheme == theme.name;

    return InkWell(
      onTap: () => setState(() => _selectedTheme = theme.name),
      borderRadius: BorderRadius.circular(12.r),
      child: Container(
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [theme.backgroundColor, theme.primaryColor],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(12.r),
          border: Border.all(
            color: isSelected ? AppColors.accent : Colors.transparent,
            width: 2,
          ),
          boxShadow: isSelected
              ? [
                  BoxShadow(
                    color: AppColors.accent.withValues(alpha: 0.4),
                    blurRadius: 8,
                    offset: const Offset(0, 2),
                  ),
                ]
              : null,
        ),
        child: Stack(
          children: [
            Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Container(
                    width: 24.w,
                    height: 24.w,
                    decoration: BoxDecoration(
                      color: theme.accentColor,
                      shape: BoxShape.circle,
                    ),
                  ),
                  SizedBox(height: 8.h),
                  Text(
                    theme.displayName,
                    style: TextStyle(
                      fontSize: 12.sp,
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              Positioned(
                top: 8.h,
                right: 8.w,
                child: Container(
                  padding: EdgeInsets.all(2.w),
                  decoration: const BoxDecoration(
                    color: AppColors.accent,
                    shape: BoxShape.circle,
                  ),
                  child: Icon(Icons.check, color: Colors.white, size: 14.sp),
                ),
              ),
          ],
        ),
      ),
    );
  }

  Future<void> _applyTheme() async {
    if (_selectedTheme == null) return;

    setState(() => _isUpdating = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      await apiClient.updateRoomTheme(widget.roomId, _selectedTheme!);
      widget.onThemeChanged?.call(_selectedTheme!);
      if (mounted) {
        Navigator.pop(context);
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('主题已更新')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('更新主题失败: $e')),
        );
      }
    } finally {
      if (mounted) setState(() => _isUpdating = false);
    }
  }
}
