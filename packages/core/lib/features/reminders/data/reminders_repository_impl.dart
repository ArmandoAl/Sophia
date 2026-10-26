import '../../../core/models/models.dart';
import '../../../core/network/api_client.dart';
import '../../../core/network/require_json_map.dart';
import '../domain/reminders_repository.dart';

class RemindersRepositoryImpl implements RemindersRepository {
  RemindersRepositoryImpl(this._api);
  final ApiClient _api;
  Future<Reminder> _reminder(dynamic body, String context) =>
      Future.value(Reminder.fromJson(requireJsonMap(body, context: context)));
  @override
  Future<Reminder> create(CreateReminderRequest r) async => _reminder(
    await _api.post('/reminders', body: r.toJson()),
    'POST /reminders',
  );
  @override
  Future<ReminderListResponse> list({
    int? limit,
    String? cursor,
    String? status,
    String? activityId,
    DateTime? from,
    DateTime? to,
  }) async => ReminderListResponse.fromJson(
    requireJsonMap(
      await _api.get(
        '/reminders',
        queryParameters: _query(limit, cursor, status, activityId, from, to),
      ),
      context: 'GET /reminders',
    ),
  );
  @override
  Future<Reminder> getById(String id) async =>
      _reminder(await _api.get('/reminders/$id'), 'GET /reminders/{id}');
  @override
  Future<Reminder> update(String id, UpdateReminderRequest r) async =>
      _reminder(
        await _api.patch('/reminders/$id', body: r.toJson()),
        'PATCH /reminders/{id}',
      );
  @override
  Future<Reminder> cancel(String id) async => _reminder(
    await _api.post('/reminders/$id/cancel'),
    'POST /reminders/{id}/cancel',
  );
  @override
  Future<Reminder> archive(String id) async => _reminder(
    await _api.post('/reminders/$id/archive'),
    'POST /reminders/{id}/archive',
  );
  @override
  Future<ReminderListResponse> listDue({int? limit}) async =>
      ReminderListResponse.fromJson(
        requireJsonMap(
          await _api.get(
            '/reminders/due',
            queryParameters: limit == null ? null : {'limit': '$limit'},
          ),
          context: 'GET /reminders/due',
        ),
      );
  Map<String, String>? _query(
    int? limit,
    String? cursor,
    String? status,
    String? activityId,
    DateTime? from,
    DateTime? to,
  ) {
    final q = <String, String>{};
    if (limit != null) q['limit'] = '$limit';
    if (cursor?.isNotEmpty == true) q['cursor'] = cursor!;
    if (status?.isNotEmpty == true) q['status'] = status!;
    if (activityId?.isNotEmpty == true) q['activity_id'] = activityId!;
    if (from != null) q['from'] = from.toUtc().toIso8601String();
    if (to != null) q['to'] = to.toUtc().toIso8601String();
    return q.isEmpty ? null : q;
  }
}
