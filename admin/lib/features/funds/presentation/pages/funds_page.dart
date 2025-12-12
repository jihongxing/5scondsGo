import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';

import '../../../../core/services/api_client.dart';

class FundsPage extends ConsumerStatefulWidget {
  const FundsPage({super.key});

  @override
  ConsumerState<FundsPage> createState() => _FundsPageState();
}

class _FundsPageState extends ConsumerState<FundsPage> {
  String _requestStatusFilter = 'pending';
  String _requestTypeFilter = 'all';

  final _txUserController = TextEditingController();
  String _txTypeFilter = 'all';

  @override
  void dispose() {
    _txUserController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return DefaultTabController(
      length: 2,
      child: Scaffold(
        appBar: AppBar(
          title: Text(l10n.fundsTitle),
          bottom: TabBar(
            tabs: [
              Tab(text: l10n.fundsTabRequests),
              Tab(text: l10n.fundsTabTransactions),
            ],
          ),
        ),
        body: TabBarView(
          children: [
            _buildFundRequestsTab(context),
            _buildTransactionsTab(context),
          ],
        ),
      ),
    );
  }

  Widget _buildFundRequestsTab(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final api = ref.read(adminApiClientProvider);

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Card(
        child: Column(
          children: [
            // Filters
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  DropdownButton<String>(
                    value: _requestStatusFilter,
                    items: [
                      DropdownMenuItem(
                        value: 'all',
                        child: Text(l10n.fundsFilterAllStatus),
                      ),
                      DropdownMenuItem(
                        value: 'pending',
                        child: Text(l10n.fundsFilterStatusPending),
                      ),
                      DropdownMenuItem(
                        value: 'approved',
                        child: Text(l10n.fundsFilterStatusApproved),
                      ),
                      DropdownMenuItem(
                        value: 'rejected',
                        child: Text(l10n.fundsFilterStatusRejected),
                      ),
                    ],
                    onChanged: (value) {
                      if (value == null) return;
                      setState(() {
                        _requestStatusFilter = value;
                      });
                    },
                  ),
                  const SizedBox(width: 16),
                  DropdownButton<String>(
                    value: _requestTypeFilter,
                    items: [
                      DropdownMenuItem(
                        value: 'all',
                        child: Text(l10n.fundsFilterAllTypes),
                      ),
                      DropdownMenuItem(
                        value: 'deposit',
                        child: Text(l10n.fundsFilterTypeDeposit),
                      ),
                      DropdownMenuItem(
                        value: 'withdraw',
                        child: Text(l10n.fundsFilterTypeWithdraw),
                      ),
                      DropdownMenuItem(
                        value: 'margin',
                        child: Text(l10n.fundsFilterTypeMargin),
                      ),
                    ],
                    onChanged: (value) {
                      if (value == null) return;
                      setState(() {
                        _requestTypeFilter = value;
                      });
                    },
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            // List
            Expanded(
              child: FutureBuilder<Map<String, dynamic>>(
                future: api.listFundRequests(
                  status: _requestStatusFilter == 'all'
                      ? null
                      : _requestStatusFilter,
                  // type: 仅在 deposit / withdraw 时传后端，其它前端过滤
                  type: _requestTypeFilter == 'deposit' ||
                          _requestTypeFilter == 'withdraw'
                      ? _requestTypeFilter
                      : null,
                ),
                builder: (context, snapshot) {
                  if (snapshot.connectionState == ConnectionState.waiting) {
                    return const Center(child: CircularProgressIndicator());
                  }
                  if (snapshot.hasError) {
                    return Center(child: Text('Error: ${snapshot.error}'));
                  }
                  final data = snapshot.data ?? {};
                  var items = (data['items'] as List<dynamic>? ?? [])
                      .cast<Map<String, dynamic>>();

                  // 本地按 type=margin 过滤
                  if (_requestTypeFilter == 'margin') {
                    items = items
                        .where((e) => (e['type'] as String?)
                                ?.startsWith('margin_') ??
                            false)
                        .toList();
                  }

                  if (items.isEmpty) {
                    return const Center(child: Text('No data'));
                  }

                  return ListView.builder(
                    itemCount: items.length,
                    itemBuilder: (context, index) {
                      final item = items[index];
                      final id = item['id'] as int? ?? 0;
                      final status = item['status'] as String? ?? 'pending';
                      final type = item['type'] as String? ?? 'deposit';
                      final amount = item['amount'] ?? '0';
                      final userId = item['user_id']?.toString() ?? '-';

                      final isDepositType = type == 'deposit' ||
                          type == 'owner_deposit' ||
                          type == 'margin_deposit';

                      final icon = isDepositType
                          ? Icons.arrow_downward
                          : Icons.arrow_upward;
                      final iconColor = isDepositType ? Colors.green : Colors.red;

                      return ListTile(
                        leading: CircleAvatar(
                          backgroundColor: iconColor.withOpacity(0.1),
                          child: Icon(icon, color: iconColor),
                        ),
                        title: Text('User $userId'),
                        subtitle: Text(
                          '${type.toUpperCase()} • ¥$amount',
                        ),
                        trailing: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            if (status == 'pending') ...[
                              IconButton(
                                icon: const Icon(Icons.check, color: Colors.green),
                                tooltip: l10n.fundsStatusApproved,
                                onPressed: () async {
                                  await _processFundRequest(
                                    context,
                                    id: id,
                                    approved: true,
                                  );
                                },
                              ),
                              IconButton(
                                icon: const Icon(Icons.close, color: Colors.red),
                                tooltip: l10n.fundsStatusRejected,
                                onPressed: () async {
                                  await _processFundRequest(
                                    context,
                                    id: id,
                                    approved: false,
                                  );
                                },
                              ),
                            ] else
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: status == 'approved'
                                      ? Colors.green[100]
                                      : Colors.red[100],
                                  borderRadius: BorderRadius.circular(16),
                                ),
                                child: Text(
                                  status == 'pending'
                                      ? l10n.fundsStatusPending
                                      : status == 'approved'
                                          ? l10n.fundsStatusApproved
                                          : l10n.fundsStatusRejected,
                                  style: TextStyle(
                                    color: status == 'approved'
                                        ? Colors.green[800]
                                        : Colors.red[800],
                                    fontWeight: FontWeight.bold,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                          ],
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

  Widget _buildTransactionsTab(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final api = ref.read(adminApiClientProvider);

    return Padding(
      padding: const EdgeInsets.all(24),
      child: Card(
        child: Column(
          children: [
            // Filters
            Padding(
              padding: const EdgeInsets.all(16),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      controller: _txUserController,
                      decoration: InputDecoration(
                        hintText: l10n.fundsSearchUserHint,
                        prefixIcon: const Icon(Icons.search),
                        border: const OutlineInputBorder(),
                      ),
                      onSubmitted: (_) => setState(() {}),
                    ),
                  ),
                  const SizedBox(width: 16),
                  DropdownButton<String>(
                    value: _txTypeFilter,
                    items: [
                      DropdownMenuItem(
                        value: 'all',
                        child: Text(l10n.fundsFilterAllTypes),
                      ),
                      DropdownMenuItem(
                        value: 'game_bet',
                        child: Text(l10n.fundsTxTypeGameBet),
                      ),
                      DropdownMenuItem(
                        value: 'game_win',
                        child: Text(l10n.fundsTxTypeGameWin),
                      ),
                      DropdownMenuItem(
                        value: 'deposit',
                        child: Text(l10n.fundsTxTypeDeposit),
                      ),
                      DropdownMenuItem(
                        value: 'withdraw',
                        child: Text(l10n.fundsTxTypeWithdraw),
                      ),
                    ],
                    onChanged: (value) {
                      if (value == null) return;
                      setState(() {
                        _txTypeFilter = value;
                      });
                    },
                  ),
                ],
              ),
            ),
            const Divider(height: 1),
            // List
            Expanded(
              child: FutureBuilder<Map<String, dynamic>>(
                future: api.listTransactions(
                  type: _txTypeFilter == 'all' ? null : _txTypeFilter,
                  userId: int.tryParse(_txUserController.text.trim()),
                ),
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
                      final tx = items[index];
                      final type = tx['type'] as String? ?? 'deposit';
                      final userId = tx['user_id']?.toString() ?? '-';
                      final amountRaw = tx['amount']?.toString() ?? '0';
                      final isPositive = !amountRaw.startsWith('-');
                      final amountDisplay = isPositive
                          ? amountRaw
                          : amountRaw.substring(1);

                      String typeLabel;
                      switch (type) {
                        case 'game_bet':
                          typeLabel = l10n.fundsTxTypeGameBet;
                          break;
                        case 'game_win':
                          typeLabel = l10n.fundsTxTypeGameWin;
                          break;
                        case 'deposit':
                          typeLabel = l10n.fundsTxTypeDeposit;
                          break;
                        case 'withdraw':
                          typeLabel = l10n.fundsTxTypeWithdraw;
                          break;
                        default:
                          typeLabel = type;
                          break;
                      }

                      return ListTile(
                        leading: CircleAvatar(
                          backgroundColor: isPositive
                              ? Colors.green[100]
                              : Colors.red[100],
                          child: Icon(
                            isPositive ? Icons.add : Icons.remove,
                            color: isPositive ? Colors.green : Colors.red,
                          ),
                        ),
                        title: Text('User $userId'),
                        subtitle: Text(typeLabel),
                        trailing: Text(
                          '${isPositive ? '+' : '-'}¥$amountDisplay',
                          style: TextStyle(
                            color: isPositive ? Colors.green : Colors.red,
                            fontWeight: FontWeight.bold,
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

  Future<void> _processFundRequest(
    BuildContext context, {
    required int id,
    required bool approved,
  }) async {
    final l10n = AppLocalizations.of(context)!;
    final api = ref.read(adminApiClientProvider);

    try {
      await api.processFundRequest(id: id, approved: approved);
      if (mounted) {
        setState(() {}); // 重新拉取列表
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(
            content: Text(
              approved
                  ? l10n.fundsStatusApproved
                  : l10n.fundsStatusRejected,
            ),
          ),
        );
      }
    } catch (e) {
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text(e.toString())),
      );
    }
  }
}
