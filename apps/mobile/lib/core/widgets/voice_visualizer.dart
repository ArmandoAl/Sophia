import 'package:flutter/material.dart';
import 'dart:math' as math;

import '../theme/design_tokens.dart';
import '../theme/motion.dart';

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
    _controller = AnimationController(vsync: this, duration: SophiaMotion.long);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (SophiaMotion.reduced(context)) {
      _controller
        ..stop()
        ..value = .5;
    } else if (!_controller.isAnimating) {
      _controller.repeat(reverse: true);
    }
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return SizedBox(
      height: SophiaSpace.xxxl,
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
                  gradient: LinearGradient(
                    colors: [
                      context.colors.accent,
                      context.colors.accent.withValues(alpha: .55),
                    ],
                    begin: Alignment.bottomCenter,
                    end: Alignment.topCenter,
                  ),
                  borderRadius: BorderRadius.circular(SophiaRadius.control),
                  boxShadow: [
                    BoxShadow(
                      color: context.colors.accent.withValues(alpha: .2),
                      blurRadius: SophiaSpace.xs,
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
