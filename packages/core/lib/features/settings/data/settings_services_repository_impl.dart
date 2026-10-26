import '../../../core/network/api_client.dart';
import '../domain/models.dart';
import '../domain/settings_services_repository.dart';

class SettingsServicesRepositoryImpl implements SettingsServicesRepository {
  SettingsServicesRepositoryImpl(this._api);

  final ApiClient _api;

  @override
  Future<Map<String, dynamic>> exportData() async =>
      await _api.get('/users/me/export') as Map<String, dynamic>;

  @override
  Future<void> requestDeletion() => _api.post('/users/me/delete-request');

  @override
  Future<List<DeviceToken>> listDeviceTokens() async =>
      (await _api.get('/notifications/device-tokens') as List<dynamic>)
          .cast<Map<String, dynamic>>()
          .map(DeviceToken.fromJson)
          .toList();

  @override
  Future<DeviceToken> registerDevice({
    required String platform,
    required String token,
  }) async => DeviceToken.fromJson(
    await _api.post(
          '/notifications/device-tokens',
          body: {'platform': platform, 'token': token},
        )
        as Map<String, dynamic>,
  );

  @override
  Future<void> removeDevice(String id) =>
      _api.delete('/notifications/device-tokens/$id');

  @override
  Future<List<IngestionBatch>> listIngestionBatches() async =>
      (await _api.get('/ingestion/batches') as List<dynamic>)
          .cast<Map<String, dynamic>>()
          .map(IngestionBatch.fromJson)
          .toList();

  @override
  Future<void> undoIngestionBatch(String id) =>
      _api.delete('/ingestion/batches/$id');
}
