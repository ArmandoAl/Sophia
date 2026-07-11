import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/widgets/sophia_card.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import '../cubit/settings_cubit.dart';
import '../cubit/settings_state.dart';

class SettingsPage extends StatelessWidget {
  const SettingsPage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) => SettingsCubit(),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          body: SafeArea(
          child: Padding(
            padding: const EdgeInsets.all(24.0),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  "Settings",
                  style: Theme.of(context).textTheme.headlineMedium?.copyWith(
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                ),
                const SizedBox(height: 8),
                Text(
                  "Customize your Sophia AI experience",
                  style: Theme.of(context).textTheme.bodyMedium?.copyWith(
                        color: Colors.grey,
                      ),
                ),
                const SizedBox(height: 32),
                Expanded(
                  child: SingleChildScrollView(
                    child: BlocBuilder<SettingsCubit, SettingsState>(
                      builder: (context, state) {
                        return Column(
                          children: [
                            // General Settings Section
                            _buildSectionTitle(context, "General"),
                            const SizedBox(height: 16),
                            SophiaCard(
                              child: Column(
                                children: [
                                  _buildSwitchTile(
                                    context,
                                    icon: Icons.notifications_outlined,
                                    title: "Notifications",
                                    subtitle: "Receive alerts and updates",
                                    value: state.notificationsEnabled,
                                    onChanged: (value) =>
                                        context.read<SettingsCubit>().toggleNotifications(value),
                                  ),
                                  const Divider(color: Colors.white10, height: 1),
                                  _buildSwitchTile(
                                    context,
                                    icon: Icons.dark_mode_outlined,
                                    title: "Dark Mode",
                                    subtitle: "Always enabled for optimal experience",
                                    value: state.darkModeEnabled,
                                    onChanged: null, // Disabled - siempre dark mode
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 32),

                            // Voice Assistant Section
                            _buildSectionTitle(context, "Voice Assistant"),
                            const SizedBox(height: 16),
                            SophiaCard(
                              child: Column(
                                children: [
                                  _buildSwitchTile(
                                    context,
                                    icon: Icons.mic_outlined,
                                    title: "Voice Assistant",
                                    subtitle: "Enable Sophia voice commands",
                                    value: state.voiceAssistantEnabled,
                                    onChanged: (value) =>
                                        context.read<SettingsCubit>().toggleVoiceAssistant(value),
                                  ),
                                  const Divider(color: Colors.white10, height: 1),
                                  Padding(
                                    padding: const EdgeInsets.all(16.0),
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Row(
                                          children: [
                                            Icon(
                                              Icons.volume_up_outlined,
                                              color: Theme.of(context).primaryColor,
                                              size: 20,
                                            ),
                                            const SizedBox(width: 12),
                                            Text(
                                              "Voice Volume",
                                              style: Theme.of(context).textTheme.bodyLarge,
                                            ),
                                          ],
                                        ),
                                        const SizedBox(height: 8),
                                        Slider(
                                          value: state.voiceVolume,
                                          onChanged: (value) =>
                                              context.read<SettingsCubit>().changeVoiceVolume(value),
                                          activeColor: Theme.of(context).primaryColor,
                                          inactiveColor: Colors.grey.shade800,
                                        ),
                                        Text(
                                          "${(state.voiceVolume * 100).toInt()}%",
                                          style: Theme.of(context).textTheme.bodySmall?.copyWith(
                                                color: Colors.grey,
                                              ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 32),

                            // Language & Region
                            _buildSectionTitle(context, "Language & Region"),
                            const SizedBox(height: 16),
                            SophiaCard(
                              child: _buildTile(
                                context,
                                icon: Icons.language_outlined,
                                title: "Language",
                                subtitle: state.language,
                                trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                                onTap: () => _showLanguageDialog(context),
                              ),
                            ),
                            const SizedBox(height: 32),

                            // About Section
                            _buildSectionTitle(context, "About"),
                            const SizedBox(height: 16),
                            SophiaCard(
                              child: Column(
                                children: [
                                  _buildTile(
                                    context,
                                    icon: Icons.info_outline,
                                    title: "Version",
                                    subtitle: "1.0.0",
                                  ),
                                  const Divider(color: Colors.white10, height: 1),
                                  _buildTile(
                                    context,
                                    icon: Icons.privacy_tip_outlined,
                                    title: "Privacy Policy",
                                    trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                                    onTap: () {},
                                  ),
                                  const Divider(color: Colors.white10, height: 1),
                                  _buildTile(
                                    context,
                                    icon: Icons.description_outlined,
                                    title: "Terms of Service",
                                    trailing: const Icon(Icons.chevron_right, color: Colors.grey),
                                    onTap: () {},
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 32),

                            // Logout Button
                            SizedBox(
                              width: double.infinity,
                              child: OutlinedButton.icon(
                                onPressed: () {
                                  // Implementar logout
                                },
                                icon: const Icon(Icons.logout, color: Colors.red),
                                label: const Text(
                                  "Sign Out",
                                  style: TextStyle(color: Colors.red),
                                ),
                                style: OutlinedButton.styleFrom(
                                  padding: const EdgeInsets.symmetric(vertical: 16),
                                  side: const BorderSide(color: Colors.red),
                                  shape: RoundedRectangleBorder(
                                    borderRadius: BorderRadius.circular(12),
                                  ),
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                          ],
                        );
                      },
                    ),
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

  Widget _buildSectionTitle(BuildContext context, String title) {
    return Text(
      title.toUpperCase(),
      style: Theme.of(context).textTheme.labelLarge?.copyWith(
            color: Theme.of(context).primaryColor,
            fontWeight: FontWeight.bold,
            letterSpacing: 1.2,
          ),
    );
  }

  Widget _buildSwitchTile(
    BuildContext context, {
    required IconData icon,
    required String title,
    String? subtitle,
    required bool value,
    required ValueChanged<bool>? onChanged,
  }) {
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: Theme.of(context).primaryColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: Theme.of(context).primaryColor, size: 20),
      ),
      title: Text(
        title,
        style: Theme.of(context).textTheme.bodyLarge,
      ),
      subtitle: subtitle != null
          ? Text(
              subtitle,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey,
                  ),
            )
          : null,
      trailing: Switch(
        value: value,
        onChanged: onChanged,
        activeColor: Theme.of(context).primaryColor,
      ),
    );
  }

  Widget _buildTile(
    BuildContext context, {
    required IconData icon,
    required String title,
    String? subtitle,
    Widget? trailing,
    VoidCallback? onTap,
  }) {
    return ListTile(
      contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      leading: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: Theme.of(context).primaryColor.withValues(alpha: 0.1),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Icon(icon, color: Theme.of(context).primaryColor, size: 20),
      ),
      title: Text(
        title,
        style: Theme.of(context).textTheme.bodyLarge,
      ),
      subtitle: subtitle != null
          ? Text(
              subtitle,
              style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: Colors.grey,
                  ),
            )
          : null,
      trailing: trailing,
      onTap: onTap,
    );
  }

  void _showLanguageDialog(BuildContext context) {
    showDialog(
      context: context,
      builder: (dialogContext) => AlertDialog(
        backgroundColor: const Color(0xFF151B24),
        title: const Text("Select Language"),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            _buildLanguageOption(context, dialogContext, "English"),
            _buildLanguageOption(context, dialogContext, "Español"),
            _buildLanguageOption(context, dialogContext, "Français"),
            _buildLanguageOption(context, dialogContext, "Deutsch"),
          ],
        ),
      ),
    );
  }

  Widget _buildLanguageOption(BuildContext context, BuildContext dialogContext, String language) {
    return ListTile(
      title: Text(language),
      onTap: () {
        context.read<SettingsCubit>().changeLanguage(language);
        Navigator.of(dialogContext).pop();
      },
    );
  }
}
