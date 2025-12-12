import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:five_seconds_go_admin/l10n/app_localizations.dart';

import '../../../../core/services/api_client.dart';

class UsersPage extends ConsumerStatefulWidget {
  const UsersPage({super.key});

  @override
  ConsumerState<UsersPage> createState() => _UsersPageState();
}

class _UsersPageState extends ConsumerState<UsersPage> {
  final _searchController = TextEditingController();
  String _roleFilter = 'all';

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final l10n = AppLocalizations.of(context)!;
    final api = ref.read(adminApiClientProvider);

    return Scaffold(
      appBar: AppBar(
        title: Text(l10n.usersTitle),
        actions: [
          FilledButton.icon(
            onPressed: () {
              _showCreateOwnerDialog(context, ref);
            },
            icon: const Icon(Icons.add),
            label: Text(l10n.usersCreateOwner),
          ),
          const SizedBox(width: 16),
        ],
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
                          hintText: l10n.usersSearchHint,
                          prefixIcon: const Icon(Icons.search),
                          border: const OutlineInputBorder(),
                        ),
                        onSubmitted: (_) => setState(() {}),
                      ),
                    ),
                    const SizedBox(width: 16),
                    DropdownButton<String>(
                      value: _roleFilter,
                      items: [
                        DropdownMenuItem(
                          value: 'all',
                          child: Text(l10n.usersFilterAllRoles),
                        ),
                        DropdownMenuItem(
                          value: 'player',
                          child: Text(l10n.usersFilterRolePlayer),
                        ),
                        DropdownMenuItem(
                          value: 'owner',
                          child: Text(l10n.usersFilterRoleOwner),
                        ),
                        DropdownMenuItem(
                          value: 'admin',
                          child: Text(l10n.usersFilterRoleAdmin),
                        ),
                      ],
                      onChanged: (value) {
                        if (value == null) return;
                        setState(() {
                          _roleFilter = value;
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
                  future: api.listUsers(
                    role: _roleFilter,
                    search: _searchController.text.trim().isEmpty
                        ? null
                        : _searchController.text.trim(),
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
                        final user = items[index];
                        final username = user['username'] as String? ?? '-';
                        final role = user['role'] as String? ?? '-';
                        final inviteCode = user['invite_code'] as String?;
                        final invitedBy = user['invited_by'];
                        final balance = user['balance'] ?? '0';

                        final isOwner = role == 'owner';
                        final avatarText = username.isNotEmpty
                            ? username.substring(0, username.length >= 2 ? 2 : 1).toUpperCase()
                            : '?';

                        return ListTile(
                          leading: CircleAvatar(
                            backgroundColor: isOwner
                                ? Colors.orange[100]
                                : Colors.blue[100],
                            child: Text(
                              avatarText,
                              style: TextStyle(
                                color: isOwner
                                    ? Colors.orange[800]
                                    : Colors.blue[800],
                              ),
                            ),
                          ),
                          title: Text(username),
                          subtitle: Text(
                            isOwner
                                ? l10n.usersItemSubtitleOwner(
                                    inviteCode ?? '--',
                                  )
                                : l10n.usersItemSubtitlePlayer(
                                    invitedBy?.toString() ?? '-',
                                  ),
                          ),
                          trailing: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Text(
                                '¥$balance',
                                style: const TextStyle(fontWeight: FontWeight.bold),
                              ),
                              const SizedBox(width: 16),
                              IconButton(
                                icon: const Icon(Icons.more_vert),
                                onPressed: () {},
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

  void _showCreateOwnerDialog(BuildContext context, WidgetRef ref) {
    final usernameController = TextEditingController();
    final passwordController = TextEditingController();

    showDialog(
      context: context,
      builder: (dialogContext) {
        final l10n = AppLocalizations.of(dialogContext)!;
        return AlertDialog(
          title: Text(l10n.usersDialogCreateOwnerTitle),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              TextField(
                controller: usernameController,
                decoration: InputDecoration(
                  labelText: l10n.usersDialogUsernameLabel,
                  border: const OutlineInputBorder(),
                ),
              ),
              const SizedBox(height: 16),
              TextField(
                controller: passwordController,
                obscureText: true,
                decoration: InputDecoration(
                  labelText: l10n.usersDialogPasswordLabel,
                  border: const OutlineInputBorder(),
                ),
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(dialogContext),
              child: Text(l10n.usersDialogCancel),
            ),
            FilledButton(
              onPressed: () async {
                final username = usernameController.text.trim();
                final password = passwordController.text;
                if (username.isEmpty || password.length < 6) {
                  ScaffoldMessenger.of(dialogContext).showSnackBar(
                    SnackBar(content: Text(l10n.loginPasswordTooShort)),
                  );
                  return;
                }
                try {
                  await ref.read(adminApiClientProvider).createOwner(
                        username: username,
                        password: password,
                      );
                  if (mounted) {
                    Navigator.pop(dialogContext);
                    setState(() {}); // 刷新列表
                  }
                } catch (e) {
                  ScaffoldMessenger.of(dialogContext).showSnackBar(
                    SnackBar(content: Text(e.toString())),
                  );
                }
              },
              child: Text(l10n.usersDialogCreate),
            ),
          ],
        );
      },
    );
  }
}
