import 'package:flutter/material.dart';
import 'package:flutter_localizations/flutter_localizations.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';

import 'features/dashboard/presentation/pages/dashboard_page.dart';
import 'features/users/presentation/pages/users_page.dart';
import 'features/rooms/presentation/pages/rooms_page.dart';
import 'features/funds/presentation/pages/funds_page.dart';
import 'features/auth/presentation/pages/login_page.dart';
import 'features/auth/providers/auth_provider.dart';
import 'features/monitoring/presentation/pages/monitoring_dashboard_page.dart';
import 'features/risk/presentation/pages/risk_flags_page.dart';
import 'features/alerts/presentation/pages/alerts_page.dart';

// 当前选择的语言（null 表示跟随浏览器 / 系统）
final localeProvider = StateProvider<Locale?>((ref) => null);

void main() {
  runApp(const ProviderScope(child: AdminApp()));
}

class AdminApp extends ConsumerWidget {
  const AdminApp({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final locale = ref.watch(localeProvider);

    return MaterialApp.router(
      onGenerateTitle: (context) => AppLocalizations.of(context)!.appTitle,
      locale: locale,
      supportedLocales: const [
        Locale('en'),
        Locale('zh'),
      ],
      localizationsDelegates: const [
        AppLocalizations.delegate,
        GlobalMaterialLocalizations.delegate,
        GlobalWidgetsLocalizations.delegate,
        GlobalCupertinoLocalizations.delegate,
      ],
      debugShowCheckedModeBanner: false,
      theme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF6C5CE7),
          brightness: Brightness.light,
        ),
      ),
      darkTheme: ThemeData(
        useMaterial3: true,
        colorScheme: ColorScheme.fromSeed(
          seedColor: const Color(0xFF6C5CE7),
          brightness: Brightness.dark,
        ),
      ),
      routerConfig: _router,
    );
  }
}

final _router = GoRouter(
  initialLocation: '/login',
  routes: [
    GoRoute(
      path: '/login',
      builder: (context, state) => const AdminLoginPage(),
    ),
    ShellRoute(
      builder: (context, state, child) => AdminShell(child: child),
      routes: [
        GoRoute(
          path: '/dashboard',
          builder: (context, state) => const DashboardPage(),
        ),
        GoRoute(
          path: '/users',
          builder: (context, state) => const UsersPage(),
        ),
        GoRoute(
          path: '/rooms',
          builder: (context, state) => const RoomsPage(),
        ),
        GoRoute(
          path: '/funds',
          builder: (context, state) => const FundsPage(),
        ),
        GoRoute(
          path: '/monitoring',
          builder: (context, state) => const MonitoringDashboardPage(),
        ),
        GoRoute(
          path: '/risk-flags',
          builder: (context, state) => const RiskFlagsPage(),
        ),
        GoRoute(
          path: '/alerts',
          builder: (context, state) => const AlertsPage(),
        ),
      ],
    ),
  ],
);

class AdminShell extends StatelessWidget {
  final Widget child;

  const AdminShell({super.key, required this.child});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      body: Row(
        children: [
          NavigationRail(
            selectedIndex: _getSelectedIndex(context),
            onDestinationSelected: (index) => _onDestinationSelected(context, index),
            labelType: NavigationRailLabelType.all,
            leading: Padding(
              padding: const EdgeInsets.all(16),
              child: Icon(
                Icons.casino_rounded,
                size: 40,
                color: Theme.of(context).colorScheme.primary,
              ),
            ),
            destinations: [
              NavigationRailDestination(
                icon: const Icon(Icons.dashboard_outlined),
                selectedIcon: const Icon(Icons.dashboard),
                label: Text(l10n.navDashboard),
              ),
              NavigationRailDestination(
                icon: const Icon(Icons.people_outline),
                selectedIcon: const Icon(Icons.people),
                label: Text(l10n.navUsers),
              ),
              NavigationRailDestination(
                icon: const Icon(Icons.meeting_room_outlined),
                selectedIcon: const Icon(Icons.meeting_room),
                label: Text(l10n.navRooms),
              ),
              NavigationRailDestination(
                icon: const Icon(Icons.account_balance_wallet_outlined),
                selectedIcon: const Icon(Icons.account_balance_wallet),
                label: Text(l10n.navFunds),
              ),
              NavigationRailDestination(
                icon: const Icon(Icons.monitor_heart_outlined),
                selectedIcon: const Icon(Icons.monitor_heart),
                label: Text(l10n.navMonitoring),
              ),
              NavigationRailDestination(
                icon: const Icon(Icons.warning_amber_outlined),
                selectedIcon: const Icon(Icons.warning_amber),
                label: Text(l10n.navRiskFlags),
              ),
              NavigationRailDestination(
                icon: const Icon(Icons.notifications_outlined),
                selectedIcon: const Icon(Icons.notifications),
                label: Text(l10n.navAlerts),
              ),
            ],
            trailing: Expanded(
              child: Align(
                alignment: Alignment.bottomCenter,
                child: Padding(
                  padding: const EdgeInsets.only(bottom: 16),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // 语言切换按钮
                      Consumer(
                        builder: (context, ref, _) {
                          final locale = ref.watch(localeProvider);
                          final currentCode = locale?.languageCode ?? Localizations.localeOf(context).languageCode;

                          return PopupMenuButton<Locale>(
                            tooltip: 'Language',
                            icon: const Icon(Icons.language),
                            onSelected: (value) {
                              ref.read(localeProvider.notifier).state = value;
                            },
                            itemBuilder: (context) => [
                              CheckedPopupMenuItem<Locale>(
                                value: const Locale('en'),
                                checked: currentCode == 'en',
                                child: const Text('English'),
                              ),
                              CheckedPopupMenuItem<Locale>(
                                value: const Locale('zh'),
                                checked: currentCode == 'zh',
                                child: const Text('中文'),
                              ),
                            ],
                          );
                        },
                      ),
                      Consumer(
                        builder: (context, ref, _) {
                          return IconButton(
                            icon: const Icon(Icons.logout),
                            tooltip: l10n.navLogout,
                            onPressed: () async {
                              await ref.read(adminAuthProvider.notifier).logout();
                              if (context.mounted) {
                                context.go('/login');
                              }
                            },
                          );
                        },
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          const VerticalDivider(width: 1),
          Expanded(child: child),
        ],
      ),
    );
  }

  int _getSelectedIndex(BuildContext context) {
    final location = GoRouterState.of(context).matchedLocation;
    if (location.startsWith('/dashboard')) return 0;
    if (location.startsWith('/users')) return 1;
    if (location.startsWith('/rooms')) return 2;
    if (location.startsWith('/funds')) return 3;
    if (location.startsWith('/monitoring')) return 4;
    if (location.startsWith('/risk-flags')) return 5;
    if (location.startsWith('/alerts')) return 6;
    return 0;
  }

  void _onDestinationSelected(BuildContext context, int index) {
    switch (index) {
      case 0:
        context.go('/dashboard');
        break;
      case 1:
        context.go('/users');
        break;
      case 2:
        context.go('/rooms');
        break;
      case 3:
        context.go('/funds');
        break;
      case 4:
        context.go('/monitoring');
        break;
      case 5:
        context.go('/risk-flags');
        break;
      case 6:
        context.go('/alerts');
        break;
    }
  }
}
