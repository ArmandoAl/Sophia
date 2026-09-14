import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/models/enums/autonomy_level.dart';
import 'package:sophia_ai/core/models/enums/proactivity_level.dart';
import 'package:sophia_ai/core/models/users/update_ai_settings_request.dart';
import 'package:sophia_ai/core/models/users/update_profile_request.dart';
import 'package:sophia_ai/core/widgets/auth_scaffold.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';
import 'package:sophia_ai/features/users/presentation/cubit/onboarding_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/onboarding_state.dart';

class OnboardingFlowScreen extends StatelessWidget {
  const OnboardingFlowScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<OnboardingCubit>(),
      child: const _OnboardingView(),
    );
  }
}

class _OnboardingView extends StatefulWidget {
  const _OnboardingView();

  @override
  State<_OnboardingView> createState() => _OnboardingViewState();
}

class _OnboardingViewState extends State<_OnboardingView> {
  final _displayName = TextEditingController();
  final _preferredName = TextEditingController();
  final _timezone = TextEditingController(text: 'America/Tijuana');
  final _locale = TextEditingController(text: 'es-MX');
  final _avatarUrl = TextEditingController();
  final _assistantName = TextEditingController(text: 'Sofía');
  final _communicationStyle = TextEditingController(text: 'warm');
  final _preferredLanguage = TextEditingController(text: 'es');

  ProactivityLevel _proactivity = ProactivityLevel.medium;
  AutonomyLevel _autonomy = AutonomyLevel.suggestive;
  bool _memory = true;
  bool _reminders = true;
  bool _planning = true;
  int _step = 0;

  @override
  void initState() {
    super.initState();
    final session = sl<SessionCubit>().state;
    if (session is SessionAuthenticated) {
      _displayName.text = session.profile.displayName;
      _preferredName.text = session.profile.preferredName;
      _timezone.text = session.profile.timezone;
      _locale.text = session.profile.locale;
      _avatarUrl.text = session.profile.avatarUrl ?? '';
      _assistantName.text = session.aiSettings.assistantName;
      _communicationStyle.text = session.aiSettings.communicationStyle;
      _preferredLanguage.text = session.aiSettings.preferredLanguage;
      if (session.aiSettings.proactivityLevel.isKnown) {
        _proactivity = session.aiSettings.proactivityLevel;
      }
      if (session.aiSettings.autonomyLevel.isKnown) {
        _autonomy = session.aiSettings.autonomyLevel;
      }
      _memory = session.aiSettings.memoryEnabled;
      _reminders = session.aiSettings.remindersEnabled;
      _planning = session.aiSettings.planningEnabled;
    }
  }

  @override
  void dispose() {
    _displayName.dispose();
    _preferredName.dispose();
    _timezone.dispose();
    _locale.dispose();
    _avatarUrl.dispose();
    _assistantName.dispose();
    _communicationStyle.dispose();
    _preferredLanguage.dispose();
    super.dispose();
  }

