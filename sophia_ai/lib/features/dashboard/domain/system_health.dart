/// Prototype UI metrics for the glassmorphism dashboard mock.
///
/// **Not** the Sofia Backend health contract. Real backend status uses
/// [HealthResponse] via [HealthCubit] (`GET /health`).
///
/// Kept only so existing dashboard UI compiles until F1 UI wiring replaces it.
class SystemHealth {
  final double cpuUsage;
  final double ramUsage;
  final List<double> historyGraph; // Para la gráfica de líneas

  SystemHealth({
    required this.cpuUsage,
    required this.ramUsage,
    required this.historyGraph,
  });

  factory SystemHealth.fromJson(Map<String, dynamic> json) {
    return SystemHealth(
      cpuUsage: json['cpuUsage'] as double,
      ramUsage: json['ramUsage'] as double,
      historyGraph: json['historyGraph'] as List<double>,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'cpuUsage': cpuUsage,
      'ramUsage': ramUsage,
      'historyGraph': historyGraph,
    };
  }
}
