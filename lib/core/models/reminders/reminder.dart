import 'package:equatable/equatable.dart';

enum ReminderStatus { pending, due, delivered, failed, canceled, archived }

enum ReminderSource { user, activity, aiSuggested, aiCreated }

enum RecurrenceRule { none, daily, weekly, monthly }

ReminderStatus reminderStatus(String value) => ReminderStatus.values.firstWhere(
  (item) => item.name == value,
  orElse: () => ReminderStatus.pending,
);

ReminderSource reminderSource(String value) => ReminderSource.values.firstWhere(
  (item) =>
      item.name == value ||
      (item == ReminderSource.aiSuggested && value == 'ai_suggested') ||
      (item == ReminderSource.aiCreated && value == 'ai_created'),
  orElse: () => ReminderSource.user,
);

RecurrenceRule parseRecurrenceRule(String value) =>
    RecurrenceRule.values.firstWhere(
      (item) => item.name == value,
      orElse: () => RecurrenceRule.none,
    );

String recurrenceValue(RecurrenceRule rule) => rule.name;

class Reminder extends Equatable {
  const Reminder({
    required this.id,
    required this.userId,
    this.activityId,
    required this.title,
    required this.description,
    required this.status,
    required this.scheduledAt,
    required this.timezone,
    required this.recurrenceRule,
    required this.nextRunAt,
    this.lastRunAt,
    this.deliveredAt,
    this.failedAt,
    required this.failureReason,
    required this.retryCount,
    required this.maxRetries,
    required this.source,
    required this.createdBy,
    required this.createdAt,
    required this.updatedAt,
  });
  final String id,
      userId,
      title,
      description,
      timezone,
      failureReason,
      createdBy;
  final String? activityId;
  final ReminderStatus status;
  final DateTime scheduledAt, nextRunAt, createdAt, updatedAt;
  final DateTime? lastRunAt, deliveredAt, failedAt;
  final RecurrenceRule recurrenceRule;
  final int retryCount, maxRetries;
  final ReminderSource source;

  factory Reminder.fromJson(Map<String, dynamic> j) => Reminder(
    id: j['id'] as String,
    userId: j['user_id'] as String,
    activityId: j['activity_id'] as String?,
    title: j['title'] as String? ?? '',
    description: j['description'] as String? ?? '',
    status: reminderStatus(j['status'] as String? ?? 'pending'),
    scheduledAt: DateTime.parse(j['scheduled_at'] as String),
    timezone: j['timezone'] as String? ?? '',
    recurrenceRule: parseRecurrenceRule(
      j['recurrence_rule'] as String? ?? 'none',
    ),
    nextRunAt: DateTime.parse(j['next_run_at'] as String),
    lastRunAt: _date(j['last_run_at']),
    deliveredAt: _date(j['delivered_at']),
    failedAt: _date(j['failed_at']),
    failureReason: j['failure_reason'] as String? ?? '',
    retryCount: j['retry_count'] as int? ?? 0,
    maxRetries: j['max_retries'] as int? ?? 3,
    source: reminderSource(j['source'] as String? ?? 'user'),
    createdBy: j['created_by'] as String? ?? 'user',
    createdAt: DateTime.parse(j['created_at'] as String),
    updatedAt: DateTime.parse(j['updated_at'] as String),
  );
  static DateTime? _date(dynamic value) =>
      value == null ? null : DateTime.parse(value as String);
  @override
  List<Object?> get props => [
    id,
    userId,
    activityId,
    title,
    description,
    status,
    scheduledAt,
    timezone,
    recurrenceRule,
    nextRunAt,
    lastRunAt,
    deliveredAt,
    failedAt,
    failureReason,
    retryCount,
    maxRetries,
    source,
    createdBy,
    createdAt,
    updatedAt,
  ];
}