  Future<void> _finish() async {
    await context.read<OnboardingCubit>().run(
      profile: UpdateProfileRequest(
        displayName: _displayName.text.trim(),
        preferredName: _preferredName.text.trim(),
        timezone: _timezone.text.trim(),
        locale: _locale.text.trim(),
        avatarUrl: _avatarUrl.text.trim().isEmpty
            ? null
            : _avatarUrl.text.trim(),
      ),
      aiSettings: UpdateAiSettingsRequest(
        assistantName: _assistantName.text.trim(),
        communicationStyle: _communicationStyle.text.trim(),
        proactivityLevel: _proactivity,
        autonomyLevel: _autonomy,
        memoryEnabled: _memory,
        remindersEnabled: _reminders,
        planningEnabled: _planning,
        preferredLanguage: _preferredLanguage.text.trim(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return BlocConsumer<OnboardingCubit, OnboardingState>(
      listener: (context, state) {
        // Redirect handled by GoRouter when session profile updates.
      },
      builder: (context, state) {
        final busy = state is OnboardingInProgress;
        final error = state is OnboardingFailure ? state.message : null;

        return AuthScaffold(
          title: _step == 0 ? 'Your profile' : 'Assistant preferences',
          subtitle: 'Quick setup — you can change this later',
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              if (_step == 0) ...[
                TextField(
                  key: const Key('onboarding_display_name'),
                  controller: _displayName,
                  decoration: const InputDecoration(labelText: 'Display name'),
                ),
                const SizedBox(height: SophiaSpace.md),
                TextField(
                  controller: _preferredName,
                  decoration: const InputDecoration(
                    labelText: 'Preferred name',
                  ),
                ),
                const SizedBox(height: SophiaSpace.md),
                TextField(
                  controller: _timezone,
                  decoration: const InputDecoration(labelText: 'Timezone'),
                ),
                const SizedBox(height: SophiaSpace.md),
                TextField(
                  controller: _locale,
                  decoration: const InputDecoration(labelText: 'Locale'),
                ),
                const SizedBox(height: SophiaSpace.md),
                TextField(
                  controller: _avatarUrl,
                  decoration: const InputDecoration(
                    labelText: 'Avatar URL (optional)',
                  ),
                ),
              ] else ...[
                TextField(
                  key: const Key('onboarding_assistant_name'),
                  controller: _assistantName,
                  decoration: const InputDecoration(
                    labelText: 'Assistant name',
                  ),
                ),
                const SizedBox(height: SophiaSpace.md),
                TextField(
                  controller: _communicationStyle,
                  decoration: const InputDecoration(
                    labelText: 'Communication style',
                  ),
                ),
                const SizedBox(height: SophiaSpace.md),
                TextField(
                  controller: _preferredLanguage,
                  decoration: const InputDecoration(
                    labelText: 'Preferred language',
                  ),
                ),
                const SizedBox(height: SophiaSpace.md),
                DropdownButtonFormField<ProactivityLevel>(
                  key: ValueKey('proactivity_${_proactivity.value}'),
                  initialValue: _proactivity,
                  decoration: const InputDecoration(labelText: 'Proactivity'),
                  items: ProactivityLevel.knownValues
                      .map(
                        (e) => DropdownMenuItem(value: e, child: Text(e.value)),
                      )
                      .toList(),
                  onChanged: busy
                      ? null
                      : (v) {
                          HapticFeedback.selectionClick();
                          setState(() => _proactivity = v!);
                        },
                ),
                const SizedBox(height: SophiaSpace.md),
                DropdownButtonFormField<AutonomyLevel>(
                  key: ValueKey('autonomy_${_autonomy.value}'),
                  initialValue: _autonomy,
                  isExpanded: true,
                  decoration: const InputDecoration(
                    labelText: 'Autonomy preference',
                  ),
                  items: AutonomyLevel.knownValues
                      .map(
                        (e) => DropdownMenuItem(
                          value: e,
                          child: Text(
                            e == AutonomyLevel.semiAutonomous
                                ? 'semi_autonomous (future / controlled)'
                                : e.value,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      )
                      .toList(),
                  onChanged: busy
                      ? null
                      : (v) {
                          HapticFeedback.selectionClick();
                          setState(() => _autonomy = v!);
                        },
                ),
                const SizedBox(height: SophiaSpace.xs),
                Text(
                  'Proposal-only: Sofia suggests; you confirm actions.',
                  style: Theme.of(context).textTheme.labelSmall,
                ),
                const SizedBox(height: SophiaSpace.sm),
                _SettingToggle(
                  label: 'Memory enabled',
                  value: _memory,
                  onChanged: busy ? null : (v) => setState(() => _memory = v),
                ),
                const SizedBox(height: SophiaSpace.xs),
                _SettingToggle(
                  label: 'Reminders enabled',
                  value: _reminders,
                  onChanged: busy
                      ? null
                      : (v) => setState(() => _reminders = v),
                ),
                const SizedBox(height: SophiaSpace.xs),
                _SettingToggle(
                  label: 'Planning enabled',
                  value: _planning,
                  onChanged: busy ? null : (v) => setState(() => _planning = v),
                ),
              ],
              if (error != null) ...[
                const SizedBox(height: 12),
                Text(
                  error,
                  key: const Key('onboarding_error'),
                  style: TextStyle(color: Theme.of(context).colorScheme.error),
                ),
              ],
              const SizedBox(height: 24),
              if (_step == 0)
                NeonButton(
                  key: const Key('onboarding_next'),
                  onPressed: busy ? null : () => setState(() => _step = 1),
                  child: const Text('Next'),
                )
              else
                NeonButton(
                  key: const Key('onboarding_finish'),
                  onPressed: busy ? null : _finish,
                  child: busy
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Finish setup'),
                ),
            ],
          ),
        );
      },
    );
  }
}

class _SettingToggle extends StatelessWidget {
  const _SettingToggle({
    required this.label,
    required this.value,
    required this.onChanged,
  });

  final String label;
  final bool value;
  final ValueChanged<bool>? onChanged;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
    child: Row(
      children: [
        Expanded(child: Text(label)),
        SophiaSwitch(value: value, onChanged: onChanged, semanticLabel: label),
      ],
    ),
  );
}
