import 'package:flutter/material.dart';

import '../theme/design_tokens.dart';

class CircularLightControl extends StatelessWidget {
  const CircularLightControl({
    super.key,
    required this.percentage,
    required this.label,
    required this.subLabel,
  });

  final double percentage;
  final String label;
  final String subLabel;

  @override
  Widget build(BuildContext context) => SizedBox.square(
    dimension: 300,
    child: Stack(
      alignment: Alignment.center,
      children: [
        SizedBox.square(
          dimension: 260,
          child: CircularProgressIndicator(
            value: (percentage / 100).clamp(0, 1),
            strokeWidth: SophiaSpace.xl,
            strokeCap: StrokeCap.round,
            color: context.colors.accent,
            backgroundColor: context.colors.line,
          ),
        ),
        Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(label, style: Theme.of(context).textTheme.bodyMedium),
            const SizedBox(height: SophiaSpace.xs),
            Text('${percentage.toInt()}%', style: SophiaType.data(context)),
            const SizedBox(height: SophiaSpace.xs),
            Text(
              subLabel,
              style: Theme.of(
                context,
              ).textTheme.labelLarge?.copyWith(color: context.colors.accent),
            ),
          ],
        ),
      ],
    ),
  );
}
