import 'package:flutter/material.dart';

import '../theme/design_tokens.dart';
import 'motion/motion_widgets.dart';

class SophiaCard extends StatelessWidget {
  const SophiaCard({
    super.key,
    required this.child,
    this.padding,
    this.height,
    this.borderColor,
    this.onTap,
    this.enableHover = true,
  });

  final Widget child;
  final EdgeInsetsGeometry? padding;
  final double? height;
  final Color? borderColor;
  final VoidCallback? onTap;
  final bool enableHover;

  @override
  Widget build(BuildContext context) {
    final content = Container(
      height: height,
      padding: padding ?? const EdgeInsets.all(SophiaSpace.md),
      decoration: BoxDecoration(
        color: context.colors.elevated,
        borderRadius: BorderRadius.circular(SophiaRadius.card),
        border: Border.all(color: borderColor ?? context.colors.line),
      ),
      child: child,
    );
    return onTap == null
        ? content
        : TactileButton(onPressed: onTap, child: content);
  }
}
