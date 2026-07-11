import 'package:flutter/material.dart';
import 'dart:math' as math;

/// Widget que muestra una visualización animada de ondas de sonido
/// cuando el usuario está hablando con el asistente de voz
class VoiceVisualizer extends StatefulWidget {
  const VoiceVisualizer({super.key});

  @override
  State<VoiceVisualizer> createState() => _VoiceVisualizerState();
}

class _VoiceVisualizerState extends State<VoiceVisualizer>
    with SingleTickerProviderStateMixin {
  late AnimationController _controller;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1500),
    )..repeat();
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: 60,
      child: AnimatedBuilder(
        animation: _controller,
        builder: (context, child) {
          return Row(
            mainAxisAlignment: MainAxisAlignment.center,
            crossAxisAlignment: CrossAxisAlignment.center,
            children: List.generate(5, (index) {
              // Crear diferentes fases para cada barra
              final phase = (index * 0.2) + _controller.value;
              final height = 10 + (math.sin(phase * math.pi * 2) * 20).abs();

              return Container(
                width: 4,
                height: height,
                margin: const EdgeInsets.symmetric(horizontal: 3),
                decoration: BoxDecoration(
                  gradient: const LinearGradient(
                    colors: [
                      Color(0xFF00B4D8), // Cyan
                      Color(0xFF4CC9F0), // Cyan claro
                    ],
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                  ),
                  borderRadius: BorderRadius.circular(4),
                  boxShadow: [
                    BoxShadow(
                      color: const Color(0xFF00B4D8).withValues(alpha: 0.5),
                      blurRadius: 8,
                      spreadRadius: 1,
                    ),
                  ],
                ),
              );
            }),
          );
        },
      ),
    );
  }
}
