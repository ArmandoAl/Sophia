import 'package:flutter/material.dart';

/// Botón con efecto neon y animaciones
class NeonButton extends StatefulWidget {
  final VoidCallback? onPressed;
  final Widget child;
  final Color? color;
  final bool isPrimary;
  final double? width;
  final double? height;
  final IconData? icon;

  const NeonButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.color,
    this.isPrimary = true,
    this.width,
    this.height,
    this.icon,
  });

  const NeonButton.icon({
    super.key,
    required this.onPressed,
    required this.child,
    required this.icon,
    this.color,
    this.isPrimary = true,
    this.width,
    this.height,
  });

  @override
  State<NeonButton> createState() => _NeonButtonState();
}

class _NeonButtonState extends State<NeonButton>
    with SingleTickerProviderStateMixin {
  bool _isPressed = false;
  bool _isHovered = false;
  late AnimationController _controller;
  late Animation<double> _glowAnimation;
  late Animation<double> _scaleAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 150),
    );
    _glowAnimation = Tween<double>(begin: 0.3, end: 0.6).animate(_controller);
    _scaleAnimation = Tween<double>(
      begin: 1.0,
      end: 0.95,
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
    final buttonColor = widget.color ?? theme.primaryColor;
    final isDisabled = widget.onPressed == null;

    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      child: GestureDetector(
        onTapDown: isDisabled
            ? null
            : (_) {
                setState(() => _isPressed = true);
                _controller.forward();
              },
        onTapUp: isDisabled
            ? null
            : (_) {
                setState(() => _isPressed = false);
                _controller.reverse();
              },
        onTapCancel: isDisabled
            ? null
            : () {
                setState(() => _isPressed = false);
                _controller.reverse();
              },
        onTap: widget.onPressed,
        child: AnimatedBuilder(
          animation: _controller,
          builder: (context, child) {
            return Transform.scale(
              scale: _isPressed ? _scaleAnimation.value : 1.0,
              child: AnimatedContainer(
                duration: const Duration(milliseconds: 200),
                width: widget.width,
                height: widget.height ?? 50,
                decoration: BoxDecoration(
                  gradient: widget.isPrimary
                      ? LinearGradient(
                          colors: [
                            buttonColor,
                            buttonColor.withValues(
                              red: (buttonColor.r * 0.8).clamp(0, 1),
                              green: (buttonColor.g * 0.8).clamp(0, 1),
                              blue: (buttonColor.b * 0.8).clamp(0, 1),
                            ),
                          ],
                          begin: Alignment.topLeft,
                          end: Alignment.bottomRight,
                        )
                      : null,
                  color: widget.isPrimary ? null : Colors.transparent,
                  borderRadius: BorderRadius.circular(12),
                  border: widget.isPrimary
                      ? null
                      : Border.all(
                          color: buttonColor.withValues(alpha: 0.5),
                          width: 2,
                        ),
                  boxShadow: widget.isPrimary && !isDisabled
                      ? [
                          BoxShadow(
                            color: buttonColor.withValues(
                              alpha: _isHovered || _isPressed
                                  ? _glowAnimation.value
                                  : 0.3,
                            ),
                            blurRadius: _isHovered || _isPressed ? 20 : 10,
                            spreadRadius: _isHovered || _isPressed ? 2 : 0,
                          ),
                        ]
                      : [],
                ),
                child: Material(
                  color: Colors.transparent,
                  child: InkWell(
                    borderRadius: BorderRadius.circular(12),
                    onTap: widget.onPressed,
                    child: Center(
                      child: widget.icon != null
                          ? Row(
                              mainAxisSize: MainAxisSize.min,
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  widget.icon,
                                  color: widget.isPrimary
                                      ? Colors.white
                                      : buttonColor,
                                  size: 20,
                                ),
                                const SizedBox(width: 8),
                                DefaultTextStyle(
                                  style: TextStyle(
                                    color: widget.isPrimary
                                        ? Colors.white
                                        : buttonColor,
                                    fontSize: 16,
                                    fontWeight: FontWeight.w600,
                                  ),
                                  child: widget.child,
                                ),
                              ],
                            )
                          : DefaultTextStyle(
                              style: TextStyle(
                                color: widget.isPrimary
                                    ? Colors.white
                                    : buttonColor,
                                fontSize: 16,
                                fontWeight: FontWeight.w600,
                              ),
                              child: widget.child,
                            ),
                    ),
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
