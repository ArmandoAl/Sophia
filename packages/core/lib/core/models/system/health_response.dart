import 'package:equatable/equatable.dart';

/// GET `/health` — flat response (no wrapper).
class HealthResponse extends Equatable {
  const HealthResponse({
    required this.status,
    required this.environment,
    required this.firestore,
  });

  final String status;
  final String environment;
  final String firestore;

  bool get isOk => status == 'ok';
  bool get isDegraded => status == 'degraded';

  factory HealthResponse.fromJson(Map<String, dynamic> json) {
    return HealthResponse(
      status: json['status'] as String,
      environment: json['environment'] as String,
      firestore: json['firestore'] as String,
    );
  }

  Map<String, dynamic> toJson() => {
    'status': status,
    'environment': environment,
    'firestore': firestore,
  };

  @override
  List<Object?> get props => [status, environment, firestore];
}
