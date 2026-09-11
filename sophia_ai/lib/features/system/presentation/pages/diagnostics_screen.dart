import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/theme/motion.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../../beliefs/domain/models.dart';
import '../cubit/health_cubit.dart';
import '../cubit/health_state.dart';
import '../cubit/learning_diagnostics_cubit.dart';

class DiagnosticsScreen extends StatelessWidget {
  const DiagnosticsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    final learningAvailable = sl.isRegistered<LearningDiagnosticsCubit>();
    return MultiBlocProvider(
      providers: [
        BlocProvider(create: (_) => sl<HealthCubit>()..check()),
        if (learningAvailable)
          BlocProvider(create: (_) => sl<LearningDiagnosticsCubit>()..load()),
      ],
      child: _DiagnosticsView(learningAvailable: learningAvailable),
    );
  }
}

class _DiagnosticsView extends StatelessWidget {
  const _DiagnosticsView({required this.learningAvailable});

  final bool learningAvailable;

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Diagnóstico')),
    body: BlocBuilder<HealthCubit, HealthState>(
      builder: (context, state) {
        if (state is HealthLoading || state is HealthInitial) {
          return const MotionSwap(
            child: Padding(
              key: ValueKey('diagnostics-loading'),
              padding: EdgeInsets.all(SophiaSpace.lg),
              child: Column(
                children: [
                  ContentSkeleton(),
                  SizedBox(height: SophiaSpace.lg),
                  ContentSkeleton(),
                ],
              ),
            ),
          );
        }
        if (state is HealthFailure) {
          return MotionSwap(
            child: Center(
              key: const ValueKey('diagnostics-error'),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Text(
                    state.message,
                    key: const Key('diagnostics_error'),
                    textAlign: TextAlign.center,
                  ),
                  const SizedBox(height: SophiaSpace.md),
                  TactileButton(
                    semanticLabel: 'Reintentar diagnóstico',
                    onPressed: context.read<HealthCubit>().check,
                    child: const Padding(
                      padding: EdgeInsets.all(SophiaSpace.sm),
                      child: Text('Reintentar'),
                    ),
                  ),
                ],
              ),
            ),
          );
        }

