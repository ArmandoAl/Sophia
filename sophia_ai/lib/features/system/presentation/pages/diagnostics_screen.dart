import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/core/widgets/sophia_card.dart';
import 'package:sophia_ai/features/system/presentation/cubit/health_cubit.dart';
import 'package:sophia_ai/features/system/presentation/cubit/health_state.dart';

/// Backend diagnostics from `GET /health` only (no CPU/RAM mock).
class DiagnosticsScreen extends StatelessWidget {
  const DiagnosticsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) => sl<HealthCubit>()..check(),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          body: SafeArea(
            child: Padding(
              padding: const EdgeInsets.all(24),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Sophia System\nDiagnostics',
                    style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  const SizedBox(height: 8),
                  Text(
                    'Live backend health contract',
                    style: Theme.of(
                      context,
                    ).textTheme.bodyMedium?.copyWith(color: Colors.grey),
                  ),
                  const SizedBox(height: 24),
                  Expanded(
                    child: BlocBuilder<HealthCubit, HealthState>(
                      builder: (context, state) {
                        if (state is HealthLoading || state is HealthInitial) {
                          return const Center(
                            child: CircularProgressIndicator(),
                          );
                        }
                        if (state is HealthFailure) {
                          return Center(
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                Text(
                                  state.message,
                                  key: const Key('diagnostics_error'),
                                  textAlign: TextAlign.center,
                                ),
                                const SizedBox(height: 16),
                                NeonButton(
                                  onPressed: () =>
                                      context.read<HealthCubit>().check(),
                                  child: const Text('Retry'),
                                ),
                              ],
                            ),
                          );
                        }

                        final health = state is HealthHealthy
                            ? state.health
                            : (state as HealthDegraded).health;
                        final statusColor = state is HealthDegraded
                            ? Colors.orangeAccent
                            : Colors.tealAccent;

                        return ListView(
                          children: [
                            SophiaCard(
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Text(
                                    'Backend health',
                                    style: Theme.of(
                                      context,
                                    ).textTheme.titleLarge,
                                  ),
                                  const SizedBox(height: 20),
                                  _row('Status', health.status, statusColor),
                                  const SizedBox(height: 12),
                                  _row(
                                    'Environment',
                                    health.environment,
                                    Colors.white70,
                                  ),
                                  const SizedBox(height: 12),
                                  _row(
                                    'Firestore',
                                    health.firestore,
                                    Colors.white70,
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 16),
                            NeonButton(
                              key: const Key('diagnostics_refresh'),
                              onPressed: () =>
                                  context.read<HealthCubit>().check(),
                              child: const Text('Refresh'),
                            ),
                          ],
                        );
                      },
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  Widget _row(String label, String value, Color color) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.spaceBetween,
      children: [
        Text(label, style: const TextStyle(color: Colors.grey)),
        Text(
          value,
          key: Key('diagnostics_$label'.toLowerCase()),
          style: TextStyle(color: color, fontWeight: FontWeight.w600),
        ),
      ],
    );
  }
}
