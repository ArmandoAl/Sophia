import 'models.dart';

abstract interface class SettingsServicesRepository {
  Future<Map<String, dynamic>> exportData();
  Future<void> requestDeletion();
  Future<List<DeviceToken>> listDeviceTokens();
  Future<DeviceToken> registerDevice({
    required String platform,
    required String token,
  });
  Future<void> removeDevice(String id);
  Future<List<IngestionBatch>> listIngestionBatches();
  Future<void> undoIngestionBatch(String id);
}
