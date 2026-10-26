import '../../../core/models/system/health_response.dart';

abstract interface class SystemRepository {
  Future<HealthResponse> getHealth();
}
