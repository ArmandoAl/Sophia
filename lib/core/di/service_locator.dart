import 'package:get_it/get_it.dart';
import '../../features/dashboard/presentation/cubit/dashboard_cubit.dart';
import '../../features/settings/presentation/cubit/settings_cubit.dart';

final sl = GetIt.instance;

Future<void> initDependencies() async {
  // Features - Dashboard
  // Bloc
  sl.registerFactory(() => DashboardCubit());

  // Features - Settings
  sl.registerFactory(() => SettingsCubit());

  // UseCases y Repositories se registrarían aquí
  // sl.registerLazySingleton(() => GetSystemHealth(sl()));
  // sl.registerLazySingleton<DashboardRepository>(() => DashboardRepositoryImpl(sl()));
}
