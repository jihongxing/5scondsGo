import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:go_router/go_router.dart';

import '../../features/auth/presentation/pages/login_page.dart';
import '../../features/auth/presentation/pages/register_page.dart';
import '../../features/home/presentation/pages/home_page.dart';
import '../../features/room/presentation/pages/create_room_page.dart';
import '../../features/room/presentation/pages/room_page.dart';
import '../../features/wallet/presentation/pages/wallet_page.dart';
import '../../features/game_history/presentation/pages/game_history_page.dart';
import '../../features/game_history/presentation/pages/game_replay_page.dart';
import '../../features/friends/presentation/pages/friend_list_page.dart';
import '../../features/friends/presentation/pages/friend_requests_page.dart';
import '../../features/profile/presentation/pages/profile_page.dart';
import '../../features/invite/presentation/pages/invite_link_page.dart';
import '../../features/auth/providers/auth_provider.dart';

final routerProvider = Provider<GoRouter>((ref) {
  final authState = ref.watch(authProvider);

  return GoRouter(
    initialLocation: '/login',
    redirect: (context, state) {
      final isLoggedIn = authState.isLoggedIn;
      final isAuthRoute = state.matchedLocation == '/login' || state.matchedLocation == '/register';

      if (!isLoggedIn && !isAuthRoute) {
        return '/login';
      }
      if (isLoggedIn && isAuthRoute) {
        return '/home';
      }
      return null;
    },
    routes: [
      GoRoute(
        path: '/login',
        builder: (context, state) => const LoginPage(),
      ),
      GoRoute(
        path: '/register',
        builder: (context, state) => const RegisterPage(),
      ),
      GoRoute(
        path: '/home',
        builder: (context, state) => const HomePage(),
      ),
      GoRoute(
        path: '/create-room',
        builder: (context, state) => const CreateRoomPage(),
      ),
      GoRoute(
        path: '/wallet',
        builder: (context, state) => const WalletPage(),
      ),
      GoRoute(
        path: '/room/:id',
        builder: (context, state) {
          final roomId = int.parse(state.pathParameters['id']!);
          return RoomPage(roomId: roomId);
        },
      ),
      GoRoute(
        path: '/game-history',
        builder: (context, state) => const GameHistoryPage(),
      ),
      GoRoute(
        path: '/game-history/:id',
        builder: (context, state) {
          final roundId = int.parse(state.pathParameters['id']!);
          return GameReplayPage(roundId: roundId);
        },
      ),
      GoRoute(
        path: '/friends',
        builder: (context, state) => const FriendListPage(),
      ),
      GoRoute(
        path: '/friend-requests',
        builder: (context, state) => const FriendRequestsPage(),
      ),
      GoRoute(
        path: '/profile',
        builder: (context, state) => const ProfilePage(),
      ),
      GoRoute(
        path: '/invite/:code',
        builder: (context, state) {
          final code = state.pathParameters['code']!;
          return InviteLinkPage(code: code);
        },
      ),
    ],
    errorBuilder: (context, state) => ErrorPage(location: state.matchedLocation),
  );
});

/// 错误页面 - 处理无效路由
class ErrorPage extends StatelessWidget {
  final String location;

  const ErrorPage({super.key, required this.location});

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      body: Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.error_outline, size: 64, color: Colors.red[300]),
            const SizedBox(height: 16),
            Text(
              '页面未找到',
              style: Theme.of(context).textTheme.headlineSmall,
            ),
            const SizedBox(height: 8),
            Text(
              location,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey,
                  ),
            ),
            const SizedBox(height: 24),
            ElevatedButton.icon(
              onPressed: () => context.go('/home'),
              icon: const Icon(Icons.home),
              label: const Text('返回首页'),
            ),
          ],
        ),
      ),
    );
  }
}
