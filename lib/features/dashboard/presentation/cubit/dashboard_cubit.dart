import 'package:flutter_bloc/flutter_bloc.dart';
import 'dashboard_state.dart';
import '../../domain/system_health.dart';

class DashboardCubit extends Cubit<DashboardState> {
  // Aquí inyectarías el UseCase en el constructor
  DashboardCubit() : super(DashboardInitial());

  void loadDashboardData() async {
    emit(DashboardLoading());
    // Simulación de delay y datos (reemplazar con UseCase real)
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
