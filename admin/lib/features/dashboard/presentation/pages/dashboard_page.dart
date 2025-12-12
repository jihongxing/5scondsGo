import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';

import '../../../../core/services/api_client.dart';
import '../../../auth/providers/auth_provider.dart';

class DashboardPage extends ConsumerWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    final l10n = AppLocalizations.of(context)!;
    final api = ref.read(adminApiClientProvider);
    final auth = ref.watch(adminAuthProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.dashboardTitle),
      ),
      body: Padding(
        padding: const EdgeInsets.all(24),
        child: FutureBuilder<Map<String, dynamic>>(
          future: api.getBalanceCheckReport(),
          builder: (context, snapshot) {
            if (snapshot.connectionState == ConnectionState.waiting) {
              return const Center(child: CircularProgressIndicator());
            }
            if (snapshot.hasError) {
              return Center(child: Text('Error: ${snapshot.error}'));
            }
            final data = snapshot.data ?? {};
            final check = (data['check'] ?? {}) as Map<String, dynamic>;
            final summary = (data['summary'] ?? {}) as Map<String, dynamic>;

            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Admin info
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Row(
                      children: [
                        Icon(
                          Icons.admin_panel_settings,
                          color: Theme.of(context).colorScheme.primary,
                        ),
                        const SizedBox(width: 12),
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              l10n.dashboardAdminInviteTitle,
                              style: Theme.of(context).textTheme.titleMedium,
                            ),
                            const SizedBox(height: 4),
                            Text(
                              // 这里用当前登录账号名做展示
                              l10n.dashboardAdminInviteCodeLabel(
                                auth.username ?? 'ADMIN',
                              ),
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ],
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 16),

                // Stats cards（改为展示资金汇总）
                Row(
                  children: [
                    Expanded(
                      child: _buildStatCard(
                        context,
                        l10n.dashboardCardTotalUsers,
                        _formatAmount(summary['total_deposit']),
                        Icons.account_balance,
                        Colors.blue,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: _buildStatCard(
                        context,
                        l10n.dashboardCardActiveRooms,
                        _formatAmount(summary['total_withdraw']),
                        Icons.money_off,
                        Colors.green,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: _buildStatCard(
                        context,
                        l10n.dashboardCardOnlinePlayers,
                        _formatAmount(summary['total_bet']),
                        Icons.casino,
                        Colors.orange,
                      ),
                    ),
                    const SizedBox(width: 16),
                    Expanded(
                      child: _buildStatCard(
                        context,
                        l10n.dashboardCardPlatformBalance,
                        _formatAmount(summary['platform_balance']),
                        Icons.account_balance_wallet,
                        Colors.purple,
                      ),
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                // Fund conservation check
                Card(
                  child: Padding(
                    padding: const EdgeInsets.all(16),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(Icons.verified, color: Colors.green[700]),
                            const SizedBox(width: 8),
                            Text(
                              l10n.dashboardFundCheckTitle,
                              style: Theme.of(context).textTheme.titleMedium,
                            ),
                          ],
                        ),
                        const SizedBox(height: 16),
                        Text(
                          l10n.dashboardFundCheckPlayerBalance(
                            _formatAmount(check['total_player_balance']),
                          ),
                        ),
                        Text(
                          l10n.dashboardFundCheckCustodyQuota(
                            _formatAmount(check['total_custody_quota']),
                          ),
                        ),
                        Text(
                          l10n.dashboardFundCheckMargin(
                            _formatAmount(check['total_margin']),
                          ),
                        ),
                        Text(
                          l10n.dashboardFundCheckPlatformProfit(
                            _formatAmount(check['platform_balance']),
                          ),
                        ),
                        const SizedBox(height: 8),
                        Container(
                          padding: const EdgeInsets.symmetric(
                            horizontal: 12,
                            vertical: 6,
                          ),
                          decoration: BoxDecoration(
                            color: (check['is_balanced'] == true)
                                ? Colors.green[100]
                                : Colors.red[100],
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Text(
                            check['is_balanced'] == true
                                ? l10n.dashboardFundCheckBalanced
                                : 'Unbalanced',
                            style: TextStyle(
                              color: (check['is_balanced'] == true)
                                  ? Colors.green[800]
                                  : Colors.red[800],
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: 24),

                // Recent activity
                Expanded(
                  child: _RecentFundRequestsList(api: api),
                ),
              ],
            );
          },
        ),
      ),
    );
  }

  Widget _buildStatCard(
    BuildContext context,
    String title,
    String value,
    IconData icon,
    Color color,
  ) {
    return Card(
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Container(
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: color.withAlpha(25),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Icon(icon, color: color),
                ),
                const Spacer(),
              ],
            ),
            const SizedBox(height: 16),
            Text(
              value,
              style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                    fontWeight: FontWeight.bold,
                  ),
            ),
            const SizedBox(height: 4),
            Text(
              title,
              style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                    color: Colors.grey[600],
                  ),
            ),
          ],
        ),
      ),
    );
  }

  static String _formatAmount(dynamic value) {
    if (value == null) return '-';
    return '¥$value';
  }
}

