import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/theme/design_tokens.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../../../core/widgets/neon_wrapper.dart';
import '../../../../core/widgets/sophia_card.dart';
import '../cubit/insight_cubit.dart';
import '../cubit/insight_state.dart';

class WeeklyInsightPage extends StatelessWidget {
  const WeeklyInsightPage({super.key});

  @override
  Widget build(BuildContext context) => NeonWrapper(
    child: BlocProvider(
      create: (_) => InsightCubit(),
      child: Scaffold(
        backgroundColor: context.colors.surface.withValues(alpha: 0),
        appBar: AppBar(
          title: const Text('Weekly Insight'),
          actions: [
            TactileButton(
              semanticLabel: 'Más opciones',
              onPressed: () {},
              child: const Padding(
                padding: EdgeInsets.all(SophiaSpace.sm),
                child: Icon(Icons.more_vert),
              ),
            ),
            const SizedBox(width: SophiaSpace.xs),
          ],
        ),
        body: const _InsightBody(),
        bottomNavigationBar: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(SophiaSpace.lg),
            child: TactileButton(
              onPressed: () {},
              child: Container(
                padding: const EdgeInsets.all(SophiaSpace.md),
                alignment: Alignment.center,
                decoration: BoxDecoration(
                  color: context.colors.accent,
                  borderRadius: BorderRadius.circular(SophiaRadius.control),
                ),
                child: Text(
                  'Commit Learning',
                  style: TextStyle(color: context.colors.surface),
                ),
              ),
            ),
          ),
        ),
      ),
    ),
  );
}

class _InsightBody extends StatelessWidget {
  const _InsightBody();

  @override
  Widget build(BuildContext context) => SingleChildScrollView(
    padding: const EdgeInsets.all(SophiaSpace.lg),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        BlocBuilder<InsightCubit, InsightState>(
          builder: (context, state) => _Accuracy(percentage: state.accuracy),
        ),
        const SizedBox(height: SophiaSpace.md),
        Text(
          'Sophia is learning and improving every day thanks to your feedback.',
          style: TextStyle(color: context.colors.softInk),
        ),
        const SizedBox(height: SophiaSpace.xl),
        Text(
          'Decisions Needing Review',
          style: Theme.of(context).textTheme.titleMedium,
        ),
        const SizedBox(height: SophiaSpace.md),
        BlocBuilder<InsightCubit, InsightState>(
          builder: (context, state) => Column(
            children: [
              for (final decision in state.pendingDecisions)
                Padding(
                  padding: const EdgeInsets.only(bottom: SophiaSpace.sm),
                  child: _DecisionCard(item: decision),
                ),
            ],
          ),
        ),
      ],
    ),
  );
}

class _Accuracy extends StatelessWidget {
  const _Accuracy({required this.percentage});

  final double percentage;

  @override
  Widget build(BuildContext context) => SophiaCard(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            const Expanded(child: Text('Learning confidence')),
            Text(
              '${(percentage * 100).toInt()}%',
              style: SophiaType.dataLabel(context),
            ),
          ],
        ),
        const SizedBox(height: SophiaSpace.sm),
        LinearProgressIndicator(
          value: percentage,
          color: context.colors.accent,
          backgroundColor: context.colors.line,
          borderRadius: BorderRadius.circular(SophiaRadius.control),
        ),
      ],
    ),
  );
}

class _DecisionCard extends StatelessWidget {
  const _DecisionCard({required this.item});

  final DecisionItem item;

  @override
  Widget build(BuildContext context) => SophiaCard(
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Container(
              padding: const EdgeInsets.all(SophiaSpace.xs),
              decoration: BoxDecoration(
                color: context.colors.line,
                shape: BoxShape.circle,
              ),
              child: Text(item.iconEmoji),
            ),
            const SizedBox(width: SophiaSpace.sm),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.title,
                    style: Theme.of(context).textTheme.labelLarge,
                  ),
                  Text(
                    item.timestamp,
                    style: Theme.of(context).textTheme.labelSmall,
                  ),
                ],
              ),
            ),
          ],
        ),
        const SizedBox(height: SophiaSpace.md),
        Text(item.reasoning),
        const SizedBox(height: SophiaSpace.lg),
        Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            _Action(
              icon: Icons.close,
              label: 'Reject',
              color: context.colors.critical,
            ),
            _Action(
              icon: Icons.edit_note,
              label: 'Add Note',
              color: context.colors.softInk,
            ),
            _Action(
              icon: Icons.check,
              label: 'Approve',
              color: context.colors.positive,
            ),
          ],
        ),
      ],
    ),
  );
}

class _Action extends StatelessWidget {
  const _Action({required this.icon, required this.label, required this.color});

  final IconData icon;
  final String label;
  final Color color;

  @override
  Widget build(BuildContext context) => TactileButton(
    semanticLabel: label,
    onPressed: () {},
    child: Padding(
      padding: const EdgeInsets.all(SophiaSpace.xs),
      child: Column(
        children: [
          Icon(icon, color: color),
          const SizedBox(height: SophiaSpace.xxs),
          Text(label, style: TextStyle(color: color)),
        ],
      ),
    ),
  );
}
