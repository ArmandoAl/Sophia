import 'dart:ui';
import 'package:flutter/material.dart';

class SophiaCard extends StatefulWidget {
  final Widget child;
  final EdgeInsetsGeometry? padding;
  final double? height;
  final Color? borderColor;
  final VoidCallback? onTap;
  final bool enableHover;

  const SophiaCard({
    super.key,
    required this.child,
    this.padding,
    this.height,
    this.borderColor,
    this.onTap,
    this.enableHover = true,
  });

  @override
  State<SophiaCard> createState() => _SophiaCardState();
}

class _SophiaCardState extends State<SophiaCard>
    with SingleTickerProviderStateMixin {
  bool _isHovered = false;
  late AnimationController _controller;
  late Animation<double> _scaleAnimation;
  late Animation<double> _glowAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 200),
    );
    _scaleAnimation = Tween<double>(
      begin: 1.0,
      end: 1.02,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOut));
    _glowAnimation = Tween<double>(
      begin: 0.05,
      end: 0.15,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOut));
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return MouseRegion(
      onEnter: widget.enableHover
          ? (_) {
              setState(() => _isHovered = true);
              _controller.forward();
            }
          : null,
      onExit: widget.enableHover
          ? (_) {
              setState(() => _isHovered = false);
              _controller.reverse();
            }
          : null,
      child: GestureDetector(
        onTap: widget.onTap,
        child: AnimatedBuilder(
          animation: _controller,
          builder: (context, child) {
            return Transform.scale(
              scale: widget.enableHover ? _scaleAnimation.value : 1.0,
              child: ClipRRect(
                borderRadius: BorderRadius.circular(16),
                child: BackdropFilter(
                  filter: ImageFilter.blur(
                    sigmaX: _isHovered ? 5 : 0,
                    sigmaY: _isHovered ? 5 : 0,
                  ),
                  child: Container(
                    height: widget.height,
                    padding: widget.padding ?? const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: theme.colorScheme.surface.withValues(
                        alpha: _isHovered ? 0.9 : 0.8,
                      ),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(
                        color:
                            widget.borderColor ??
                            (_isHovered
                                ? theme.primaryColor.withValues(
                                    alpha: _glowAnimation.value,
                                  )
                                : Colors.white.withValues(alpha: 0.05)),
                        width: 1,
                      ),
                      boxShadow: [
                        BoxShadow(
                          color: _isHovered
                              ? theme.primaryColor.withValues(
                                  alpha: _glowAnimation.value,
                                )
                              : Colors.black.withValues(alpha: 0.2),
                          blurRadius: _isHovered ? 20 : 10,
                          offset: const Offset(0, 4),
                        ),
                      ],
                    ),
                    child: widget.child,
                  ),
                ),
              ),
            );
          },
        ),
      ),
    );
  }
}
