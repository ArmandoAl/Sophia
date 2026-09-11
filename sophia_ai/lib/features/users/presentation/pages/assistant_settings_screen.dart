import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:go_router/go_router.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/core/models/enums/autonomy_level.dart';
import 'package:sophia_ai/core/models/enums/proactivity_level.dart';
import 'package:sophia_ai/core/models/users/update_ai_settings_request.dart';
import 'package:sophia_ai/core/widgets/neon_button.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/core/widgets/sophia_card.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_cubit.dart';
import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';
import 'package:sophia_ai/features/users/presentation/cubit/ai_settings_cubit.dart';
import 'package:sophia_ai/features/users/presentation/cubit/ai_settings_state.dart';

class AssistantSettingsScreen extends StatelessWidget {
  const AssistantSettingsScreen({super.key});

  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (_) => sl<AiSettingsCubit>()..loadFromSession(),
      child: const _AssistantSettingsView(),
    );
  }
}

class _AssistantSettingsView extends StatefulWidget {
  const _AssistantSettingsView();

  @override
  State<_AssistantSettingsView> createState() => _AssistantSettingsViewState();
}

class _AssistantSettingsViewState extends State<_AssistantSettingsView> {
  late final TextEditingController _assistantName;
  late final TextEditingController _communicationStyle;
  late final TextEditingController _preferredLanguage;
  ProactivityLevel _proactivity = ProactivityLevel.medium;
  AutonomyLevel _autonomy = AutonomyLevel.suggestive;
  bool _memory = true;
  bool _reminders = true;
  bool _planning = true;
  var _initialized = false;

  @override
  void initState() {
    super.initState();
    _assistantName = TextEditingController();
    _communicationStyle = TextEditingController();
    _preferredLanguage = TextEditingController();
  }

  void _hydrate() {
    if (_initialized) return;
    final session = context.read<SessionCubit>().state;
    if (session is SessionAuthenticated) {
      final s = session.aiSettings;
      _assistantName.text = s.assistantName;
      _communicationStyle.text = s.communicationStyle;
      _preferredLanguage.text = s.preferredLanguage;
      if (s.proactivityLevel.isKnown) _proactivity = s.proactivityLevel;
      if (s.autonomyLevel.isKnown) _autonomy = s.autonomyLevel;
      _memory = s.memoryEnabled;
      _reminders = s.remindersEnabled;
      _planning = s.planningEnabled;
      _initialized = true;
    }
  }

