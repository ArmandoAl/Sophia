class CreateReminderRequest {
  const CreateReminderRequest({
    this.activityId,
    required this.title,
    this.description = '',
    required this.scheduledAt,
    required this.timezone,
    this.recurrenceRule = 'none',
    this.recurrenceInterval = 0,
    this.recurrenceUntil,
    this.recurrenceCount = 0,
    this.maxRetries = 3,
  });
  final String? activityId, description, recurrenceUntil;
  final String title, scheduledAt, timezone, recurrenceRule;
  final int recurrenceInterval, recurrenceCount, maxRetries;
  Map<String, dynamic> toJson() => {
    'title': title,
    'description': description,
    'scheduled_at': scheduledAt,
    'timezone': timezone,
    'recurrence_rule': recurrenceRule,
    'recurrence_interval': recurrenceInterval,
    'recurrence_count': recurrenceCount,
    'max_retries': maxRetries,
    if (activityId != null) 'activity_id': activityId,
    if (recurrenceUntil != null) 'recurrence_until': recurrenceUntil,
  };
}

class UpdateReminderRequest {
  const UpdateReminderRequest({
    this.title,
    this.description,
    this.scheduledAt,
    this.timezone,
    this.recurrenceRule,
    this.recurrenceInterval,
    this.recurrenceUntil,
    this.recurrenceCount,
    this.maxRetries,
  });
  final String? title,
      description,
      scheduledAt,
      timezone,
      recurrenceRule,
      recurrenceUntil;
  final int? recurrenceInterval, recurrenceCount, maxRetries;
  Map<String, dynamic> toJson() => {
    'title': title,
    'description': description,
    'scheduled_at': scheduledAt,
    'timezone': timezone,
    'recurrence_rule': recurrenceRule,
    'recurrence_interval': recurrenceInterval,
    'recurrence_until': recurrenceUntil,
    'recurrence_count': recurrenceCount,
    'max_retries': maxRetries,
  }..removeWhere((_, value) => value == null);
}
