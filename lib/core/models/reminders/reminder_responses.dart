import 'reminder.dart';

class ReminderListResponse {
  const ReminderListResponse({required this.reminders, this.nextCursor});
  final List<Reminder> reminders;
  final String? nextCursor;
  factory ReminderListResponse.fromJson(Map<String, dynamic> j) =>
      ReminderListResponse(
        reminders: ((j['reminders'] as List?) ?? [])
            .map((e) => Reminder.fromJson(e as Map<String, dynamic>))
            .toList(),
        nextCursor: j['next_cursor'] as String?,
      );
}
