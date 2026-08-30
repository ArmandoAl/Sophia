import 'package:flutter_test/flutter_test.dart';
import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/features/reminders/domain/reminders_repository.dart';
import 'package:sophia_ai/features/reminders/presentation/cubit/reminders_cubit.dart';

Reminder reminder() => Reminder.fromJson({
  'id': 'r1',
  'user_id': 'u1',
  'title': 'Study',
  'description': 'Go',
  'status': 'pending',
  'scheduled_at': '2026-07-16T20:00:00Z',
  'timezone': 'America/Tijuana',
  'recurrence_rule': 'weekly',
  'next_run_at': '2026-07-16T20:00:00Z',
  'retry_count': 0,
  'max_retries': 3,
  'source': 'user',
  'created_by': 'user',
  'created_at': '2026-07-15T20:00:00Z',
  'updated_at': '2026-07-15T20:00:00Z',
});

class FakeRemindersRepository implements RemindersRepository {
  final item = reminder();
  @override
  Future<Reminder> create(CreateReminderRequest request) async => item;
  @override
  Future<ReminderListResponse> list({
    int? limit,
    String? cursor,
    String? status,
    String? activityId,
    DateTime? from,
    DateTime? to,
  }) async => ReminderListResponse(reminders: [item]);
  @override
  Future<Reminder> getById(String id) async => item;
  @override
  Future<Reminder> update(String id, UpdateReminderRequest request) async =>
      item;
  @override
  Future<Reminder> cancel(String id) async => item;
  @override
  Future<Reminder> archive(String id) async => item;
  @override
  Future<ReminderListResponse> listDue({int? limit}) async =>
      ReminderListResponse(reminders: [item]);
}

void main() {
  test('Reminder and list response parse backend JSON', () {
    final parsed = reminder();
    expect(parsed.recurrenceRule, RecurrenceRule.weekly);
    expect(
      ReminderListResponse.fromJson({
        'reminders': [parsedJson()],
      }).reminders,
      hasLength(1),
    );
  });

  test('create request serializes supported fields and update omits nulls', () {
    final create = CreateReminderRequest(
      title: 'x',
      scheduledAt: '2026-07-16T20:00:00Z',
      timezone: 'UTC',
    );
    expect(create.toJson()['title'], 'x');
    expect(UpdateReminderRequest(title: 'new').toJson(), {'title': 'new'});
  });

  test('RemindersCubit loads and creates through repository', () async {
    final cubit = RemindersCubit(repository: FakeRemindersRepository());
    await cubit.load();
    expect(cubit.state.reminders.single.id, 'r1');
    expect(
      await cubit.create(
        CreateReminderRequest(
          title: 'x',
          scheduledAt: '2026-07-16T20:00:00Z',
          timezone: 'UTC',
        ),
      ),
      isTrue,
    );
    await cubit.close();
  });
}

Map<String, dynamic> parsedJson() => {
  'id': 'r2',
  'user_id': 'u1',
  'title': 'x',
  'description': '',
  'status': 'pending',
  'scheduled_at': '2026-07-16T20:00:00Z',
  'timezone': 'UTC',
  'recurrence_rule': 'none',
  'next_run_at': '2026-07-16T20:00:00Z',
  'retry_count': 0,
  'max_retries': 3,
  'source': 'user',
  'created_by': 'user',
  'created_at': '2026-07-15T20:00:00Z',
  'updated_at': '2026-07-15T20:00:00Z',
};
