import '../../../core/models/models.dart';

abstract class RemindersRepository {
  Future<Reminder> create(CreateReminderRequest request);
  Future<ReminderListResponse> list({
    int? limit,
    String? cursor,
    String? status,
    String? activityId,
    DateTime? from,
    DateTime? to,
  });
  Future<Reminder> getById(String id);
  Future<Reminder> update(String id, UpdateReminderRequest request);
  Future<Reminder> cancel(String id);
  Future<Reminder> archive(String id);
  Future<ReminderListResponse> listDue({int? limit});
}
