import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sophia_ai/core/models/reminders/reminder.dart';
import 'package:sophia_ai/core/models/reminders/reminder_requests.dart';
import 'package:sophia_ai/core/models/reminders/reminder_responses.dart';
import 'package:sophia_ai/core/theme/app_theme.dart';
import 'package:sophia_ai/core/theme/motion.dart';
import 'package:sophia_ai/features/dashboard/presentation/pages/dashboard_page.dart';
import 'package:sophia_ai/features/reminders/domain/reminders_repository.dart';
import 'package:sophia_ai/features/reminders/presentation/cubit/reminders_cubit.dart';

class _Repository implements RemindersRepository {
  _Repository(this.items);

  final List<Reminder> items;

  @override
  Future<ReminderListResponse> list({
    int? limit,
    String? cursor,
    String? status,
    String? activityId,
    DateTime? from,
    DateTime? to,
  }) async => ReminderListResponse(reminders: items);

  @override
  Future<Reminder> archive(String id) => throw UnimplementedError();

  @override
  Future<Reminder> cancel(String id) => throw UnimplementedError();

  @override
  Future<Reminder> create(CreateReminderRequest request) =>
      throw UnimplementedError();

  @override
  Future<Reminder> getById(String id) => throw UnimplementedError();

  @override
  Future<ReminderListResponse> listDue({int? limit}) =>
      throw UnimplementedError();

  @override
  Future<Reminder> update(String id, UpdateReminderRequest request) =>
      throw UnimplementedError();
}

Reminder _reminder(String id, String title, DateTime nextRunAt) =>
    Reminder.fromJson({
      'id': id,
      'user_id': 'user',
      'title': title,
      'description': '',
      'status': 'pending',
      'scheduled_at': nextRunAt.toUtc().toIso8601String(),
      'timezone': 'America/Tijuana',
      'recurrence_rule': 'none',
      'next_run_at': nextRunAt.toUtc().toIso8601String(),
      'retry_count': 0,
      'max_retries': 3,
      'source': 'user',
      'created_by': 'user',
      'created_at': nextRunAt.toUtc().toIso8601String(),
      'updated_at': nextRunAt.toUtc().toIso8601String(),
    });

void main() {
  testWidgets('prioritizes today and leaves future reminders below', (
    tester,
  ) async {
    final now = DateTime.now();
    final cubit = RemindersCubit(
      repository: _Repository([
        _reminder('today', 'Llamar a Ana', now),
        _reminder('later', 'Revisar mañana', now.add(const Duration(days: 1))),
      ]),
    );
    await cubit.load();

    await tester.pumpWidget(
      MaterialApp(
        theme: AppTheme.lightTheme,
        home: DashboardPage(cubit: cubit, autoLoad: false),
      ),
    );
    await tester.pump(SophiaMotion.long);

    expect(find.text('PARA HOY'), findsOneWidget);
    expect(find.text('Llamar a Ana'), findsOneWidget);
    expect(find.text('Después'), findsOneWidget);
    expect(find.text('Revisar mañana'), findsOneWidget);
    expect(find.text('recordatorio que merece tu atención'), findsOneWidget);
  });
}
