import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/models/users/update_profile_request.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/core/widgets/sophia_card.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';
import 'package:sophia_ai/features/users/presentation/cubit/user_profile_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/user_profile_state.dart';

class ProfileScreen extends StatelessWidget {
  const ProfileScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<UserProfileCubit>()..loadFromSession(),
      child: const _ProfileView(),
    );
  }
}

class _ProfileView extends StatefulWidget {
  const _ProfileView();

  @override
  State<_ProfileView> createState() => _ProfileViewState();
}

class _ProfileViewState extends State<_ProfileView> {
  late final TextEditingController _displayName;
  late final TextEditingController _preferredName;
  late final TextEditingController _timezone;
  late final TextEditingController _locale;
  late final TextEditingController _avatarUrl;
  var _initialized = false;

  @override
  void initState() {
    super.initState();
    _displayName = TextEditingController();
    _preferredName = TextEditingController();
    _timezone = TextEditingController();
    _locale = TextEditingController();
    _avatarUrl = TextEditingController();
  }

  void _hydrateFromSession() {
    if (_initialized) return;
    final session = context.read<SessionCubit>().state;
    if (session is SessionAuthenticated) {
      _displayName.text = session.profile.displayName;
      _preferredName.text = session.profile.preferredName;
      _timezone.text = session.profile.timezone;
      _locale.text = session.profile.locale;
      _avatarUrl.text = session.profile.avatarUrl ?? '';
      _initialized = true;
    }
  }

  @override
  void dispose() {
    _displayName.dispose();
    _preferredName.dispose();
    _timezone.dispose();
    _locale.dispose();
    _avatarUrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    _hydrateFromSession();

    return NeonWrapper(
      child: Scaffold(
        backgroundColor: context.colors.surface.withValues(alpha: 0),
        appBar: AppBar(
          title: const Text('Profile'),
          leading: TactileButton(
            semanticLabel: 'Volver a ajustes',
            onPressed: context.pop,
            child: const Padding(
              padding: EdgeInsets.all(SophiaSpace.sm),
              child: Icon(Icons.arrow_back),
            ),
          ),
        ),
        body: BlocConsumer<UserProfileCubit, UserProfileState>(
          listenWhen: (previous, current) =>
              previous is UserProfileLoaded &&
              previous.saving &&
              ((current is UserProfileLoaded && !current.saving) ||
                  current is UserProfileFailure),
          listener: (context, state) {
            if (state is UserProfileLoaded) {
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(const SnackBar(content: Text('Profile saved')));
            }
            if (state is UserProfileFailure) {
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(SnackBar(content: Text(state.message)));
            }
          },
          builder: (context, state) {
            final saving = state is UserProfileLoaded && state.saving;
            return ListView(
              padding: const EdgeInsets.all(24),
              children: [
                SophiaCard(
                  child: Column(
                    children: [
                      TextField(
                        key: const Key('profile_display_name'),
                        controller: _displayName,
                        decoration: const InputDecoration(
                          labelText: 'Display name',
                        ),
                      ),
                      TextField(
                        controller: _preferredName,
                        decoration: const InputDecoration(
                          labelText: 'Preferred name',
                        ),
                      ),
                      TextField(
                        controller: _timezone,
                        decoration: const InputDecoration(
                          labelText: 'Timezone',
                        ),
                      ),
                      TextField(
                        controller: _locale,
                        decoration: const InputDecoration(labelText: 'Locale'),
                      ),
                      TextField(
                        controller: _avatarUrl,
                        decoration: const InputDecoration(
                          labelText: 'Avatar URL',
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                NeonButton(
                  key: const Key('profile_save'),
                  onPressed: saving
                      ? null
                      : () {
                          context.read<UserProfileCubit>().update(
                            UpdateProfileRequest(
                              displayName: _displayName.text.trim(),
                              preferredName: _preferredName.text.trim(),
                              timezone: _timezone.text.trim(),
                              locale: _locale.text.trim(),
                              avatarUrl: _avatarUrl.text.trim().isEmpty
                                  ? null
                                  : _avatarUrl.text.trim(),
                            ),
                          );
                        },
                  child: saving
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Save profile'),
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}
