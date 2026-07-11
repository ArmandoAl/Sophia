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
