import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';

import '../../../../core/services/api_client.dart';

class RoomsPage extends ConsumerStatefulWidget {
  const RoomsPage({super.key});

  @override
  ConsumerState<RoomsPage> createState() => _RoomsPageState();
}

class _RoomsPageState extends ConsumerState<RoomsPage> {
  final _searchController = TextEditingController();
  String _statusFilter = 'all';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<void> _changeStatus(
    BuildContext context, {
    required int roomId,
    required String newStatus,
  }) async {
    if (roomId <= 0) return;
    final api = ref.read(adminApiClientProvider);

    try {
      await api.updateRoomStatus(id: roomId, status: newStatus);
      if (!mounted) return;
      setState(() {});
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Room #$roomId status -> $newStatus')),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Failed to update: $e')),
      );
    }
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final api = ref.read(adminApiClientProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.roomsTitle),
      ),
      body: Padding(
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
                        controller: _searchController,
                        decoration: InputDecoration(
                          hintText: l10n.roomsSearchHint,
                          prefixIcon: const Icon(Icons.search),
                          border: const OutlineInputBorder(),
                        ),
                        onChanged: (_) => setState(() {}),
                      ),
                    ),
                    const SizedBox(width: 16),
                    DropdownButton<String>(
                      value: _statusFilter,
                      items: [
                        DropdownMenuItem(
                          value: 'all',
                          child: Text(l10n.roomsFilterAllStatus),
                        ),
                        DropdownMenuItem(
                          value: 'active',
                          child: Text(l10n.roomsFilterStatusActive),
                        ),
                        DropdownMenuItem(
                          value: 'paused',
                          child: Text(l10n.roomsFilterStatusPaused),
                        ),
                        DropdownMenuItem(
                          value: 'locked',
                          child: Text(l10n.roomsFilterStatusLocked),
                        ),
                      ],
                      onChanged: (value) {
                        if (value == null) return;
                        setState(() {
                          _statusFilter = value;
                        });
                      },
                    ),
                  ],
                ),
              ),
              const Divider(height: 1),
              // Table
              Expanded(
                child: FutureBuilder<Map<String, dynamic>>(
                  future: api.listRooms(
                    status: _statusFilter,
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

                    final search = _searchController.text.trim().toLowerCase();
                    if (search.isNotEmpty) {
                      items = items
                          .where((room) => (room['name'] as String? ?? '')
                              .toLowerCase()
                              .contains(search))
                          .toList();
                    }

                    if (items.isEmpty) {
                      return const Center(child: Text('No data'));
                    }

                    return ListView.builder(
                      itemCount: items.length,
                      itemBuilder: (context, index) {
                        final room = items[index];
                        final name = room['name'] as String? ?? '-';
                        final id = room['id']?.toString() ?? '';
                        final status = room['status'] as String? ?? 'active';
                        final betAmount = room['bet_amount'] ?? '0';
                        final ownerId = room['owner_id']?.toString() ?? '-';
                        final currentPlayers = room['current_players']?.toString() ?? '0';

                        Color bgColor;
                        Color fgColor;
                        String statusLabel;
                        switch (status) {
                          case 'paused':
                            bgColor = Colors.orange[100]!;
                            fgColor = Colors.orange;
                            statusLabel = l10n.roomsStatusPaused;
                            break;
                          case 'locked':
                            bgColor = Colors.red[100]!;
                            fgColor = Colors.red;
                            statusLabel = l10n.roomsStatusLocked;
                            break;
                          case 'active':
                          default:
                            bgColor = Colors.green[100]!;
                            fgColor = Colors.green;
                            statusLabel = l10n.roomsStatusActive;
                            break;
                        }

                        return ListTile(
                          leading: Container(
                            width: 48,
                            height: 48,
                            decoration: BoxDecoration(
                              color: bgColor,
                              borderRadius: BorderRadius.circular(8),
                            ),
                            child: Icon(
                              Icons.meeting_room,
                              color: fgColor,
                            ),
                          ),
                          title: Text('$name (#$id)'),
                          subtitle: Text(
                            l10n.roomsItemSubtitle(
                              '¥$betAmount',
                              room['owner_name']?.toString().isNotEmpty == true
                                  ? room['owner_name'].toString()
                                  : 'Owner #$ownerId',
                              currentPlayers,
                            ),
                          ),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 12,
                                  vertical: 4,
                                ),
                                decoration: BoxDecoration(
                                  color: bgColor,
                                  borderRadius: BorderRadius.circular(16),
                                ),
                                child: Text(
                                  statusLabel,
                                  style: TextStyle(
                                    color: fgColor.withOpacity(0.9),
                                    fontWeight: FontWeight.bold,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 16),
                              PopupMenuButton<String>(
                                icon: const Icon(Icons.more_vert),
                                onSelected: (value) async {
                                  await _changeStatus(context,
                                      roomId: int.tryParse(id) ?? 0,
                                      newStatus: value);
                                },
                                itemBuilder: (context) => [
                                  if (status != 'active')
                                    PopupMenuItem(
                                      value: 'active',
                                      child: Text(l10n.roomsStatusActive),
                                    ),
                                  if (status != 'paused')
                                    PopupMenuItem(
                                      value: 'paused',
                                      child: Text(l10n.roomsStatusPaused),
                                    ),
                                  if (status != 'locked')
                                    PopupMenuItem(
                                      value: 'locked',
                                      child: Text(l10n.roomsStatusLocked),
                                    ),
                                ],
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
      ),
    );
  }
}
