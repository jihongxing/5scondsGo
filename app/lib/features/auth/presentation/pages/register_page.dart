import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../l10n/app_localizations.dart';

import '../../../../core/theme/app_theme.dart';
import '../../providers/auth_provider.dart';

class RegisterPage extends ConsumerStatefulWidget {
  const RegisterPage({super.key});

  @override
  ConsumerState<RegisterPage> createState() => _RegisterPageState();
}

class _RegisterPageState extends ConsumerState<RegisterPage> {
  final _formKey = GlobalKey<FormState>();
  final _usernameController = TextEditingController();
  final _passwordController = TextEditingController();
  final _inviteCodeController = TextEditingController();
  bool _isLoading = false;
  String _selectedRole = 'player'; // default role

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _inviteCodeController.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    if (!_formKey.currentState!.validate()) return;

    setState(() => _isLoading = true);

    try {
      await ref.read(authProvider.notifier).registerWithApi(
        username: _usernameController.text,
        password: _passwordController.text,
        inviteCode: _inviteCodeController.text,
        role: _selectedRole,
      );
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('Registration successful!')),
        );
        context.go('/home');
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(e.toString())),
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
        body: SafeArea(
          child: SingleChildScrollView(
            padding: EdgeInsets.all(24.w),
            child: Form(
              key: _formKey,
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: [
                  SizedBox(height: 40.h),

                  // Logo with glow effect
                  Center(
                    child: Container(
                      padding: EdgeInsets.all(16.w),
                      decoration: BoxDecoration(
                        shape: BoxShape.circle,
                        gradient: AppColors.gradientPrimary,
                        boxShadow: [
                          BoxShadow(
                            color: AppColors.primary.withValues(alpha: 0.5),
                            blurRadius: 30,
                            spreadRadius: 5,
                          ),
                        ],
                      ),
                      child: Icon(
                        Icons.person_add_rounded,
                        size: 48.sp,
                        color: Colors.white,
                      ),
                    ),
                  ),
                  SizedBox(height: 20.h),

                  // Title
                  ShaderMask(
                    shaderCallback: (bounds) => const LinearGradient(
                      colors: [AppColors.accent, AppColors.primary],
                    ).createShader(bounds),
                    child: Text(
                      l10n.register,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: 28.sp,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                        letterSpacing: 2,
                      ),
                    ),
                  ),
                  SizedBox(height: 8.h),
                  Text(
                    '创建您的游戏账户',
                    textAlign: TextAlign.center,
                    style: TextStyle(
                      fontSize: 14.sp,
                      color: AppColors.textSecondary,
                      letterSpacing: 2,
                    ),
                  ),
                  SizedBox(height: 32.h),

                  // Register form card
                  GlassCard(
                    padding: EdgeInsets.all(24.w),
                    child: Column(
                      children: [
                        // Username
                        TextFormField(
                          controller: _usernameController,
                          style: TextStyle(color: AppColors.textPrimary),
                          decoration: InputDecoration(
                            labelText: l10n.username,
                            prefixIcon: Icon(Icons.person_outline, color: AppColors.accent),
                          ),
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return '请输入用户名';
                            }
                            if (value.length < 3) {
                              return '用户名至少3个字符';
                            }
                            return null;
                          },
                        ),
                        SizedBox(height: 16.h),

                        // Password
                        TextFormField(
                          controller: _passwordController,
                          obscureText: true,
                          style: TextStyle(color: AppColors.textPrimary),
                          decoration: InputDecoration(
                            labelText: l10n.password,
                            prefixIcon: Icon(Icons.lock_outline, color: AppColors.accent),
                          ),
                          validator: (value) {
                            if (value == null || value.length < 6) {
                              return '密码至少6位';
                            }
                            return null;
                          },
                        ),
                        SizedBox(height: 16.h),

                        // Role Selection
                        DropdownButtonFormField<String>(
                          value: _selectedRole,
                          dropdownColor: AppColors.cardDark,
                          style: TextStyle(color: AppColors.textPrimary),
                          decoration: InputDecoration(
                            labelText: '注册角色',
                            prefixIcon: Icon(Icons.people_outline, color: AppColors.accent),
                          ),
                          items: [
                            DropdownMenuItem(
                              value: 'player',
                              child: Text('玩家 (需填房主邀请码)', style: TextStyle(color: AppColors.textPrimary)),
                            ),
                            DropdownMenuItem(
                              value: 'owner',
                              child: Text('房主 (需填系统/上级邀请码)', style: TextStyle(color: AppColors.textPrimary)),
                            ),
                          ],
                          onChanged: (value) {
                            if (value != null) {
                              setState(() => _selectedRole = value);
                            }
                          },
                        ),
                        SizedBox(height: 16.h),

                        // Invite Code
                        TextFormField(
                          controller: _inviteCodeController,
                          style: TextStyle(color: AppColors.textPrimary),
                          decoration: InputDecoration(
                            labelText: l10n.inviteCode,
                            prefixIcon: Icon(Icons.card_giftcard_outlined, color: AppColors.accent),
                            hintText: _selectedRole == 'owner'
                                ? '请输入管理员邀请码 (如: ADMIN1)'
                                : '请输入房主邀请码',
                            hintStyle: TextStyle(color: AppColors.textSecondary),
                          ),
                          validator: (value) {
                            if (value == null || value.isEmpty) {
                              return '请输入邀请码';
                            }
                            if (value.length != 6) {
                              return '邀请码必须是6位';
                            }
                            return null;
                          },
                        ),
                        SizedBox(height: 32.h),

                        // Register Button
                        GradientButton(
                          onPressed: _isLoading ? null : _register,
                          gradient: AppColors.gradientAccent,
                          width: double.infinity,
                          child: _isLoading
                              ? SizedBox(
                                  width: 20.w,
                                  height: 20.w,
                                  child: const CircularProgressIndicator(
                                    strokeWidth: 2,
                                    color: Colors.white,
                                  ),
                                )
                              : Text(
                                  l10n.registerButton,
                                  style: TextStyle(
                                    color: Colors.white,
                                    fontSize: 16.sp,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                        ),
                      ],
                    ),
                  ),
                  SizedBox(height: 24.h),

                  // Login Link
                  Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      Text(
                        l10n.hasAccount,
                        style: TextStyle(color: AppColors.textSecondary),
                      ),
                      TextButton(
                        onPressed: () => context.go('/login'),
                        child: Text(
                          l10n.login,
                          style: TextStyle(
                            color: AppColors.accent,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}