  @override
  void dispose() {
    _assistantName.dispose();
    _communicationStyle.dispose();
    _preferredLanguage.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    _hydrate();

    return NeonWrapper(
      child: Scaffold(
        backgroundColor: context.colors.surface.withValues(alpha: 0),
        appBar: AppBar(
          title: const Text('Assistant settings'),
          leading: TactileButton(
            semanticLabel: 'Volver a ajustes',
            onPressed: context.pop,
            child: const Padding(
              padding: EdgeInsets.all(SophiaSpace.sm),
              child: Icon(Icons.arrow_back),
            ),
          ),
        ),
        body: BlocConsumer<AiSettingsCubit, AiSettingsState>(
          listenWhen: (previous, current) =>
              previous is AiSettingsLoaded &&
              previous.saving &&
              ((current is AiSettingsLoaded && !current.saving) ||
                  current is AiSettingsFailure),
          listener: (context, state) {
            if (state is AiSettingsLoaded) {
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(const SnackBar(content: Text('Settings saved')));
            }
            if (state is AiSettingsFailure) {
              ScaffoldMessenger.of(
                context,
              ).showSnackBar(SnackBar(content: Text(state.message)));
            }
          },
          builder: (context, state) {
            final saving = state is AiSettingsLoaded && state.saving;
            return ListView(
              padding: const EdgeInsets.all(24),
              children: [
                SophiaCard(
                  child: Column(
                    children: [
                      TextField(
                        key: const Key('ai_assistant_name'),
                        controller: _assistantName,
                        decoration: const InputDecoration(
                          labelText: 'Assistant name',
                        ),
                      ),
                      TextField(
                        controller: _communicationStyle,
                        decoration: const InputDecoration(
                          labelText: 'Communication style',
                        ),
                      ),
                      TextField(
                        controller: _preferredLanguage,
                        decoration: const InputDecoration(
                          labelText: 'Preferred language',
                        ),
                      ),
                      DropdownButtonFormField<ProactivityLevel>(
                        isExpanded: true,
                        key: ValueKey('proactivity_${_proactivity.value}'),
                        initialValue: _proactivity,
                        decoration: const InputDecoration(
                          labelText: 'Proactivity',
                        ),
                        items: ProactivityLevel.knownValues
                            .map(
                              (e) => DropdownMenuItem(
                                value: e,
                                child: Text(e.value),
                              ),
                            )
                            .toList(),
                        onChanged: saving
                            ? null
                            : (v) {
                                HapticFeedback.selectionClick();
                                setState(() => _proactivity = v!);
                              },
                      ),
                      DropdownButtonFormField<AutonomyLevel>(
                        isExpanded: true,
                        key: ValueKey('autonomy_${_autonomy.value}'),
                        initialValue: _autonomy,
                        decoration: const InputDecoration(
                          labelText: 'Autonomy preference',
                          helperText:
                              'Backend remains proposal-only. '
                              'semi_autonomous is a preference, not live auto-execution.',
                        ),
                        items: AutonomyLevel.knownValues
                            .map(
                              (e) => DropdownMenuItem(
                                value: e,
                                child: Text(
                                  e == AutonomyLevel.semiAutonomous
                                      ? 'semi_autonomous (controlled / future)'
                                      : e.value,
                                ),
                              ),
                            )
                            .toList(),
                        onChanged: saving
                            ? null
                            : (v) {
                                HapticFeedback.selectionClick();
                                setState(() => _autonomy = v!);
                              },
                      ),
                      _SettingToggle(
                        controlKey: const Key('ai_memory_switch'),
                        label: 'Memory enabled',
                        value: _memory,
                        onChanged: saving
                            ? null
                            : (v) => setState(() => _memory = v),
                      ),
                      _SettingToggle(
                        label: 'Reminders enabled',
                        value: _reminders,
                        onChanged: saving
                            ? null
                            : (v) => setState(() => _reminders = v),
                      ),
                      _SettingToggle(
                        label: 'Planning enabled',
                        value: _planning,
                        onChanged: saving
                            ? null
                            : (v) => setState(() => _planning = v),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 24),
                NeonButton(
                  key: const Key('ai_settings_save'),
                  onPressed: saving
                      ? null
                      : () {
                          context.read<AiSettingsCubit>().update(
                            UpdateAiSettingsRequest(
                              assistantName: _assistantName.text.trim(),
                              communicationStyle: _communicationStyle.text
                                  .trim(),
                              preferredLanguage: _preferredLanguage.text.trim(),
                              proactivityLevel: _proactivity,
                              autonomyLevel: _autonomy,
                              memoryEnabled: _memory,
                              remindersEnabled: _reminders,
                              planningEnabled: _planning,
                            ),
                          );
                        },
                  child: saving
                      ? const SizedBox(
                          height: 20,
                          width: 20,
                          child: CircularProgressIndicator(strokeWidth: 2),
                        )
                      : const Text('Save settings'),
                ),
              ],
            );
          },
        ),
      ),
    );
  }
}

class _SettingToggle extends StatelessWidget {
  const _SettingToggle({
    this.controlKey,
    required this.label,
    required this.value,
    required this.onChanged,
  });

  final String label;
  final bool value;
  final ValueChanged<bool>? onChanged;
  final Key? controlKey;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
    child: Row(
      children: [
        Expanded(child: Text(label)),
        SophiaSwitch(
          key: controlKey,
          value: value,
          onChanged: onChanged,
          semanticLabel: label,
        ),
      ],
    ),
  );
}
