import 'package:flutter/material.dart';

import '../theme/design_tokens.dart';
import 'motion/motion_widgets.dart';

/// Legacy name kept while every call site migrates to the shared tactile control.
class NeonButton extends StatelessWidget {
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

  final VoidCallback? onPressed;
  final Widget child;
  final Color? color;
  final bool isPrimary;
  final double? width;
  final double? height;
  final IconData? icon;

  @override
  Widget build(BuildContext context) {
    final buttonColor = color ?? context.colors.accent;
    final foreground = isPrimary ? context.colors.surface : buttonColor;
    return TactileButton(
      onPressed: onPressed,
      child: Opacity(
        opacity: onPressed == null ? SophiaOpacity.quiet : 1,
        child: Container(
          width: width,
          height: height ?? SophiaSpace.xxl,
          padding: const EdgeInsets.symmetric(horizontal: SophiaSpace.md),
          alignment: Alignment.center,
          decoration: BoxDecoration(
            color: isPrimary
                ? buttonColor
                : context.colors.surface.withValues(alpha: 0),
            border: Border.all(color: buttonColor),
            borderRadius: BorderRadius.circular(SophiaRadius.control),
          ),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              if (icon != null) ...[
                Icon(icon, color: foreground, size: SophiaSpace.lg),
                const SizedBox(width: SophiaSpace.xs),
              ],
              Flexible(
                child: DefaultTextStyle(
                  style: Theme.of(
                    context,
                  ).textTheme.labelLarge!.copyWith(color: foreground),
                  child: child,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
