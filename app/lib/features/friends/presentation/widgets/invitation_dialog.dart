import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import '../../../../core/services/api_client.dart';

// 房间邀请模型
class RoomInvitation {
  final int id;
  final int roomId;
  final String roomName;
  final String betAmount;
  final int playerCount;
  final int fromUserId;
  final String fromUsername;

  RoomInvitation({
    required this.id,
    required this.roomId,
    required this.roomName,
    required this.betAmount,
    required this.playerCount,
    required this.fromUserId,
    required this.fromUsername,
  });

  factory RoomInvitation.fromJson(Map<String, dynamic> json) {
    return RoomInvitation(
      id: json['id'] as int,
      roomId: json['room_id'] as int,
      roomName: json['room_name'] as String? ?? 'Room',
      betAmount: json['bet_amount']?.toString() ?? '0',
      playerCount: json['player_count'] as int? ?? 0,
      fromUserId: json['from_user_id'] as int,
      fromUsername: json['from_username'] as String? ?? 'Unknown',
    );
  }
}

/// 邀请通知弹窗
class InvitationDialog extends ConsumerWidget {
  final RoomInvitation invitation;

  const InvitationDialog({super.key, required this.invitation});

  @override
  Widget build(BuildContext context, WidgetRef ref) {
    return AlertDialog(
      title: const Text('房间邀请'),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            '${invitation.fromUsername} 邀请你加入房间',
            style: const TextStyle(fontSize: 16),
          ),
          const SizedBox(height: 16),
          _InfoRow(label: '房间名称', value: invitation.roomName),
          _InfoRow(label: '下注金额', value: '¥${invitation.betAmount}'),
          _InfoRow(label: '当前人数', value: '${invitation.playerCount}人'),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () async {
            try {
              await ref.read(apiClientProvider).declineInvitation(invitation.id);
              if (context.mounted) {
                Navigator.pop(context, false);
              }
            } catch (e) {
              if (context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('操作失败: $e')),
                );
              }
            }
          },
          child: const Text('拒绝'),
        ),
        ElevatedButton(
          onPressed: () async {
            try {
              await ref.read(apiClientProvider).acceptInvitation(invitation.id);
              if (context.mounted) {
                Navigator.pop(context, true);
                // 导航到房间页面
                Navigator.pushNamed(context, '/room', arguments: invitation.roomId);
              }
            } catch (e) {
              if (context.mounted) {
                ScaffoldMessenger.of(context).showSnackBar(
                  SnackBar(content: Text('加入失败: $e')),
                );
              }
            }
          },
          child: const Text('接受'),
        ),
      ],
    );
  }
}

class _InfoRow extends StatelessWidget {
  final String label;
  final String value;

  const _InfoRow({required this.label, required this.value});

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(color: Colors.grey[600])),
          Text(value, style: const TextStyle(fontWeight: FontWeight.bold)),
        ],
      ),
    );
  }
}

/// 显示邀请弹窗
Future<bool?> showInvitationDialog(BuildContext context, RoomInvitation invitation) {
  return showDialog<bool>(
    context: context,
    builder: (context) => InvitationDialog(invitation: invitation),
  );
}

/// 邀请链接分享对话框
class InviteLinkDialog extends ConsumerStatefulWidget {
  final int roomId;

  const InviteLinkDialog({super.key, required this.roomId});

  @override
  ConsumerState<InviteLinkDialog> createState() => _InviteLinkDialogState();
}

class _InviteLinkDialogState extends ConsumerState<InviteLinkDialog> {
  String? _inviteCode;
  String? _inviteLink;
  bool _isLoading = false;
  String? _error;

  @override
  void initState() {
    super.initState();
    _createInviteLink();
  }

  Future<void> _createInviteLink() async {
    setState(() {
      _isLoading = true;
      _error = null;
    });

    try {
      final result = await ref.read(apiClientProvider).createInviteLink(widget.roomId);
      setState(() {
        _inviteCode = result['code'] as String?;
        _inviteLink = result['link'] as String?;
        _isLoading = false;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: const Text('邀请链接'),
      content: _isLoading
          ? const SizedBox(
              height: 100,
              child: Center(child: CircularProgressIndicator()),
            )
          : _error != null
              ? Text('创建失败: $_error')
              : Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const Text('分享此链接邀请好友加入房间'),
                    const SizedBox(height: 16),
                    Container(
                      padding: const EdgeInsets.all(12),
                      decoration: BoxDecoration(
                        color: Colors.grey[100],
                        borderRadius: BorderRadius.circular(8),
                      ),
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              _inviteCode ?? '',
                              style: const TextStyle(
                                fontFamily: 'monospace',
                                fontSize: 14,
                              ),
                            ),
                          ),
                          IconButton(
                            icon: const Icon(Icons.copy),
                            onPressed: () {
                              Clipboard.setData(ClipboardData(text: _inviteCode ?? ''));
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(content: Text('已复制到剪贴板')),
                              );
                            },
                          ),
                        ],
                      ),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '链接24小时内有效',
                      style: TextStyle(color: Colors.grey[600], fontSize: 12),
                    ),
                  ],
                ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: const Text('关闭'),
        ),
      ],
    );
  }
}

/// 显示邀请链接对话框
Future<void> showInviteLinkDialog(BuildContext context, int roomId) {
  return showDialog(
    context: context,
    builder: (context) => InviteLinkDialog(roomId: roomId),
  );
}
