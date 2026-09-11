import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/network/api_exception.dart';
import 'package:sophia_ai/core/widgets/auth_scaffold.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/features/auth/presentation/cubit/auth_cubit.dart';
import 'package:sophia_ai/features/auth/presentation/cubit/auth_state.dart';

class LoginScreen extends StatelessWidget {
  const LoginScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AuthCubit>(),
      child: const _LoginView(),
    );
  }
}

class _LoginView extends StatefulWidget {
  const _LoginView();

  @override
  State<_LoginView> createState() => _LoginViewState();
}

class _LoginViewState extends State<_LoginView> {
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _email.dispose();
    _password.dispose();
    super.dispose();
  }

  String _friendlyError(AuthFailure failure) {
    final e = failure.exception;
    if (e is ApiException) {
      if (e.isUnauthorized) return 'Invalid email or password.';
      if (e.isRateLimited) {
        return 'Too many attempts. Please wait and try again.';
      }
      if (e.isBadRequest) return failure.message;
    }
    return failure.message;
  }

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<AuthCubit, AuthState>(
      listener: (context, state) {
        // Navigation is driven by SessionCubit + GoRouter redirect.
      },
      builder: (context, state) {
        final loading = state is AuthLoading;
        final error = state is AuthFailure ? _friendlyError(state) : null;

        return AuthScaffold(
          title: 'Sign in',
          subtitle: 'Use your Sofia account',
          footer: TactileButton(
            onPressed: loading
                ? null
                : () => GoRouter.maybeOf(context)?.go('/register'),
            child: const Padding(
              padding: EdgeInsets.all(SophiaSpace.sm),
              child: Text('Create an account'),
            ),
          ),
          child: Form(
            key: _formKey,
            child: Column(
              children: [
                TextFormField(
                  key: const Key('login_email'),
                  controller: _email,
                  keyboardType: TextInputType.emailAddress,
                  decoration: const InputDecoration(labelText: 'Email'),
                  validator: (v) {
                    if (v == null || v.trim().isEmpty) return 'Email required';
                    if (!v.contains('@')) return 'Enter a valid email';
                    return null;
                  },
                ),
                const SizedBox(height: 12),
                TextFormField(
                  key: const Key('login_password'),
                  controller: _password,
                  obscureText: true,
                  decoration: const InputDecoration(labelText: 'Password'),
                  validator: (v) {
                    if (v == null || v.isEmpty) return 'Password required';
                    return null;
                  },
                ),
                if (error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    error,
                    key: const Key('login_error'),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ],
                const SizedBox(height: 24),
                NeonButton(
                  key: const Key('login_submit'),
                  onPressed: loading
                      ? null
                      : () {
                          if (!_formKey.currentState!.validate()) return;
                          context.read<AuthCubit>().login(
                            email: _email.text.trim(),
                            password: _password.text,
                          );
                        },
                  child: loading
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Sign in'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
