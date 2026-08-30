import 'package:flutter_bloc/flutter_bloc.dart';
import 'dashboard_state.dart';
import '../../domain/system_health.dart';

/// Legacy dashboard mock (CPU/RAM). Prefer [HealthCubit] for backend health.
class DashboardCubit extends Cubit<DashboardState> {
  DashboardCubit() : super(DashboardInitial());

  void loadDashboardData() async {
    emit(DashboardLoading());
    // Prototype only — not GET /health. See HealthCubit for real backend status.
    await Future.delayed(const Duration(seconds: 1));

    emit(
      DashboardLoaded(
        SystemHealth(
          cpuUsage: 78,
          ramUsage: 62,
          historyGraph: [10, 40, 30, 70, 50, 90, 20, 80],
        ),
      ),
    );
  }
}
