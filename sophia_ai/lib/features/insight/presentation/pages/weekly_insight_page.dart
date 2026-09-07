import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'dart:math';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import '../../../../core/widgets/sophia_card.dart';
import '../cubit/insight_cubit.dart';
import '../cubit/insight_state.dart';

class WeeklyInsightPage extends StatelessWidget {
  const WeeklyInsightPage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) => InsightCubit(),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          appBar: AppBar(
            title: const Text("Weekly Insight"),
            actions: [
              IconButton(icon: const Icon(Icons.more_vert), onPressed: () {}),
            ],
          ),
          body: const _InsightBody(),
          bottomNavigationBar: Padding(
            padding: const EdgeInsets.all(24.0),
            child: ElevatedButton(
              onPressed: () {},
              style: ElevatedButton.styleFrom(
                backgroundColor: const Color(0xFFFFD60A), // Amarillo del diseño
                foregroundColor: Colors.black,
                padding: const EdgeInsets.symmetric(vertical: 16),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
              ),
              child: const Text(
                "Commit Learning",
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 16),
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _InsightBody extends StatelessWidget {
  const _InsightBody();

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(24),
      child: Column(
        children: [
          // 1. Gráfico Circular Amarillo
          BlocBuilder<InsightCubit, InsightState>(
            builder: (context, state) {
              return _AccuracyCircle(percentage: state.accuracy);
            },
          ),
          const SizedBox(height: 20),
          const Text(
            "Sophia is learning and improving every day thanks to your feedback.",
            textAlign: TextAlign.center,
            style: TextStyle(color: Colors.white70),
          ),
          const SizedBox(height: 40),

          // 2. Sección de Decisiones
          Align(
            alignment: Alignment.centerLeft,
            child: Text(
              "Decisions Needing Review",
              style: Theme.of(context).textTheme.titleMedium,
            ),
          ),
          const SizedBox(height: 16),

          BlocBuilder<InsightCubit, InsightState>(
            builder: (context, state) {
              return Column(
                children: state.pendingDecisions
                    .map((decision) => _DecisionCard(item: decision))
                    .toList(),
              );
            },
          ),
        ],
      ),
    );
  }
}

class _AccuracyCircle extends StatelessWidget {
  final double percentage;

  const _AccuracyCircle({required this.percentage});

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 200,
      width: 200,
      child: Stack(
        alignment: Alignment.center,
        children: [
          CustomPaint(
            size: const Size(200, 200),
            painter: _YellowCirclePainter(percentage: percentage),
          ),
          Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                "${(percentage * 100).toInt()}%",
                style: const TextStyle(
                  fontSize: 48,
                  fontWeight: FontWeight.bold,
                  color: Colors.white,
                ),
              ),
              const Text("Accuracy", style: TextStyle(color: Colors.grey)),
            ],
          ),
        ],
      ),
    );
  }
}

class _DecisionCard extends StatelessWidget {
  final DecisionItem item;

  const _DecisionCard({required this.item});

  @override
  Widget build(BuildContext context) {
    return SophiaCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Container(
                padding: const EdgeInsets.all(10),
                decoration: const BoxDecoration(
                  color: Color(0xFF2C2C2E),
                  shape: BoxShape.circle,
                ),
                child: Text(
                  item.iconEmoji,
                  style: const TextStyle(fontSize: 20),
                ),
              ),
              const SizedBox(width: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    item.title,
                    style: const TextStyle(
                      fontWeight: FontWeight.bold,
                      color: Colors.white,
                    ),
                  ),
                  Text(
                    item.timestamp,
                    style: const TextStyle(color: Colors.grey, fontSize: 12),
                  ),
                ],
              ),
            ],
          ),
          const SizedBox(height: 16),
          Text(item.reasoning, style: const TextStyle(color: Colors.white70)),
          const SizedBox(height: 24),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              _ActionBtn(
                icon: Icons.close,
                label: "Reject",
                color: Colors.redAccent,
                onTap: () {},
              ),
              _ActionBtn(
                icon: Icons.edit_note,
                label: "Add Note",
                color: Colors.grey,
                onTap: () {},
              ),
              _ActionBtn(
                icon: Icons.check,
                label: "Approve",
                color: Colors.greenAccent,
                onTap: () {},
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class _ActionBtn extends StatelessWidget {
  final IconData icon;
  final String label;
  final Color color;
  final VoidCallback onTap;

  const _ActionBtn({
    required this.icon,
    required this.label,
    required this.color,
    required this.onTap,
  });

  @override
  Widget build(BuildContext context) {
    return InkWell(
      onTap: onTap,
      child: Column(
        children: [
          Icon(icon, color: color),
          const SizedBox(height: 4),
          Text(label, style: TextStyle(color: color, fontSize: 12)),
        ],
      ),
    );
  }
}

// Pintor simple para el círculo amarillo (Screen 13)
class _YellowCirclePainter extends CustomPainter {
  final double percentage;
  _YellowCirclePainter({required this.percentage});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2 - 10;

    // Track (Fondo oscuro)
    final trackPaint = Paint()
      ..color = const Color(0xFF2C2C2E)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 20;
    canvas.drawCircle(center, radius, trackPaint);

    // Progress (Amarillo)
    final progressPaint = Paint()
      ..color = const Color(0xFFFFD60A)
      ..style = PaintingStyle.stroke
      ..strokeWidth = 20
      ..strokeCap = StrokeCap.round;

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      -pi / 2, // Empieza arriba
      2 * pi * percentage,
      false,
      progressPaint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => false;
}