class _RecentFundRequestsList extends StatelessWidget {
  final AdminApiClient api;

  const _RecentFundRequestsList({required this.api});

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              l10n.dashboardRecentFundRequests,
              style: Theme.of(context).textTheme.titleMedium,
            ),
            const SizedBox(height: 16),
            Expanded(
              child: FutureBuilder<Map<String, dynamic>>(
                future: api.listFundRequests(page: 1, pageSize: 10),
                builder: (context, snapshot) {
                  if (snapshot.connectionState == ConnectionState.waiting) {
                    return const Center(child: CircularProgressIndicator());
                  }
                  if (snapshot.hasError) {
                    return Center(child: Text('Error: ${snapshot.error}'));
                  }
                  final data = snapshot.data ?? {};
                  final items = (data['items'] as List<dynamic>? ?? [])
                      .cast<Map<String, dynamic>>();
                  if (items.isEmpty) {
                    return const Center(child: Text('No data'));
                  }

                  return ListView.builder(
                    itemCount: items.length,
                    itemBuilder: (context, index) {
                      final item = items[index];
                      final id = item['id']?.toString() ?? '';
                      final type = (item['type'] as String? ?? '');
                      final status = (item['status'] as String? ?? '');
                      final amount = item['amount'] ?? '';

                      final isDeposit = type == 'deposit' || type == 'owner_deposit';
                      Color statusColor;
                      IconData statusIcon;
                      if (status == 'pending') {
                        statusColor = Colors.orange;
                        statusIcon = Icons.hourglass_empty;
                      } else if (status == 'approved') {
                        statusColor = Colors.green;
                        statusIcon = Icons.check;
                      } else {
                        statusColor = Colors.red;
                        statusIcon = Icons.close;
                      }

                      return ListTile(
                        leading: CircleAvatar(
                          backgroundColor: statusColor.withOpacity(0.1),
                          child: Icon(statusIcon, color: statusColor),
                        ),
                        title: Text(
                          l10n.dashboardRecentFundTitle(
                            isDeposit
                                ? l10n.dashboardRecentFundTypeDeposit
                                : l10n.dashboardRecentFundTypeWithdraw,
                            item['username']?.toString().isNotEmpty == true
                                ? item['username'].toString()
                                : 'User #$id',
                          ),
                        ),
                        subtitle: Text('¥$amount'),
                        trailing: Text(
                          status == 'pending'
                              ? l10n.dashboardRecentFundStatusPending
                              : status == 'approved'
                                  ? l10n.dashboardRecentFundStatusApproved
                                  : l10n.dashboardRecentFundStatusRejected,
                          style: TextStyle(
                            color: statusColor,
                          ),
                        ),
                      );
                    },
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }
}
