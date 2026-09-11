import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';

class SplashSessionScreen extends StatelessWidget {
  const SplashSessionScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: Scaffold(
        backgroundColor: context.colors.surface.withValues(alpha: 0),
        body: BlocBuilder<SessionCubit, SessionState>(
          builder: (context, state) {
            if (state is SessionFailure) {
              return Center(
                child: Padding(
                  padding: const EdgeInsets.all(SophiaSpace.lg),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Text(
                        'Could not restore session',
                        style: Theme.of(context).textTheme.titleLarge,
                      ),
                      const SizedBox(height: SophiaSpace.xs),
                      Text(
                        state.message,
                        textAlign: TextAlign.center,
                        style: TextStyle(color: context.colors.softInk),
                      ),
                      const SizedBox(height: SophiaSpace.lg),
                      NeonButton(
                        onPressed: () =>
                            context.read<SessionCubit>().bootstrap(),
                        child: const Text('Retry'),
                      ),
                      if (state.tokenRetained) ...[
                        const SizedBox(height: SophiaSpace.sm),
                        TactileButton(
                          onPressed: context.read<SessionCubit>().logout,
                          child: const Padding(
                            padding: EdgeInsets.all(SophiaSpace.sm),
                            child: Text('Sign out locally'),
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              );
            }

            return const MotionSwap(
              child: Center(
                key: ValueKey('session-loading'),
                child: SizedBox(
                  width: SophiaSpace.xxxl * 2,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      ContentSkeleton(height: SophiaSpace.xs),
                      SizedBox(height: SophiaSpace.md),
                      Text('Starting Sophia…'),
                    ],
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
