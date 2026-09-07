/// Wire value for `autonomy_level`.
///
/// Known values: `manual`, `suggestive`, `semi_autonomous`.
/// Unknown backend values parse without throwing and keep [value] as raw.
class AutonomyLevel {
  const AutonomyLevel._(this.value, {required this.isKnown});

  final String value;
  final bool isKnown;

  static const manual = AutonomyLevel._('manual', isKnown: true);
  static const suggestive = AutonomyLevel._('suggestive', isKnown: true);
  static const semiAutonomous = AutonomyLevel._(
    'semi_autonomous',
    isKnown: true,
  );

  static const List<AutonomyLevel> knownValues = [
    manual,
    suggestive,
    semiAutonomous,
  ];

  factory AutonomyLevel.fromJson(String raw) {
    for (final level in knownValues) {
      if (level.value == raw) return level;
    }
    return AutonomyLevel._(raw, isKnown: false);
  }

  /// Encodes only known values for requests.
  String toJson() {
    if (!isKnown) {
      throw StateError(
        'Cannot encode unknown autonomy_level "$value" in requests',
      );
    }
    return value;
  }

  @override
  bool operator ==(Object other) =>
      other is AutonomyLevel &&
      other.value == value &&
      other.isKnown == isKnown;

  @override
  int get hashCode => Object.hash(value, isKnown);

  @override
  String toString() =>
      isKnown ? 'AutonomyLevel.$value' : 'AutonomyLevel.unknown($value)';
}
