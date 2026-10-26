/// Wire value for `proactivity_level`.
///
/// Known values: `low`, `medium`, `high`.
/// Unknown backend values parse without throwing and keep [value] as raw.
class ProactivityLevel {
  const ProactivityLevel._(this.value, {required this.isKnown});

  final String value;
  final bool isKnown;

  static const low = ProactivityLevel._('low', isKnown: true);
  static const medium = ProactivityLevel._('medium', isKnown: true);
  static const high = ProactivityLevel._('high', isKnown: true);

  static const List<ProactivityLevel> knownValues = [low, medium, high];

  factory ProactivityLevel.fromJson(String raw) {
    for (final level in knownValues) {
      if (level.value == raw) return level;
    }
    return ProactivityLevel._(raw, isKnown: false);
  }

  /// Encodes only known values for requests.
  String toJson() {
    if (!isKnown) {
      throw StateError(
        'Cannot encode unknown proactivity_level "$value" in requests',
      );
    }
    return value;
  }

  @override
  bool operator ==(Object other) =>
      other is ProactivityLevel &&
      other.value == value &&
      other.isKnown == isKnown;

  @override
  int get hashCode => Object.hash(value, isKnown);

  @override
  String toString() =>
      isKnown ? 'ProactivityLevel.$value' : 'ProactivityLevel.unknown($value)';
}