        final health = state is HealthHealthy
            ? state.health
            : (state as HealthDegraded).health;
        final healthy = state is HealthHealthy;
        return MotionSwap(
          child: RefreshIndicator(
            key: const ValueKey('diagnostics-content'),
            onRefresh: () async {
              await context.read<HealthCubit>().check();
              if (learningAvailable && context.mounted) {
                await context.read<LearningDiagnosticsCubit>().load();
              }
            },
            child: Align(
              alignment: Alignment.topCenter,
              child: ConstrainedBox(
                constraints: const BoxConstraints(
                  maxWidth: SophiaSize.contentMaxWidth,
                ),
                child: ListView(
                  padding: const EdgeInsets.all(SophiaSpace.lg),
                  children: [
                    StaggeredEntry(
                      index: 0,
                      child: _Panel(
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Row(
                              children: [
                                Icon(
                                  Icons.circle,
                                  size: SophiaSpace.sm,
                                  color: healthy
                                      ? context.colors.positive
                                      : context.colors.attention,
                                ),
                                const SizedBox(width: SophiaSpace.xs),
                                Expanded(
                                  child: Text(
                                    healthy
                                        ? 'Todo funciona'
                                        : 'Servicio degradado',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.titleMedium,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: SophiaSpace.lg),
                            _row(
                              context,
                              'Estado',
                              health.status,
                              key: const Key('diagnostics_status'),
                            ),
                            _row(
                              context,
                              'Entorno',
                              health.environment,
                              key: const Key('diagnostics_environment'),
                            ),
                            _row(
                              context,
                              'Firestore',
                              health.firestore,
                              key: const Key('diagnostics_firestore'),
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(height: SophiaSpace.lg),
                    if (learningAvailable)
                      StaggeredEntry(
                        index: 1,
                        child:
                            BlocBuilder<
                              LearningDiagnosticsCubit,
                              LearningDiagnosticsState
                            >(
                              builder: (context, learning) {
                                if (learning.loading) {
                                  return const MotionSwap(
                                    child: ContentSkeleton(
                                      key: ValueKey('learning-loading'),
                                    ),
                                  );
                                }
                                if (learning.error != null) {
                                  return MotionSwap(
                                    child: _Panel(
                                      key: const ValueKey('learning-error'),
                                      child: Text(
                                        'Aprendizaje: ${learning.error}',
                                        style: TextStyle(
                                          color: context.colors.critical,
                                        ),
                                      ),
                                    ),
                                  );
                                }
                                return MotionSwap(
                                  child: _LearningPanel(
                                    key: const ValueKey('learning-content'),
                                    prompt: learning.prompt,
                                    summaries: learning.summaries,
                                  ),
                                );
                              },
                            ),
                      ),
                    const SizedBox(height: SophiaSpace.lg),
                    TactileButton(
                      semanticLabel: 'Actualizar diagnóstico',
                      onPressed: () {
                        context.read<HealthCubit>().check();
                        if (learningAvailable) {
                          context.read<LearningDiagnosticsCubit>().load();
                        }
                      },
                      child: Container(
                        key: const Key('diagnostics_refresh'),
                        padding: const EdgeInsets.all(SophiaSpace.md),
                        alignment: Alignment.center,
                        decoration: BoxDecoration(
                          color: context.colors.accent,
                          borderRadius: BorderRadius.circular(
                            SophiaRadius.control,
                          ),
                        ),
                        child: Text(
                          'Actualizar',
                          style: TextStyle(color: context.colors.surface),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    ),
  );

  Widget _row(BuildContext context, String label, String value, {Key? key}) =>
      Padding(
        padding: const EdgeInsets.only(bottom: SophiaSpace.sm),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Text(
                label,
                style: TextStyle(color: context.colors.softInk),
              ),
            ),
            const SizedBox(width: SophiaSpace.md),
            Flexible(
              child: Text(
                value,
                key: key,
                textAlign: TextAlign.end,
                overflow: TextOverflow.ellipsis,
                maxLines: 2,
              ),
            ),
          ],
        ),
      );
}

class _LearningPanel extends StatelessWidget {
  const _LearningPanel({
    super.key,
    required this.prompt,
    required this.summaries,
  });

  final PromptVersion? prompt;
  final List<LearningSummary> summaries;

  @override
  Widget build(BuildContext context) {
    final totalTokens = summaries.fold<int>(
      0,
      (total, item) => total + item.inputTokens + item.outputTokens,
    );
    return _Panel(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            'Estado del aprendizaje',
            style: Theme.of(context).textTheme.titleMedium,
          ),
          const SizedBox(height: SophiaSpace.xs),
          Text(
            prompt == null
                ? 'Sin prompt activo'
                : 'Prompt v${prompt!.version} · ${prompt!.tokenCount} tokens',
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
            style: SophiaType.dataLabel(
              context,
            ).copyWith(color: context.colors.softInk),
          ),
          const SizedBox(height: SophiaSpace.lg),
          Row(
            children: [
              Expanded(
                child: Text(
                  'Coste de los últimos 7 días',
                  style: Theme.of(context).textTheme.labelLarge,
                ),
              ),
              Text('$totalTokens tokens', style: SophiaType.dataLabel(context)),
            ],
          ),
          const SizedBox(height: SophiaSpace.md),
          LearningWeekChart(summaries: summaries),
        ],
      ),
    );
  }
}

class LearningWeekChart extends StatelessWidget {
  const LearningWeekChart({super.key, required this.summaries});

  final List<LearningSummary> summaries;

  @override
  Widget build(BuildContext context) {
    final days = _week(summaries, DateTime.now());
    final peak = days.fold<int>(1, (value, day) {
      final cost = day.cost;
      return cost > value ? cost : value;
    });
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        SizedBox(
          height: SophiaSize.chartHeight,
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              for (final day in days)
                Expanded(
                  child: Tooltip(
                    message: day.summary == null
                        ? '${day.key}: el worker no registró una ejecución'
                        : '${day.key}: ${day.cost} tokens; ${day.decisions} ${day.decisions == 1 ? 'decisión revisada' : 'decisiones revisadas'}',
                    child: Semantics(
                      label: day.summary == null
                          ? '${day.key}, sin ejecución'
                          : '${day.key}, worker ejecutado, coste ${day.cost} tokens',
                      child: Padding(
                        padding: const EdgeInsets.symmetric(
                          horizontal: SophiaSpace.xxs,
                        ),
                        child: Column(
                          children: [
                            Icon(
                              day.summary == null
                                  ? Icons.remove_circle_outline
                                  : Icons.check_circle_outline,
                              size: SophiaSpace.md,
                              color: day.summary == null
                                  ? context.colors.muted
                                  : context.colors.positive,
                            ),
                            const SizedBox(height: SophiaSpace.xs),
                            Expanded(
                              child: Align(
                                alignment: Alignment.bottomCenter,
                                child: TweenAnimationBuilder<double>(
                                  tween: Tween(
                                    begin: SophiaDataViz.minimumBarFraction,
                                    end: day.summary == null
                                        ? SophiaDataViz.minimumBarFraction
                                        : (day.cost / peak).clamp(
                                            SophiaDataViz.minimumBarFraction,
                                            1,
                                          ),
                                  ),
                                  duration: SophiaMotion.resolve(
                                    context,
                                    SophiaMotion.medium,
                                  ),
                                  curve: SophiaMotion.contentCurve,
                                  builder: (context, value, child) =>
                                      FractionallySizedBox(
                                        heightFactor: value,
                                        widthFactor: 1,
                                        child: child,
                                      ),
                                  child: DecoratedBox(
                                    decoration: BoxDecoration(
                                      color: day.summary == null
                                          ? context.colors.line
                                          : context.colors.accent,
                                      borderRadius: BorderRadius.circular(
                                        SophiaRadius.control,
                                      ),
                                    ),
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(height: SophiaSpace.xs),
                            Text(
                              day.day.day.toString().padLeft(2, '0'),
                              maxLines: 1,
                              overflow: TextOverflow.clip,
                              style: SophiaType.dataLabel(
                                context,
                              ).copyWith(color: context.colors.softInk),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
        const SizedBox(height: SophiaSpace.md),
        Wrap(
          spacing: SophiaSpace.md,
          runSpacing: SophiaSpace.xs,
          children: [
            _Legend(color: context.colors.positive, label: 'Worker ejecutado'),
            _Legend(color: context.colors.muted, label: 'Sin ejecución'),
          ],
        ),
      ],
    );
  }
}

class _Legend extends StatelessWidget {
  const _Legend({required this.color, required this.label});

  final Color color;
  final String label;

  @override
  Widget build(BuildContext context) => Row(
    mainAxisSize: MainAxisSize.min,
    children: [
      Icon(Icons.circle, size: SophiaSpace.xs, color: color),
      const SizedBox(width: SophiaSpace.xs),
      Text(label, style: Theme.of(context).textTheme.labelMedium),
    ],
  );
}

class _LearningDay {
  const _LearningDay(this.day, this.key, this.summary);

  final DateTime day;
  final String key;
  final LearningSummary? summary;

  int get cost => (summary?.inputTokens ?? 0) + (summary?.outputTokens ?? 0);
  int get decisions =>
      (summary?.approved ?? 0) +
      (summary?.corrected ?? 0) +
      (summary?.rejected ?? 0);
}

List<_LearningDay> _week(List<LearningSummary> summaries, DateTime now) {
  final byDate = {
    for (final item in summaries)
      if (item.date.length >= 10) item.date.substring(0, 10): item,
  };
  final today = DateUtils.dateOnly(now);
  return List.generate(7, (index) {
    final day = today.subtract(Duration(days: 6 - index));
    final month = day.month.toString().padLeft(2, '0');
    final date = day.day.toString().padLeft(2, '0');
    final key = '${day.year}-$month-$date';
    return _LearningDay(day, key, byDate[key]);
  });
}

class _Panel extends StatelessWidget {
  const _Panel({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(SophiaSpace.lg),
    decoration: BoxDecoration(
      color: context.colors.elevated,
      border: Border.all(color: context.colors.line),
      borderRadius: BorderRadius.circular(SophiaRadius.card),
    ),
    child: child,
  );
}
