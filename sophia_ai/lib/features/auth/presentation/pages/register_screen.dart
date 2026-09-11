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

/// Register screen.
///
/// **Policy:** after successful register, navigate to login (no auto-login).
/// Backend register does not return a JWT.
class RegisterScreen extends StatelessWidget {
  const RegisterScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AuthCubit>(),
      child: const _RegisterView(),
    );
  }
}

class _RegisterView extends StatefulWidget {
  const _RegisterView();

  @override
  State<_RegisterView> createState() => _RegisterViewState();
}

class _RegisterViewState extends State<_RegisterView> {
  final _name = TextEditingController();
  final _email = TextEditingController();
  final _password = TextEditingController();
  final _confirm = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  @override
  void dispose() {
    _name.dispose();
    _email.dispose();
    _password.dispose();
    _confirm.dispose();
    super.dispose();
  }

  String _friendlyError(AuthFailure failure) {
    final e = failure.exception;
    if (e is ApiException) {
      if (e.isConflict) return 'That email is already registered.';
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
        if (state is AuthRegisterSuccess) {
          ScaffoldMessenger.of(context).showSnackBar(
            const SnackBar(content: Text('Account created. Please sign in.')),
          );
          GoRouter.maybeOf(context)?.go('/login');
        }
      },
      builder: (context, state) {
        final loading = state is AuthLoading;
        final error = state is AuthFailure ? _friendlyError(state) : null;

        return AuthScaffold(
          title: 'Create account',
          subtitle: 'You will sign in after registering',
          footer: TactileButton(
            onPressed: loading
                ? null
                : () => GoRouter.maybeOf(context)?.go('/login'),
            child: const Padding(
              padding: EdgeInsets.all(SophiaSpace.sm),
              child: Text('Already have an account?'),
            ),
          ),
          child: Form(
            key: _formKey,
            child: Column(
              children: [
                TextFormField(
                  key: const Key('register_name'),
                  controller: _name,
                  decoration: const InputDecoration(labelText: 'Name'),
                  validator: (v) =>
                      (v == null || v.trim().isEmpty) ? 'Name required' : null,
                ),
                const SizedBox(height: 12),
                TextFormField(
                  key: const Key('register_email'),
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
                  key: const Key('register_password'),
                  controller: _password,
                  obscureText: true,
                  decoration: const InputDecoration(
                    labelText: 'Password (min 8)',
                  ),
                  validator: (v) {
                    if (v == null || v.length < 8) {
                      return 'Password must be at least 8 characters';
                    }
                    return null;
                  },
                ),
                const SizedBox(height: 12),
                TextFormField(
                  key: const Key('register_confirm'),
                  controller: _confirm,
                  obscureText: true,
                  decoration: const InputDecoration(
                    labelText: 'Confirm password',
                  ),
                  validator: (v) {
                    if (v != _password.text) return 'Passwords do not match';
                    return null;
                  },
                ),
                if (error != null) ...[
                  const SizedBox(height: 16),
                  Text(
                    error,
                    key: const Key('register_error'),
                    style: TextStyle(
                      color: Theme.of(context).colorScheme.error,
                    ),
                  ),
                ],
                const SizedBox(height: 24),
                NeonButton(
                  key: const Key('register_submit'),
                  onPressed: loading
                      ? null
                      : () {
                          if (!_formKey.currentState!.validate()) return;
                          context.read<AuthCubit>().register(
                            name: _name.text.trim(),
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
                      : const Text('Create account'),
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}
