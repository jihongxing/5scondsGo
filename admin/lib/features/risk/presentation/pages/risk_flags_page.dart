import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/services/api_client.dart';

class RiskFlagsPage extends ConsumerStatefulWidget {
  const RiskFlagsPage({super.key});

  @override
  ConsumerState<RiskFlagsPage> createState() => _RiskFlagsPageState();
}

class _RiskFlagsPageState extends ConsumerState<RiskFlagsPage> {
  String? _selectedStatus;
  String? _selectedType;
  int _page = 1;
  final int _pageSize = 20;

  @override
  Widget build(BuildContext context) {
    final api = ref.read(adminApiClientProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('风控标记'),
      ),
      body: Column(
        children: [
          // 筛选栏
          Padding(
            padding: const EdgeInsets.all(16),
            child: Row(
              children: [
                // 状态筛选
                DropdownButton<String?>(
                  value: _selectedStatus,
                  hint: const Text('状态'),
                  items: const [
                    DropdownMenuItem(value: null, child: Text('全部')),
                    DropdownMenuItem(value: 'pending', child: Text('待审核')),
                    DropdownMenuItem(value: 'confirmed', child: Text('已确认')),
                    DropdownMenuItem(value: 'dismissed', child: Text('已驳回')),
                  ],
                  onChanged: (value) {
                    setState(() {
                      _selectedStatus = value;
                      _page = 1;
                    });
                  },
                ),
                const SizedBox(width: 16),
                // 类型筛选
                DropdownButton<String?>(
                  value: _selectedType,
                  hint: const Text('类型'),
                  items: const [
                    DropdownMenuItem(value: null, child: Text('全部')),
                    DropdownMenuItem(value: 'consecutive_wins', child: Text('连续获胜')),
                    DropdownMenuItem(value: 'high_win_rate', child: Text('高胜率')),
                    DropdownMenuItem(value: 'multi_account', child: Text('多账户')),
                    DropdownMenuItem(value: 'large_transaction', child: Text('大额交易')),
                  ],
                  onChanged: (value) {
                    setState(() {
                      _selectedType = value;
                      _page = 1;
                    });
                  },
                ),
              ],
            ),
          ),
          // 列表
          Expanded(
            child: FutureBuilder<Map<String, dynamic>>(
              future: api.listRiskFlags(
                page: _page,
                pageSize: _pageSize,
                status: _selectedStatus,
                flagType: _selectedType,
              ),
              builder: (context, snapshot) {
                if (snapshot.connectionState == ConnectionState.waiting) {
                  return const Center(child: CircularProgressIndicator());
                }
                if (snapshot.hasError) {
                  return Center(child: Text('Error: ${snapshot.error}'));
                }

                final data = snapshot.data ?? {};
                final flags = (data['flags'] as List<dynamic>? ?? [])
                    .cast<Map<String, dynamic>>();
                final total = data['total'] as int? ?? 0;

                if (flags.isEmpty) {
                  return const Center(child: Text('暂无风控标记'));
                }

                return Column(
                  children: [
                    Expanded(
                      child: ListView.builder(
                        itemCount: flags.length,
                        itemBuilder: (context, index) {
                          final flag = flags[index];
                          return _buildFlagCard(context, flag, api);
                        },
                      ),
                    ),
                    // 分页
                    Padding(
                      padding: const EdgeInsets.all(16),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          IconButton(
                            icon: const Icon(Icons.chevron_left),
                            onPressed: _page > 1
                                ? () => setState(() => _page--)
                                : null,
                          ),
                          Text('第 $_page 页'),
                          IconButton(
                            icon: const Icon(Icons.chevron_right),
                            onPressed: _page * _pageSize < total
                                ? () => setState(() => _page++)
                                : null,
                          ),
                        ],
                      ),
                    ),
                  ],
                );
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildFlagCard(BuildContext context, Map<String, dynamic> flag, AdminApiClient api) {
    final id = flag['id'] as int? ?? 0;
    final userId = flag['user_id'] as int? ?? 0;
    final flagType = flag['flag_type'] as String? ?? '';
    final status = flag['status'] as String? ?? '';
    final createdAt = flag['created_at'] as String? ?? '';

    Color statusColor;
    String statusText;
    switch (status) {
      case 'pending':
        statusColor = Colors.orange;
        statusText = '待审核';
        break;
      case 'confirmed':
        statusColor = Colors.red;
        statusText = '已确认';
        break;
      case 'dismissed':
        statusColor = Colors.grey;
        statusText = '已驳回';
        break;
      default:
        statusColor = Colors.grey;
        statusText = status;
    }

    String typeText;
    IconData typeIcon;
    switch (flagType) {
      case 'consecutive_wins':
        typeText = '连续获胜';
        typeIcon = Icons.emoji_events;
        break;
      case 'high_win_rate':
        typeText = '高胜率';
        typeIcon = Icons.trending_up;
        break;
      case 'multi_account':
        typeText = '多账户';
        typeIcon = Icons.people;
        break;
      case 'large_transaction':
        typeText = '大额交易';
        typeIcon = Icons.attach_money;
        break;
      default:
        typeText = flagType;
        typeIcon = Icons.flag;
    }

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: statusColor.withValues(alpha: 0.1),
          child: Icon(typeIcon, color: statusColor),
        ),
        title: Text('用户 #$userId - $typeText'),
        subtitle: Text('创建时间: $createdAt'),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: statusColor.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(statusText, style: TextStyle(color: statusColor)),
            ),
            if (status == 'pending') ...[
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.check, color: Colors.green),
                onPressed: () => _reviewFlag(api, id, 'confirm'),
                tooltip: '确认',
              ),
              IconButton(
                icon: const Icon(Icons.close, color: Colors.red),
                onPressed: () => _reviewFlag(api, id, 'dismiss'),
                tooltip: '驳回',
              ),
            ],
          ],
        ),
        onTap: () => _showFlagDetails(context, flag),
      ),
    );
  }

  Future<void> _reviewFlag(AdminApiClient api, int id, String action) async {
    try {
      await api.reviewRiskFlag(id, action);
      setState(() {}); // 刷新列表
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text(action == 'confirm' ? '已确认' : '已驳回')),
        );
      }
    } catch (e) {
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('操作失败: $e')),
        );
      }
    }
  }

  void _showFlagDetails(BuildContext context, Map<String, dynamic> flag) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('风控标记详情'),
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('ID: ${flag['id']}'),
              Text('用户ID: ${flag['user_id']}'),
              Text('类型: ${flag['flag_type']}'),
              Text('状态: ${flag['status']}'),
              Text('创建时间: ${flag['created_at']}'),
              const SizedBox(height: 8),
              const Text('详情:', style: TextStyle(fontWeight: FontWeight.bold)),
              Text(flag['details'] ?? '无'),
            ],
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
}
