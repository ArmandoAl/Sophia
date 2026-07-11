import 'dart:math';
import 'package:flutter/material.dart';

class CircularLightControl extends StatelessWidget {
  final double percentage; // 0 a 100
  final String label;
  final String subLabel;

  const CircularLightControl({
    super.key,
    required this.percentage,
    required this.label,
    required this.subLabel,
  });

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 300,
      width: 300,
      child: Stack(
        alignment: Alignment.center,
        children: [
          // El pintor del arco
          CustomPaint(
            size: const Size(300, 300),
            painter: _ArcPainter(percentage: percentage),
          ),
          // Texto Central
          Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                label,
                style: const TextStyle(color: Colors.white70, fontSize: 16),
              ),
              const SizedBox(height: 8),
              Text(
                "${percentage.toInt()}%",
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 56,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 8),
              Text(
                subLabel,
                style: const TextStyle(
                  color: Color(0xFF2E5CB8),
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
          // Indicadores visuales (puntos blancos a los lados como en el diseño)
          Positioned(left: 40, top: 140, child: _glowDot()),
          Positioned(right: 40, top: 140, child: _glowDot()),
        ],
      ),
    );
  }

  Widget _glowDot() {
    return Container(
      width: 12,
      height: 12,
      decoration: BoxDecoration(
        color: Colors.white,
        shape: BoxShape.circle,
        boxShadow: [
          BoxShadow(
            color: Colors.white.withValues(alpha: 0.5),
            blurRadius: 10,
            spreadRadius: 2,
          ),
        ],
      ),
    );
  }
}

class _ArcPainter extends CustomPainter {
  final double percentage;

  _ArcPainter({required this.percentage});

  @override
  void paint(Canvas canvas, Size size) {
    final center = Offset(size.width / 2, size.height / 2);
    final radius = size.width / 2 - 20;

    // 1. Fondo (Track oscuro)
    final backgroundPaint = Paint()
      ..color =
          const Color(0xFF1A212A) // Color oscuro del track
      ..style = PaintingStyle.stroke
      ..strokeWidth = 35
      ..strokeCap = StrokeCap.round;

    // Dibujamos un arco de 270 grados (dejando la parte inferior abierta)
    // startAngle: 135 grados (en radianes) -> pi * 0.75
    // sweepAngle: 270 grados (en radianes) -> pi * 1.5
    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      pi * 0.75,
      pi * 1.5,
      false,
      backgroundPaint,
    );

    // 2. Progreso (Gradiente Azul Neon)
    final gradient = const SweepGradient(
      startAngle: pi * 0.75,
      endAngle:
          pi * 2.25, // Ajustado para que el gradiente se vea bien en el arco
      colors: [
        Color(0xFF3A0CA3), // Morado oscuro
        Color(0xFF4361EE), // Azul
        Color(0xFF4CC9F0), // Cyan brillante
      ],
      tileMode: TileMode.repeated,
    ).createShader(Rect.fromCircle(center: center, radius: radius));

    final progressPaint = Paint()
      ..shader = gradient
      ..style = PaintingStyle.stroke
      ..strokeWidth = 35
      ..strokeCap = StrokeCap.round;

    // Calculamos cuánto arco pintar según el porcentaje
    final progressSweep = (pi * 1.5) * (percentage / 100);

    canvas.drawArc(
      Rect.fromCircle(center: center, radius: radius),
      pi * 0.75,
      progressSweep,
      false,
      progressPaint,
    );
  }

  @override
  bool shouldRepaint(covariant CustomPainter oldDelegate) => true;
}
