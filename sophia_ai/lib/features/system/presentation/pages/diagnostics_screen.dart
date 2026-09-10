import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
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
                  ContentSkeleton(height: 140),
                  SizedBox(height: SophiaSpace.lg),
                  ContentSkeleton(height: 180),
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
                              size: 10,
                              color: healthy
                                  ? context.colors.positive
                                  : context.colors.attention,
                            ),
                            const SizedBox(width: SophiaSpace.xs),
                            Text(
                              healthy ? 'Todo funciona' : 'Servicio degradado',
                              style: Theme.of(context).textTheme.titleMedium,
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
                                  height: 230,
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
                      borderRadius: BorderRadius.circular(SophiaRadius.control),
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
        );
      },
    ),
  );

  Widget _row(BuildContext context, String label, String value, {Key? key}) =>
      Padding(
        padding: const EdgeInsets.only(bottom: SophiaSpace.sm),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [
            Text(label, style: TextStyle(color: context.colors.softInk)),
            Flexible(
              child: Text(value, key: key, textAlign: TextAlign.end),
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
    final peak = summaries.fold<int>(1, (max, item) {
      final total = item.approved + item.corrected + item.rejected;
      return total > max ? total : max;
    });
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
            style: TextStyle(color: context.colors.softInk),
          ),
          const SizedBox(height: SophiaSpace.lg),
          SizedBox(
            height: 112,
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.end,
              children: [
                for (final summary in summaries.reversed)
                  Expanded(
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                        horizontal: SophiaSpace.xxs,
                      ),
                      child: Tooltip(
                        message:
                            '${summary.date}: ${summary.approved} aprobadas, ${summary.corrected} corregidas, ${summary.rejected} rechazadas; ${summary.inputTokens + summary.outputTokens} tokens',
                        child: Column(
                          mainAxisAlignment: MainAxisAlignment.end,
                          children: [
                            Expanded(
                              child: Align(
                                alignment: Alignment.bottomCenter,
                                child: FractionallySizedBox(
                                  heightFactor:
                                      ((summary.approved +
                                                  summary.corrected +
                                                  summary.rejected) /
                                              peak)
                                          .clamp(.08, 1),
                                  child: DecoratedBox(
                                    decoration: BoxDecoration(
                                      color: context.colors.accent.withValues(
                                        alpha: .65,
                                      ),
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
                              summary.date.length >= 10
                                  ? summary.date.substring(8)
                                  : summary.date,
                              maxLines: 1,
                              style: Theme.of(context).textTheme.labelMedium
                                  ?.copyWith(color: context.colors.softInk),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }
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
