import 'package:equatable/equatable.dart';

import '../../../../core/models/system/health_response.dart';
import '../../../../core/network/api_exception.dart';

sealed class HealthState extends Equatable {
  const HealthState();

  @override
  List<Object?> get props => [];
}

final class HealthInitial extends HealthState {
  const HealthInitial();
}

final class HealthLoading extends HealthState {
  const HealthLoading();
}

final class HealthHealthy extends HealthState {
  const HealthHealthy(this.health);

  final HealthResponse health;

  @override
  List<Object?> get props => [health];
}

final class HealthDegraded extends HealthState {
  const HealthDegraded(this.health);

  final HealthResponse health;

  @override
  List<Object?> get props => [health];
}

final class HealthFailure extends HealthState {
  const HealthFailure({required this.message, this.exception});

  final String message;
  final ApiException? exception;

  @override
  List<Object?> get props => [message, exception];
}
