import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';

class SplashSessionScreen extends StatelessWidget {
  const SplashSessionScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: Scaffold(
        backgroundColor: Colors.transparent,
        body: BlocBuilder<SessionCubit, SessionState>(
          builder: (context, state) {
            if (state is SessionFailure) {
              return Center(
                child: Padding(
                  padding: const EdgeInsets.all(24),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        'Could not restore session',
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                      const SizedBox(height: 8),
                      Text(
                        state.message,
                        textAlign: TextAlign.center,
                        style: const TextStyle(color: Colors.grey),
                      ),
                      const SizedBox(height: 24),
                      NeonButton(
                        onPressed: () =>
                            context.read<SessionCubit>().bootstrap(),
                        child: const Text('Retry'),
                      ),
                      if (state.tokenRetained) ...[
                        const SizedBox(height: 12),
                        TextButton(
                          onPressed: () =>
                              context.read<SessionCubit>().logout(),
                          child: const Text('Sign out locally'),
                        ),
                      ],
                    ],
                  ),
                ),
              );
            }

            return const Center(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  CircularProgressIndicator(),
                  SizedBox(height: 16),
                  Text('Starting Sophia…'),
                ],
              ),
            );
          },
        ),
      ),
    );
  }
}
