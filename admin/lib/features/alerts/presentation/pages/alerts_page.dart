import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

import '../../../../core/services/api_client.dart';

class AlertsPage extends ConsumerStatefulWidget {
  const AlertsPage({super.key});

  @override
  ConsumerState<AlertsPage> createState() => _AlertsPageState();
}

class _AlertsPageState extends ConsumerState<AlertsPage> {
  String? _selectedStatus;
  String? _selectedSeverity;
  int _page = 1;
  final int _pageSize = 20;

  @override
  Widget build(BuildContext context) {
    final api = ref.read(adminApiClientProvider);

    return Scaffold(
      appBar: AppBar(
        title: const Text('系统告警'),
        actions: [
          // 告警摘要
          FutureBuilder<Map<String, dynamic>>(
            future: api.getAlertSummary(),
            builder: (context, snapshot) {
              if (!snapshot.hasData) return const SizedBox();
              final data = snapshot.data!;
              final activeCount = data['active_count'] as int? ?? 0;
              return Padding(
                padding: const EdgeInsets.symmetric(horizontal: 16),
                child: Center(
                  child: Container(
                    padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
                    decoration: BoxDecoration(
                      color: activeCount > 0 ? Colors.red : Colors.green,
                      borderRadius: BorderRadius.circular(16),
                    ),
                    child: Text(
                      '活跃告警: $activeCount',
                      style: const TextStyle(color: Colors.white),
                    ),
                  ),
                ),
              );
            },
          ),
        ],
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
                    DropdownMenuItem(value: 'active', child: Text('活跃')),
                    DropdownMenuItem(value: 'acknowledged', child: Text('已确认')),
                  ],
                  onChanged: (value) {
                    setState(() {
                      _selectedStatus = value;
                      _page = 1;
                    });
                  },
                ),
                const SizedBox(width: 16),
                // 严重程度筛选
                DropdownButton<String?>(
                  value: _selectedSeverity,
                  hint: const Text('严重程度'),
                  items: const [
                    DropdownMenuItem(value: null, child: Text('全部')),
                    DropdownMenuItem(value: 'critical', child: Text('严重')),
                    DropdownMenuItem(value: 'warning', child: Text('警告')),
                    DropdownMenuItem(value: 'info', child: Text('信息')),
                  ],
                  onChanged: (value) {
                    setState(() {
                      _selectedSeverity = value;
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
              future: api.listAlerts(
                page: _page,
                pageSize: _pageSize,
                status: _selectedStatus,
                severity: _selectedSeverity,
              ),
              builder: (context, snapshot) {
                if (snapshot.connectionState == ConnectionState.waiting) {
                  return const Center(child: CircularProgressIndicator());
                }
                if (snapshot.hasError) {
                  return Center(child: Text('Error: ${snapshot.error}'));
                }

                final data = snapshot.data ?? {};
                final alerts = (data['alerts'] as List<dynamic>? ?? [])
                    .cast<Map<String, dynamic>>();
                final total = data['total'] as int? ?? 0;

                if (alerts.isEmpty) {
                  return const Center(child: Text('暂无告警'));
                }

                return Column(
                  children: [
                    Expanded(
                      child: ListView.builder(
                        itemCount: alerts.length,
                        itemBuilder: (context, index) {
                          final alert = alerts[index];
                          return _buildAlertCard(context, alert, api);
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

  Widget _buildAlertCard(BuildContext context, Map<String, dynamic> alert, AdminApiClient api) {
    final id = alert['id'] as int? ?? 0;
    final alertType = alert['alert_type'] as String? ?? '';
    final severity = alert['severity'] as String? ?? '';
    final title = alert['title'] as String? ?? '';
    final status = alert['status'] as String? ?? '';
    final createdAt = alert['created_at'] as String? ?? '';

    Color severityColor;
    IconData severityIcon;
    switch (severity) {
      case 'critical':
        severityColor = Colors.red;
        severityIcon = Icons.error;
        break;
      case 'warning':
        severityColor = Colors.orange;
        severityIcon = Icons.warning;
        break;
      default:
        severityColor = Colors.blue;
        severityIcon = Icons.info;
    }

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: ListTile(
        leading: CircleAvatar(
          backgroundColor: severityColor.withValues(alpha: 0.1),
          child: Icon(severityIcon, color: severityColor),
        ),
        title: Text(title),
        subtitle: Text('类型: $alertType | 时间: $createdAt'),
        trailing: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: status == 'active' ? Colors.red.withValues(alpha: 0.1) : Colors.grey.withValues(alpha: 0.1),
                borderRadius: BorderRadius.circular(4),
              ),
              child: Text(
                status == 'active' ? '活跃' : '已确认',
                style: TextStyle(color: status == 'active' ? Colors.red : Colors.grey),
              ),
            ),
            if (status == 'active') ...[
              const SizedBox(width: 8),
              IconButton(
                icon: const Icon(Icons.check, color: Colors.green),
                onPressed: () => _acknowledgeAlert(api, id),
                tooltip: '确认',
              ),
            ],
          ],
        ),
        onTap: () => _showAlertDetails(context, alert),
      ),
    );
  }

  Future<void> _acknowledgeAlert(AdminApiClient api, int id) async {
    try {
      await api.acknowledgeAlert(id);
      setState(() {}); // 刷新列表
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          const SnackBar(content: Text('告警已确认')),
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

  void _showAlertDetails(BuildContext context, Map<String, dynamic> alert) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: const Text('告警详情'),
        content: SingleChildScrollView(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              Text('ID: ${alert['id']}'),
              Text('类型: ${alert['alert_type']}'),
              Text('严重程度: ${alert['severity']}'),
              Text('状态: ${alert['status']}'),
              Text('标题: ${alert['title']}'),
              Text('创建时间: ${alert['created_at']}'),
              const SizedBox(height: 8),
              const Text('详情:', style: TextStyle(fontWeight: FontWeight.bold)),
              Text(alert['details'] ?? '无'),
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
