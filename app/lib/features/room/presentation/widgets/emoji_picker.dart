import 'package:flutter/material.dart';
import 'package:flutter_screenutil/flutter_screenutil.dart';

// 预定义的12个表情
const Map<String, String> predefinedEmojis = {
  'happy': '😊',
  'sad': '😢',
  'angry': '😠',
  'surprised': '😮',
  'thumbs_up': '👍',
  'thumbs_down': '👎',
  'clap': '👏',
  'fire': '🔥',
  'heart': '❤️',
  'laugh': '😂',
  'cry': '😭',
  'cool': '😎',
};

class EmojiPicker extends StatelessWidget {
  final Function(String) onEmojiSelected;
  final bool enabled;

  const EmojiPicker({
    super.key,
    required this.onEmojiSelected,
    this.enabled = true,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: EdgeInsets.all(8.w),
      decoration: BoxDecoration(
        color: Theme.of(context).cardColor,
        borderRadius: BorderRadius.circular(12.r),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withAlpha(25),
            blurRadius: 8,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Wrap(
        spacing: 8.w,
        runSpacing: 8.h,
        children: predefinedEmojis.entries.map((entry) {
          return GestureDetector(
            onTap: enabled ? () => onEmojiSelected(entry.key) : null,
            child: Container(
              width: 40.w,
              height: 40.w,
              decoration: BoxDecoration(
                color: enabled
                    ? Colors.grey.withAlpha(25)
                    : Colors.grey.withAlpha(13),
                borderRadius: BorderRadius.circular(8.r),
              ),
              child: Center(
                child: Text(
                  entry.value,
                  style: TextStyle(
                    fontSize: 24.sp,
                    color: enabled ? null : Colors.grey,
                  ),
                ),
              ),
            ),
          );
        }).toList(),
      ),
    );
  }
}

class EmojiReactionOverlay extends StatefulWidget {
  final String emoji;
  final String username;
  final VoidCallback onComplete;

  const EmojiReactionOverlay({
    super.key,
    required this.emoji,
    required this.username,
    required this.onComplete,
  });

  @override
  State<EmojiReactionOverlay> createState() => _EmojiReactionOverlayState();
}

class _EmojiReactionOverlayState extends State<EmojiReactionOverlay>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;
  late Animation<double> _opacityAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      duration: const Duration(milliseconds: 3000),
      vsync: this,
    );

    _scaleAnimation = TweenSequence<double>([
      TweenSequenceItem(
        tween: Tween(begin: 0.0, end: 1.2).chain(CurveTween(curve: Curves.easeOut)),
        weight: 20,
      ),
      TweenSequenceItem(
        tween: Tween(begin: 1.2, end: 1.0).chain(CurveTween(curve: Curves.easeIn)),
        weight: 10,
      ),
      TweenSequenceItem(
        tween: ConstantTween(1.0),
        weight: 50,
      ),
      TweenSequenceItem(
        tween: Tween(begin: 1.0, end: 0.0).chain(CurveTween(curve: Curves.easeIn)),
        weight: 20,
      ),
    ]).animate(_controller);

    _opacityAnimation = TweenSequence<double>([
      TweenSequenceItem(
        tween: Tween(begin: 0.0, end: 1.0),
        weight: 20,
      ),
      TweenSequenceItem(
        tween: ConstantTween(1.0),
        weight: 60,
      ),
      TweenSequenceItem(
        tween: Tween(begin: 1.0, end: 0.0),
        weight: 20,
      ),
    ]).animate(_controller);

    _controller.forward().then((_) => widget.onComplete());
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final emojiChar = predefinedEmojis[widget.emoji] ?? '❓';

    return AnimatedBuilder(
      animation: _controller,
      builder: (context, child) {
        return Opacity(
          opacity: _opacityAnimation.value,
          child: Transform.scale(
            scale: _scaleAnimation.value,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  emojiChar,
                  style: TextStyle(fontSize: 48.sp),
                ),
                SizedBox(height: 4.h),
                Container(
                  padding: EdgeInsets.symmetric(
                    horizontal: 8.w,
                    vertical: 4.h,
                  ),
                  decoration: BoxDecoration(
                    color: Colors.black54,
                    borderRadius: BorderRadius.circular(8.r),
                  ),
                  child: Text(
                    widget.username,
                    style: TextStyle(
                      fontSize: 12.sp,
                      color: Colors.white,
                    ),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
