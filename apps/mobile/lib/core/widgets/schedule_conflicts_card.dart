import 'package:flutter/material.dart';

import '../theme/design_tokens.dart';
import 'motion/motion_widgets.dart';

class ScheduleConflictCard extends StatelessWidget {
  const ScheduleConflictCard({super.key});

  @override
  Widget build(BuildContext context) => Container(
    width: 280,
    padding: const EdgeInsets.all(SophiaSpace.md),
    decoration: BoxDecoration(
      color: context.colors.elevated,
      borderRadius: BorderRadius.circular(SophiaRadius.card),
      border: Border.all(color: context.colors.line),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(
              Icons.warning_amber_rounded,
              color: context.colors.attention,
              size: SophiaSpace.lg,
            ),
            const SizedBox(width: SophiaSpace.xs),
            Text(
              'Conflict Detected',
              style: Theme.of(context).textTheme.labelLarge,
            ),
          ],
        ),
        const SizedBox(height: SophiaSpace.sm),
        Container(
          height: SophiaSpace.xxl,
          decoration: BoxDecoration(
            color: context.colors.surface,
            borderRadius: BorderRadius.circular(SophiaRadius.control),
          ),
          child: Row(
            children: [
              const Spacer(),
              _Block(label: 'Deep Work', color: context.colors.accent),
              _Block(label: 'Team Sync', color: context.colors.attention),
              const Spacer(),
            ],
          ),
        ),
        const SizedBox(height: SophiaSpace.xs),
        Text(
          'You have Deep Work scheduled. I can move it to 4:30 PM.',
          style: Theme.of(context).textTheme.bodySmall,
        ),
        const SizedBox(height: SophiaSpace.sm),
        Row(
          children: [
            Expanded(
              child: TactileButton(
                onPressed: () {},
                child: _Action(label: 'Reschedule', filled: true),
              ),
            ),
            const SizedBox(width: SophiaSpace.xs),
            Expanded(
              child: TactileButton(
                onPressed: () {},
                child: _Action(label: 'Cancel', filled: false),
              ),
            ),
          ],
        ),
      ],
    ),
  );
}

class _Block extends StatelessWidget {
  const _Block({required this.label, required this.color});

  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) => Expanded(
    flex: 2,
    child: Container(
      margin: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
      alignment: Alignment.center,
      decoration: BoxDecoration(
        color: color.withValues(alpha: SophiaOpacity.quiet),
        border: Border.all(color: color),
        borderRadius: BorderRadius.circular(SophiaRadius.control),
      ),
      child: Text(label, style: Theme.of(context).textTheme.labelSmall),
    ),
  );
}

class _Action extends StatelessWidget {
  const _Action({required this.label, required this.filled});

  final String label;
  final bool filled;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
    alignment: Alignment.center,
    decoration: BoxDecoration(
      color: filled
          ? context.colors.accent
          : context.colors.surface.withValues(alpha: 0),
      border: Border.all(
        color: filled ? context.colors.accent : context.colors.line,
      ),
      borderRadius: BorderRadius.circular(SophiaRadius.control),
    ),
    child: Text(
      label,
      style: Theme.of(context).textTheme.labelMedium?.copyWith(
        color: filled ? context.colors.surface : context.colors.ink,
      ),
    ),
  );
}
