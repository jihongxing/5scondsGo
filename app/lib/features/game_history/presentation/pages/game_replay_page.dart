import 'dart:async';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';
import 'package:go_router/go_router.dart';
import '../../../../core/services/api_client.dart';
import '../../../../core/theme/app_theme.dart';

class GameReplayPage extends ConsumerStatefulWidget {
  final int roundId;

  const GameReplayPage({super.key, required this.roundId});

  @override
  ConsumerState<GameReplayPage> createState() => _GameReplayPageState();
}

class _GameReplayPageState extends ConsumerState<GameReplayPage> {
  ReplayData? _replayData;
  VerificationResult? _verificationResult;
  bool _isLoading = true;
  bool _isVerifying = false;
  int _currentPhaseIndex = 0;
  Timer? _replayTimer;
  bool _isPlaying = false;

  @override
  void initState() {
    super.initState();
    _loadReplayData();
  }

  @override
  void dispose() {
    _replayTimer?.cancel();
    super.dispose();
  }

  Future<void> _loadReplayData() async {
    setState(() => _isLoading = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      final response = await apiClient.get('/game-rounds/${widget.roundId}/replay');
      setState(() {
        _replayData = ReplayData.fromJson(response);
        _isLoading = false;
      });
    } catch (e) {
      setState(() => _isLoading = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('加载失败: $e')),
        );
      }
    }
  }

  Future<void> _verifyRound() async {
    setState(() => _isVerifying = true);
    try {
      final apiClient = ref.read(apiClientProvider);
      final response = await apiClient.get('/game-rounds/${widget.roundId}/verify');
      setState(() {
        _verificationResult = VerificationResult.fromJson(response);
        _isVerifying = false;
      });
    } catch (e) {
      setState(() => _isVerifying = false);
      if (mounted) {
        ScaffoldMessenger.of(context).showSnackBar(
          SnackBar(content: Text('验证失败: $e')),
        );
      }
    }
  }


  void _startReplay() {
    if (_replayData == null || _replayData!.phases.isEmpty) return;
    
    setState(() {
      _isPlaying = true;
      _currentPhaseIndex = 0;
    });

    _playNextPhase();
  }

  void _playNextPhase() {
    if (!_isPlaying || _replayData == null) return;
    
    if (_currentPhaseIndex >= _replayData!.phases.length) {
      setState(() => _isPlaying = false);
      return;
    }

    final phase = _replayData!.phases[_currentPhaseIndex];
    _replayTimer = Timer(Duration(seconds: phase.duration), () {
      setState(() => _currentPhaseIndex++);
      _playNextPhase();
    });
  }

  void _stopReplay() {
    _replayTimer?.cancel();
    setState(() => _isPlaying = false);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text('回合 #${widget.roundId}'),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () => context.pop(),
        ),
      ),
      body: _isLoading
          ? const Center(child: CircularProgressIndicator())
          : _replayData == null
              ? const Center(child: Text('无法加载回放数据'))
              : SingleChildScrollView(
                  padding: EdgeInsets.all(16.w),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildInfoCard(),
                      SizedBox(height: 16.h),
                      _buildReplayControl(),
                      SizedBox(height: 16.h),
                      _buildParticipantsCard(),
                      SizedBox(height: 16.h),
                      _buildVerificationCard(),
                    ],
                  ),
                ),
    );
  }

  Widget _buildInfoCard() {
    final data = _replayData!;
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16.w),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              data.roomName,
              style: TextStyle(fontSize: 18.sp, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 8.h),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _buildInfoItem('回合', '#${data.roundNumber}'),
                _buildInfoItem('投注额', '¥${data.betAmount}'),
                _buildInfoItem('奖池', '¥${data.poolAmount}'),
              ],
            ),
            SizedBox(height: 8.h),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                _buildInfoItem('每人奖金', '¥${data.prizePerWinner}'),
                _buildInfoItem('赢家数', '${data.winners.length}'),
                _buildInfoItem('参与者', '${data.participants.length}'),
              ],
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoItem(String label, String value) {
    return Column(
      children: [
        Text(value, style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.w500)),
        Text(label, style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary)),
      ],
    );
  }

  Widget _buildReplayControl() {
    final phases = _replayData?.phases ?? [];
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16.w),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('回放', style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold)),
                Row(
                  children: [
                    IconButton(
                      onPressed: _isPlaying ? _stopReplay : _startReplay,
                      icon: Icon(_isPlaying ? Icons.stop : Icons.play_arrow),
                    ),
                  ],
                ),
              ],
            ),
            SizedBox(height: 8.h),
            Row(
              children: phases.asMap().entries.map((entry) {
                final index = entry.key;
                final phase = entry.value;
                final isActive = index == _currentPhaseIndex && _isPlaying;
                final isPast = index < _currentPhaseIndex;
                
                return Expanded(
                  child: Container(
                    margin: EdgeInsets.symmetric(horizontal: 2.w),
                    padding: EdgeInsets.symmetric(vertical: 8.h),
                    decoration: BoxDecoration(
                      color: isActive
                          ? AppColors.primary
                          : isPast
                              ? AppColors.primary.withAlpha(128)
                              : Colors.grey.withAlpha(51),
                      borderRadius: BorderRadius.circular(4.r),
                    ),
                    child: Center(
                      child: Text(
                        _getPhaseLabel(phase.phase),
                        style: TextStyle(
                          fontSize: 10.sp,
                          color: isActive || isPast ? Colors.white : null,
                        ),
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildParticipantsCard() {
    final data = _replayData!;
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16.w),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text('参与者', style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold)),
            SizedBox(height: 8.h),
            Wrap(
              spacing: 8.w,
              runSpacing: 8.h,
              children: data.participants.map((p) {
                return Chip(
                  avatar: p.isWinner
                      ? const Icon(Icons.emoji_events, size: 16, color: Colors.amber)
                      : null,
                  label: Text(p.username),
                  backgroundColor: p.isWinner ? Colors.amber.withAlpha(51) : null,
                );
              }).toList(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVerificationCard() {
    final data = _replayData!;
    return Card(
      child: Padding(
        padding: EdgeInsets.all(16.w),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text('公平性验证', style: TextStyle(fontSize: 16.sp, fontWeight: FontWeight.bold)),
                ElevatedButton(
                  onPressed: _isVerifying ? null : _verifyRound,
                  child: _isVerifying
                      ? SizedBox(
                          width: 16.w,
                          height: 16.w,
                          child: const CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('验证'),
                ),
              ],
            ),
            SizedBox(height: 8.h),
            _buildHashItem('Commit Hash', data.commitHash),
            _buildHashItem('Reveal Seed', data.revealSeed),
            if (_verificationResult != null) ...[
              SizedBox(height: 8.h),
              Container(
                padding: EdgeInsets.all(12.w),
                decoration: BoxDecoration(
                  color: _verificationResult!.isValid
                      ? AppColors.success.withAlpha(25)
                      : AppColors.error.withAlpha(25),
                  borderRadius: BorderRadius.circular(8.r),
                ),
                child: Row(
                  children: [
                    Icon(
                      _verificationResult!.isValid ? Icons.check_circle : Icons.error,
                      color: _verificationResult!.isValid ? AppColors.success : AppColors.error,
                    ),
                    SizedBox(width: 8.w),
                    Text(
                      _verificationResult!.isValid ? '验证通过！结果公平' : '验证失败！结果可能被篡改',
                      style: TextStyle(
                        color: _verificationResult!.isValid ? AppColors.success : AppColors.error,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ],
        ),
      ),
    );
  }

  Widget _buildHashItem(String label, String value) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: TextStyle(fontSize: 12.sp, color: AppColors.textSecondary)),
        SizedBox(height: 4.h),
        Container(
          padding: EdgeInsets.all(8.w),
          decoration: BoxDecoration(
            color: Colors.grey.withAlpha(25),
            borderRadius: BorderRadius.circular(4.r),
          ),
          child: Text(
            value.isEmpty ? '(未公开)' : value,
            style: TextStyle(fontSize: 10.sp, fontFamily: 'monospace'),
          ),
        ),
        SizedBox(height: 8.h),
      ],
    );
  }

  String _getPhaseLabel(String phase) {
    switch (phase) {
      case 'countdown': return '倒计时';
      case 'betting': return '下注';
      case 'in_game': return '游戏中';
      case 'settlement': return '结算';
      default: return phase;
    }
  }
}


class ReplayData {
  final int roundId;
  final String roomName;
  final int roundNumber;
  final String betAmount;
  final String poolAmount;
  final String prizePerWinner;
  final String commitHash;
  final String revealSeed;
  final List<Participant> participants;
  final List<Winner> winners;
  final List<PhaseData> phases;

  ReplayData({
    required this.roundId,
    required this.roomName,
    required this.roundNumber,
    required this.betAmount,
    required this.poolAmount,
    required this.prizePerWinner,
    required this.commitHash,
    required this.revealSeed,
    required this.participants,
    required this.winners,
    required this.phases,
  });

  factory ReplayData.fromJson(Map<String, dynamic> json) {
    return ReplayData(
      roundId: json['round_id'] as int? ?? 0,
      roomName: json['room_name'] as String? ?? '',
      roundNumber: json['round_number'] as int? ?? 0,
      betAmount: json['bet_amount']?.toString() ?? '0',
      poolAmount: json['pool_amount']?.toString() ?? '0',
      prizePerWinner: json['prize_per_winner']?.toString() ?? '0',
      commitHash: json['commit_hash'] as String? ?? '',
      revealSeed: json['reveal_seed'] as String? ?? '',
      participants: (json['participants'] as List?)
          ?.map((e) => Participant.fromJson(e as Map<String, dynamic>))
          .toList() ?? [],
      winners: (json['winners'] as List?)
          ?.map((e) => Winner.fromJson(e as Map<String, dynamic>))
          .toList() ?? [],
      phases: (json['phases'] as List?)
          ?.map((e) => PhaseData.fromJson(e as Map<String, dynamic>))
          .toList() ?? [],
    );
  }
}

class Participant {
  final int userId;
  final String username;
  final bool isWinner;

  Participant({required this.userId, required this.username, required this.isWinner});

  factory Participant.fromJson(Map<String, dynamic> json) {
    return Participant(
      userId: json['user_id'] as int? ?? 0,
      username: json['username'] as String? ?? '',
      isWinner: json['is_winner'] as bool? ?? false,
    );
  }
}

class Winner {
  final int userId;
  final String username;

  Winner({required this.userId, required this.username});

  factory Winner.fromJson(Map<String, dynamic> json) {
    return Winner(
      userId: json['user_id'] as int? ?? 0,
      username: json['username'] as String? ?? '',
    );
  }
}

class PhaseData {
  final String phase;
  final int duration;

  PhaseData({required this.phase, required this.duration});

  factory PhaseData.fromJson(Map<String, dynamic> json) {
    return PhaseData(
      phase: json['phase'] as String? ?? '',
      duration: json['duration'] as int? ?? 5,
    );
  }
}

class VerificationResult {
  final int roundId;
  final String commitHash;
  final String revealSeed;
  final String computedHash;
  final bool hashMatch;
  final List<int> actualWinners;
  final List<int> computedWinners;
  final bool winnersMatch;
  final bool isValid;

  VerificationResult({
    required this.roundId,
    required this.commitHash,
    required this.revealSeed,
    required this.computedHash,
    required this.hashMatch,
    required this.actualWinners,
    required this.computedWinners,
    required this.winnersMatch,
    required this.isValid,
  });

  factory VerificationResult.fromJson(Map<String, dynamic> json) {
    return VerificationResult(
      roundId: json['round_id'] as int? ?? 0,
      commitHash: json['commit_hash'] as String? ?? '',
      revealSeed: json['reveal_seed'] as String? ?? '',
      computedHash: json['computed_hash'] as String? ?? '',
      hashMatch: json['hash_match'] as bool? ?? false,
      actualWinners: (json['actual_winners'] as List?)?.map((e) => e as int).toList() ?? [],
      computedWinners: (json['computed_winners'] as List?)?.map((e) => e as int).toList() ?? [],
      winnersMatch: json['winners_match'] as bool? ?? false,
      isValid: json['is_valid'] as bool? ?? false,
    );
  }
}
