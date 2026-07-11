import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/widgets/sophia_card.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import '../cubit/dashboard_cubit.dart';
import '../cubit/dashboard_state.dart';

class DashboardPage extends StatelessWidget {
  const DashboardPage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) =>
            DashboardCubit()
              ..loadDashboardData(), // Inyección via GetIt iría aquí
        child: Scaffold(
          backgroundColor: Colors.transparent,
          body: SafeArea(
            child: Row(
              children: [
                // Sidebar Navigation (Simplificado para el ejemplo)
                if (MediaQuery.of(context).size.width > 800) const _SideBar(),

                // Contenido Principal
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(24.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          "Sophia System\nDiagnostics",
                          style: Theme.of(context).textTheme.headlineMedium
                              ?.copyWith(
                                fontWeight: FontWeight.bold,
                                color: Colors.white,
                              ),
                        ),
                        const SizedBox(height: 24),
                        Expanded(child: _DashboardContent()),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _DashboardContent extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return BlocBuilder<DashboardCubit, DashboardState>(
      builder: (context, state) {
        if (state is DashboardLoading) {
          return const Center(child: CircularProgressIndicator());
        } else if (state is DashboardLoaded) {
          return ListView(
            children: [
              // Columna Izquierda (System Health) - Basado en Screen12
              Column(
                children: [
                  SophiaCard(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          "System Health",
                          style: Theme.of(context).textTheme.titleLarge,
                        ),
                        const SizedBox(height: 20),
                        _buildStat(
                          "CPU Usage",
                          "${state.health.cpuUsage.toInt()}%",
                          Colors.tealAccent,
                        ),
                        const SizedBox(height: 16),
                        _buildStat(
                          "RAM Usage",
                          "${state.health.ramUsage.toInt()}%",
                          Colors.purpleAccent,
                        ),
                        const SizedBox(height: 20),
                        // Placeholder para la gráfica
                        Container(
                          height: 100,
                          width: double.infinity,
                          decoration: BoxDecoration(
                            gradient: LinearGradient(
                              colors: [
                                Colors.teal.withValues(alpha: 0.2),
                                Colors.transparent,
                              ],
                              begin: Alignment.topCenter,
                              end: Alignment.bottomCenter,
                            ),
                          ),
                          child: const Center(
                            child: Text("--- Graph Visualization ---"),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),

              const SizedBox(height: 20),
              // Columna Derecha (Logs)
              SophiaCard(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      "Real-time Logs",
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                    const Divider(color: Colors.white10),
                    const Text(
                      "[10:45:01] [INFO] System boot sequence initiated.",
                      style: TextStyle(
                        fontFamily: 'Courier',
                        color: Colors.grey,
                      ),
                    ),
                    const Text(
                      "[10:45:02] [INFO] Network interface eth0 connected.",
                      style: TextStyle(
                        fontFamily: 'Courier',
                        color: Colors.grey,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          );
        }
        return const SizedBox();
      },
    );
  }

  Widget _buildStat(String label, String value, Color color) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(label, style: const TextStyle(color: Colors.grey)),
        Text(
          value,
          style: TextStyle(
            color: color,
            fontSize: 32,
            fontWeight: FontWeight.bold,
          ),
        ),
      ],
    );
  }
}

class _SideBar extends StatelessWidget {
  const _SideBar();
  @override
  Widget build(BuildContext context) {
    return Container(
      width: 80,
      color: Theme.of(context).colorScheme.surface,
      child: Column(
        children: [
          const SizedBox(height: 30),
          const Icon(
            Icons.smart_toy,
            color: Colors.cyan,
            size: 32,
          ), // Logo Sophia
          const Spacer(),
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.dashboard, color: Colors.white),
          ),
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.chat_bubble, color: Colors.grey),
          ),
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.settings, color: Colors.grey),
          ),
          const SizedBox(height: 30),
        ],
      ),
    );
  }
}
