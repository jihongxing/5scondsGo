import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';
import 'package:intl/intl.dart';

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
      length: 3,
      child: Scaffold(
        appBar: AppBar(
          title: Text(l10n.fundsTitle),
          bottom: TabBar(
            tabs: [
              Tab(text: l10n.fundsTabRequests),
              Tab(text: l10n.fundsTabTransactions),
              const Tab(text: '资金对账'),
            ],
          ),
        ),
        body: TabBarView(
          children: [
            _buildFundRequestsTab(context),
            _buildTransactionsTab(context),
            _buildReconciliationTab(context),
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

  // ===== 资金对账 Tab =====
  Widget _buildReconciliationTab(BuildContext context) {
    final api = ref.read(adminApiClientProvider);

    return FutureBuilder<Map<String, dynamic>>(
      future: api.getReconciliationReport(),
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) {
          return Center(
            child: Column(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                const Icon(Icons.error_outline, size: 48, color: Colors.red),
                const SizedBox(height: 16),
                Text('加载失败: ${snapshot.error}'),
                const SizedBox(height: 16),
                ElevatedButton(
                  onPressed: () => setState(() {}),
                  child: const Text('重试'),
                ),
              ],
            ),
          );
        }

        final data = snapshot.data ?? {};
        return _buildReconciliationReport(context, data);
      },
    );
  }

  Widget _buildReconciliationReport(
      BuildContext context, Map<String, dynamic> data) {
    final external = data['external_funds'] as Map<String, dynamic>? ?? {};
    final system = data['system_funds'] as Map<String, dynamic>? ?? {};
    final reconciliation =
        data['reconciliation'] as Map<String, dynamic>? ?? {};
    final analysis = data['analysis'] as Map<String, dynamic>? ?? {};

    final isBalanced = reconciliation['is_balanced'] as bool? ?? false;
    final difference = _parseDecimal(reconciliation['difference']);

    return RefreshIndicator(
      onRefresh: () async => setState(() {}),
      child: SingleChildScrollView(
        physics: const AlwaysScrollableScrollPhysics(),
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 对账状态卡片
            _buildStatusCard(isBalanced, difference),
            const SizedBox(height: 16),

            // 左右对照表
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 左边：外部资金注入
                Expanded(child: _buildExternalFundsCard(external)),
                const SizedBox(width: 16),
                // 右边：系统内资金分布
                Expanded(child: _buildSystemFundsCard(system)),
              ],
            ),
            const SizedBox(height: 16),

            // 差异分析
            if (!isBalanced) _buildAnalysisCard(analysis),

            const SizedBox(height: 24),
            // 对账历史按钮
            Center(
              child: OutlinedButton.icon(
                onPressed: () => _showReconciliationHistory(context),
                icon: const Icon(Icons.history),
                label: const Text('查看对账历史'),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildStatusCard(bool isBalanced, double difference) {
    return Card(
      color: isBalanced ? Colors.green[50] : Colors.red[50],
      child: Padding(
        padding: const EdgeInsets.all(20),
        child: Row(
          children: [
            Icon(
              isBalanced ? Icons.check_circle : Icons.warning,
              size: 48,
              color: isBalanced ? Colors.green : Colors.red,
            ),
            const SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    isBalanced ? '资金守恒 ✓' : '资金不平衡 ⚠',
                    style: TextStyle(
                      fontSize: 20,
                      fontWeight: FontWeight.bold,
                      color: isBalanced ? Colors.green[800] : Colors.red[800],
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    isBalanced
                        ? '系统内资金与外部注入资金一致'
                        : '差异: ¥${_formatNumber(difference)}',
                    style: TextStyle(
                      color: isBalanced ? Colors.green[700] : Colors.red[700],
                    ),
                  ),
                ],
              ),
            ),
            IconButton(
              icon: const Icon(Icons.refresh),
              onPressed: () => setState(() {}),
              tooltip: '刷新',
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildExternalFundsCard(Map<String, dynamic> external) {
    final ownerDeposit = _parseDecimal(external['owner_deposit']);
    final marginDeposit = _parseDecimal(external['margin_deposit']);
    final ownerWithdraw = _parseDecimal(external['owner_withdraw']);
    final netInflow = _parseDecimal(external['net_inflow']);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.input, color: Colors.blue),
                SizedBox(width: 8),
                Text(
                  '外部资金注入（左边）',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(),
            _buildDetailRow('房主充值 (owner_deposit)', ownerDeposit, Colors.green),
            _buildDetailRow('保证金充值 (margin_deposit)', marginDeposit, Colors.green),
            _buildDetailRow('房主提现 (owner_withdraw)', -ownerWithdraw, Colors.red),
            const Divider(),
            _buildDetailRow('净流入合计', netInflow, Colors.blue, isBold: true),
          ],
        ),
      ),
    );
  }

  Widget _buildSystemFundsCard(Map<String, dynamic> system) {
    final playerBalance = _parseDecimal(system['player_balance']);
    final playerFrozen = _parseDecimal(system['player_frozen']);
    final ownerBalance = _parseDecimal(system['owner_balance']);
    final ownerCommission = _parseDecimal(system['owner_commission']);
    final ownerMargin = _parseDecimal(system['owner_margin']);
    final platformBalance = _parseDecimal(system['platform_balance']);
    final total = _parseDecimal(system['total']);

    return Card(
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.account_balance_wallet, color: Colors.orange),
                SizedBox(width: 8),
                Text(
                  '系统内资金分布（右边）',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(),
            _buildDetailRow('玩家可用余额', playerBalance, Colors.blue),
            _buildDetailRow('玩家冻结余额', playerFrozen, Colors.grey),
            _buildDetailRow('房主可用余额', ownerBalance, Colors.purple),
            _buildDetailRow('房主佣金收益', ownerCommission, Colors.teal),
            _buildDetailRow('房主保证金', ownerMargin, Colors.indigo),
            _buildDetailRow('平台余额', platformBalance, Colors.amber),
            const Divider(),
            _buildDetailRow('系统总计', total, Colors.blue, isBold: true),
          ],
        ),
      ),
    );
  }

  Widget _buildDetailRow(String label, double value, Color color,
      {bool isBold = false}) {
    final isNegative = value < 0;

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Expanded(
            child: Text(
              label,
              style: TextStyle(
                fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          ),
          Text(
            '¥${_formatNumber(value)}',
            style: TextStyle(
              color: isNegative ? Colors.red : color,
              fontWeight: isBold ? FontWeight.bold : FontWeight.normal,
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAnalysisCard(Map<String, dynamic> analysis) {
    final unrecordedMargin = _parseDecimal(analysis['unrecorded_margin']);
    final explanation = analysis['explanation'] as String? ?? '';

    return Card(
      color: Colors.amber[50],
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Row(
              children: [
                Icon(Icons.analytics, color: Colors.amber),
                SizedBox(width: 8),
                Text(
                  '差异分析',
                  style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(),
            if (unrecordedMargin > 0)
              _buildDetailRow('未记录的保证金', unrecordedMargin, Colors.orange),
            const SizedBox(height: 8),
            Text(explanation, style: const TextStyle(color: Colors.black87)),
          ],
        ),
      ),
    );
  }

  void _showReconciliationHistory(BuildContext context) {
    final api = ref.read(adminApiClientProvider);

    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('对账历史'),
        content: SizedBox(
          width: 600,
          height: 400,
          child: FutureBuilder<Map<String, dynamic>>(
            future: api.getBalanceCheckHistory(pageSize: 20),
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const Center(child: CircularProgressIndicator());
              }
              if (snapshot.hasError) {
                return Center(child: Text('加载失败: ${snapshot.error}'));
              }

              final data = snapshot.data ?? {};
              final items = (data['items'] as List<dynamic>? ?? [])
                  .cast<Map<String, dynamic>>();

              if (items.isEmpty) {
                return const Center(child: Text('暂无对账历史'));
              }

              return ListView.builder(
                itemCount: items.length,
                itemBuilder: (context, index) {
                  final item = items[index];
                  final scope = item['scope'] as String? ?? 'global';
                  final periodType = item['period_type'] as String? ?? '2h';
                  final isBalanced = item['is_balanced'] as bool? ?? false;
                  final difference = _parseDecimal(item['difference']);
                  final createdAt = item['created_at'] as String? ?? '';

                  final dateFormat = DateFormat('yyyy-MM-dd HH:mm');
                  DateTime? date;
                  try {
                    date = DateTime.parse(createdAt);
                  } catch (_) {}

                  final scopeLabel = scope == 'global' ? '全局' : '房主';
                  final periodLabel = periodType == '2h' ? '2小时' : '每日';

                  return ListTile(
                    leading: CircleAvatar(
                      backgroundColor:
                          isBalanced ? Colors.green[100] : Colors.red[100],
                      child: Icon(
                        isBalanced ? Icons.check : Icons.warning,
                        color: isBalanced ? Colors.green : Colors.red,
                        size: 20,
                      ),
                    ),
                    title: Text('$scopeLabel - $periodLabel对账'),
                    subtitle: Text(
                      date != null ? dateFormat.format(date) : createdAt,
                    ),
                    trailing: Text(
                      isBalanced ? '平衡' : '差异: ¥${_formatNumber(difference)}',
                      style: TextStyle(
                        color: isBalanced ? Colors.green : Colors.red,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  );
                },
              );
            },
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: const Text('关闭'),
          ),
        ],
      ),
    );
  }

  double _parseDecimal(dynamic value) {
    if (value == null) return 0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0;
    return 0;
  }

  String _formatNumber(double value) {
    final formatter = NumberFormat('#,##0.00');
    return formatter.format(value);
  }
}
