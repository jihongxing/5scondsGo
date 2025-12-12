import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:fl_chart/fl_chart.dart';
import 'package:five_seconds_go_admin/core/services/api_client.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';

class MonitoringDashboardPage extends ConsumerStatefulWidget {
  const MonitoringDashboardPage({super.key});

  @override
  ConsumerState<MonitoringDashboardPage> createState() =>
      _MonitoringDashboardPageState();
}

class _MonitoringDashboardPageState
    extends ConsumerState<MonitoringDashboardPage> {
  Timer? _refreshTimer;
  Map<String, dynamic>? _metrics;
  bool _isLoading = true;
  String? _error;
  String _selectedTimeRange = '1h';

  @override
  void initState() {
    super.initState();
    _loadMetrics();
    // 每10秒自动刷新
    _refreshTimer = Timer.periodic(const Duration(seconds: 10), (_) {
      _loadMetrics();
    });
  }

  @override
  void dispose() {
    _refreshTimer?.cancel();
    super.dispose();
  }

  Future<void> _loadMetrics() async {
    try {
      final api = ref.read(adminApiClientProvider);
      final metrics = await api.getRealtimeMetrics();
      if (mounted) {
        setState(() {
          _metrics = metrics;
          _isLoading = false;
          _error = null;
        });
      }
    } catch (e) {
      if (mounted) {
        setState(() {
          _error = e.toString();
          _isLoading = false;
        });
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.monitoringTitle),
        actions: [
          IconButton(
            icon: const Icon(Icons.refresh),
            onPressed: _loadMetrics,
            tooltip: l10n.refresh,
          ),
        ],
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _error != null
              ? Center(child: Text('Error: $_error'))
              : _buildContent(context),
    );
  }

  Widget _buildContent(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final metrics = _metrics ?? {};

    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // 实时指标卡片
          Text(
            l10n.monitoringRealtimeMetrics,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          _buildMetricsGrid(context, metrics),
          const SizedBox(height: 32),

          // 性能指标
          Text(
            l10n.monitoringPerformanceMetrics,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          _buildPerformanceCards(context, metrics),
          const SizedBox(height: 32),

          // 历史趋势
          Text(
            l10n.monitoringHistoricalTrends,
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          _buildTimeRangeSelector(),
          const SizedBox(height: 16),
          _buildHistoryChart(context),
        ],
      ),
    );
  }

  Widget _buildMetricsGrid(BuildContext context, Map<String, dynamic> metrics) {
    final l10n = AppLocalizations.of(context)!;

    return Wrap(
      spacing: 16,
      runSpacing: 16,
      children: [
        _buildMetricCard(
          context,
          title: l10n.monitoringOnlinePlayers,
          value: '${metrics['online_players'] ?? 0}',
          icon: Icons.people,
          color: Colors.blue,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringActiveRooms,
          value: '${metrics['active_rooms'] ?? 0}',
          icon: Icons.meeting_room,
          color: Colors.green,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringGamesPerMinute,
          value: '${(metrics['games_per_minute'] ?? 0).toStringAsFixed(1)}',
          icon: Icons.speed,
          color: Colors.orange,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringDailyActiveUsers,
          value: '${metrics['daily_active_users'] ?? 0}',
          icon: Icons.trending_up,
          color: Colors.purple,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringDailyVolume,
          value: '¥${metrics['daily_volume'] ?? '0'}',
          icon: Icons.attach_money,
          color: Colors.teal,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringPlatformRevenue,
          value: '¥${metrics['platform_revenue'] ?? '0'}',
          icon: Icons.account_balance_wallet,
          color: Colors.indigo,
        ),
      ],
    );
  }

  Widget _buildMetricCard(
    BuildContext context, {
    required String title,
    required String value,
    required IconData icon,
    required Color color,
    bool isAlert = false,
  }) {
    return SizedBox(
      width: 200,
      child: Card(
        color: isAlert ? Colors.red[50] : null,
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    padding: const EdgeInsets.all(8),
                    decoration: BoxDecoration(
                      color: color.withOpacity(0.1),
                      borderRadius: BorderRadius.circular(8),
                    ),
                    child: Icon(icon, color: color, size: 20),
                  ),
                  if (isAlert) ...[
                    const Spacer(),
                    Icon(Icons.warning, color: Colors.red[700], size: 20),
                  ],
                ],
              ),
              const SizedBox(height: 12),
              Text(
                value,
                style: Theme.of(context).textTheme.headlineSmall?.copyWith(
                      fontWeight: FontWeight.bold,
                      color: isAlert ? Colors.red[700] : null,
                    ),
              ),
              const SizedBox(height: 4),
              Text(
                title,
                style: Theme.of(context).textTheme.bodySmall?.copyWith(
                      color: Colors.grey[600],
                    ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildPerformanceCards(
      BuildContext context, Map<String, dynamic> metrics) {
    final l10n = AppLocalizations.of(context)!;

    final apiLatency = (metrics['api_latency_p95'] ?? 0).toDouble();
    final wsLatency = (metrics['ws_latency_p95'] ?? 0).toDouble();
    final dbLatency = (metrics['db_latency_p95'] ?? 0).toDouble();

    // 阈值检查
    const apiThreshold = 500.0;
    const wsThreshold = 200.0;
    const dbThreshold = 100.0;

    return Wrap(
      spacing: 16,
      runSpacing: 16,
      children: [
        _buildMetricCard(
          context,
          title: l10n.monitoringApiLatencyP95,
          value: '${apiLatency.toStringAsFixed(1)} ms',
          icon: Icons.api,
          color: Colors.blue,
          isAlert: apiLatency > apiThreshold,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringWsLatencyP95,
          value: '${wsLatency.toStringAsFixed(1)} ms',
          icon: Icons.sync_alt,
          color: Colors.green,
          isAlert: wsLatency > wsThreshold,
        ),
        _buildMetricCard(
          context,
          title: l10n.monitoringDbLatencyP95,
          value: '${dbLatency.toStringAsFixed(1)} ms',
          icon: Icons.storage,
          color: Colors.orange,
          isAlert: dbLatency > dbThreshold,
        ),
      ],
    );
  }

  Widget _buildTimeRangeSelector() {
    return SegmentedButton<String>(
      segments: const [
        ButtonSegment(value: '1h', label: Text('1H')),
        ButtonSegment(value: '24h', label: Text('24H')),
        ButtonSegment(value: '7d', label: Text('7D')),
        ButtonSegment(value: '30d', label: Text('30D')),
      ],
      selected: {_selectedTimeRange},
      onSelectionChanged: (selection) {
        setState(() {
          _selectedTimeRange = selection.first;
        });
      },
    );
  }

  Widget _buildHistoryChart(BuildContext context) {
    final api = ref.read(adminApiClientProvider);
    final l10n = AppLocalizations.of(context)!;

    return SizedBox(
      height: 400,
      child: Card(
        child: Padding(
          padding: const EdgeInsets.all(16),
          child: FutureBuilder<Map<String, dynamic>>(
            future: api.getHistoricalMetrics(timeRange: _selectedTimeRange),
            builder: (context, snapshot) {
              if (snapshot.connectionState == ConnectionState.waiting) {
                return const Center(child: CircularProgressIndicator());
              }
              if (snapshot.hasError) {
                return Center(child: Text('Error: ${snapshot.error}'));
              }

              final items =
                  (snapshot.data?['items'] as List<dynamic>?) ?? [];
              if (items.isEmpty) {
                return const Center(child: Text('No historical data'));
              }

              return _buildMetricsLineChart(context, items);
            },
          ),
        ),
      ),
    );
  }

  Widget _buildMetricsLineChart(BuildContext context, List<dynamic> items) {
    final l10n = AppLocalizations.of(context)!;
    
    // 准备数据点
    final List<FlSpot> playerSpots = [];
    final List<FlSpot> roomSpots = [];
    final List<FlSpot> gpmSpots = [];
    
    // 限制数据点数量，避免图表过于密集
    final maxPoints = 50;
    final step = items.length > maxPoints ? items.length ~/ maxPoints : 1;
    
    double maxPlayers = 0;
    double maxRooms = 0;
    double maxGpm = 0;
    
    for (int i = 0; i < items.length; i += step) {
      final item = items[i] as Map<String, dynamic>;
      final x = i.toDouble();
      
      final players = (item['online_players'] ?? 0).toDouble();
      final rooms = (item['active_rooms'] ?? 0).toDouble();
      final gpm = (item['games_per_minute'] ?? 0).toDouble();
      
      playerSpots.add(FlSpot(x, players));
      roomSpots.add(FlSpot(x, rooms));
      gpmSpots.add(FlSpot(x, gpm));
      
      if (players > maxPlayers) maxPlayers = players;
      if (rooms > maxRooms) maxRooms = rooms;
      if (gpm > maxGpm) maxGpm = gpm;
    }
    
    // 计算最大Y值
    final maxY = [maxPlayers, maxRooms, maxGpm * 10].reduce((a, b) => a > b ? a : b);
    final double adjustedMaxY = maxY > 0 ? (maxY * 1.2).toDouble() : 10.0;
    
    return Column(
      children: [
        // 图例
        Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            _buildLegendItem(l10n.monitoringOnlinePlayers, Colors.blue),
            const SizedBox(width: 24),
            _buildLegendItem(l10n.monitoringActiveRooms, Colors.green),
            const SizedBox(width: 24),
            _buildLegendItem(l10n.monitoringGamesPerMinute, Colors.orange),
          ],
        ),
        const SizedBox(height: 16),
        // 图表
        Expanded(
          child: LineChart(
            LineChartData(
              gridData: FlGridData(
                show: true,
                drawVerticalLine: true,
                horizontalInterval: adjustedMaxY / 5,
                getDrawingHorizontalLine: (value) {
                  return FlLine(
                    color: Colors.grey.withOpacity(0.2),
                    strokeWidth: 1,
                  );
                },
                getDrawingVerticalLine: (value) {
                  return FlLine(
                    color: Colors.grey.withOpacity(0.2),
                    strokeWidth: 1,
                  );
                },
              ),
              titlesData: FlTitlesData(
                show: true,
                rightTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                topTitles: const AxisTitles(
                  sideTitles: SideTitles(showTitles: false),
                ),
                bottomTitles: AxisTitles(
                  sideTitles: SideTitles(
                    showTitles: true,
                    reservedSize: 30,
                    interval: items.length > 10 ? items.length / 5 : 1,
                    getTitlesWidget: (value, meta) {
                      final index = value.toInt();
                      if (index >= 0 && index < items.length) {
                        final item = items[index] as Map<String, dynamic>;
                        final createdAt = item['created_at']?.toString() ?? '';
                        if (createdAt.length >= 16) {
                          return Padding(
                            padding: const EdgeInsets.only(top: 8),
                            child: Text(
                              createdAt.substring(11, 16),
                              style: const TextStyle(fontSize: 10),
                            ),
                          );
                        }
                      }
                      return const Text('');
                    },
                  ),
                ),
                leftTitles: AxisTitles(
                  sideTitles: SideTitles(
                    showTitles: true,
                    interval: adjustedMaxY / 5,
                    reservedSize: 42,
                    getTitlesWidget: (value, meta) {
                      return Text(
                        value.toInt().toString(),
                        style: const TextStyle(fontSize: 10),
                      );
                    },
                  ),
                ),
              ),
              borderData: FlBorderData(
                show: true,
                border: Border.all(color: Colors.grey.withOpacity(0.3)),
              ),
              minX: 0,
              maxX: (items.length - 1).toDouble(),
              minY: 0,
              maxY: adjustedMaxY,
              lineBarsData: [
                // 在线玩家
                LineChartBarData(
                  spots: playerSpots,
                  isCurved: true,
                  color: Colors.blue,
                  barWidth: 2,
                  isStrokeCapRound: true,
                  dotData: const FlDotData(show: false),
                  belowBarData: BarAreaData(
                    show: true,
                    color: Colors.blue.withOpacity(0.1),
                  ),
                ),
                // 活跃房间
                LineChartBarData(
                  spots: roomSpots,
                  isCurved: true,
                  color: Colors.green,
                  barWidth: 2,
                  isStrokeCapRound: true,
                  dotData: const FlDotData(show: false),
                  belowBarData: BarAreaData(
                    show: true,
                    color: Colors.green.withOpacity(0.1),
                  ),
                ),
                // 每分钟游戏数 (放大10倍以便在同一图表中显示)
                LineChartBarData(
                  spots: gpmSpots.map((s) => FlSpot(s.x, s.y * 10)).toList(),
                  isCurved: true,
                  color: Colors.orange,
                  barWidth: 2,
                  isStrokeCapRound: true,
                  dotData: const FlDotData(show: false),
                  belowBarData: BarAreaData(
                    show: true,
                    color: Colors.orange.withOpacity(0.1),
                  ),
                ),
              ],
              lineTouchData: LineTouchData(
                touchTooltipData: LineTouchTooltipData(
                  getTooltipItems: (touchedSpots) {
                    return touchedSpots.map((spot) {
                      String label;
                      String value;
                      if (spot.barIndex == 0) {
                        label = l10n.monitoringOnlinePlayers;
                        value = spot.y.toInt().toString();
                      } else if (spot.barIndex == 1) {
                        label = l10n.monitoringActiveRooms;
                        value = spot.y.toInt().toString();
                      } else {
                        label = l10n.monitoringGamesPerMinute;
                        value = (spot.y / 10).toStringAsFixed(1);
                      }
                      return LineTooltipItem(
                        '$label: $value',
                        TextStyle(color: spot.bar.color),
                      );
                    }).toList();
                  },
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Widget _buildLegendItem(String label, Color color) {
    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Container(
          width: 16,
          height: 3,
          decoration: BoxDecoration(
            color: color,
            borderRadius: BorderRadius.circular(2),
          ),
        ),
        const SizedBox(width: 4),
        Text(
          label,
          style: TextStyle(
            fontSize: 12,
            color: Colors.grey[600],
          ),
        ),
      ],
    );
  }
}
