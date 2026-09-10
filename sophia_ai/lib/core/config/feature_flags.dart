/// Local feature flags for future Sofia phases.
///
/// Defaults are conservative: incomplete or risky backend features stay off.
class FeatureFlags {
  const FeatureFlags({
    this.aiRuntimeEnabled = false,
    this.aiActionExecutionEnabled = true,
    this.notificationsEnabled = false,
    this.smartHomeEnabled = false,
  });

  /// F4: AI runtime chat. Keep off until dry-run client wiring is ready.
  final bool aiRuntimeEnabled;

  /// The backend supports real confirm/reject/execute operations on
  /// ai_action_proposals, so execution is enabled for the beta.
  final bool aiActionExecutionEnabled;

  /// F5: device token / FCM registration UI.
  final bool notificationsEnabled;

  /// Out of Backend v0.1 scope — keep UI gated.
  final bool smartHomeEnabled;

  static const FeatureFlags defaults = FeatureFlags();

  FeatureFlags copyWith({
    bool? aiRuntimeEnabled,
    bool? aiActionExecutionEnabled,
    bool? notificationsEnabled,
    bool? smartHomeEnabled,
  }) {
    return FeatureFlags(
      aiRuntimeEnabled: aiRuntimeEnabled ?? this.aiRuntimeEnabled,
      aiActionExecutionEnabled:
          aiActionExecutionEnabled ?? this.aiActionExecutionEnabled,
      notificationsEnabled: notificationsEnabled ?? this.notificationsEnabled,
      smartHomeEnabled: smartHomeEnabled ?? this.smartHomeEnabled,
    );
  }
}
