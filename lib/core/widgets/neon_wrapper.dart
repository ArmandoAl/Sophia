import 'dart:ui';
import 'dart:math' as math;
import 'package:flutter/material.dart';

class NeonWrapper extends StatefulWidget {
  final Widget child;

  const NeonWrapper({super.key, required this.child});

  @override
  State<NeonWrapper> createState() => _NeonWrapperState();
}

class _NeonWrapperState extends State<NeonWrapper>
    with TickerProviderStateMixin {
  late AnimationController _blobController;
  late AnimationController _particleController;

  @override
  void initState() {
    super.initState();

    // Animación para los blobs (lenta y suave)
    _blobController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 8),
    )..repeat(reverse: true);

    // Animación para las partículas (más rápida)
    _particleController = AnimationController(
      vsync: this,
      duration: const Duration(seconds: 20),
    )..repeat();
  }

  @override
  void dispose() {
    _blobController.dispose();
    _particleController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final size = MediaQuery.of(context).size;

    return Stack(
      children: [
        // 1. Capa Base Oscura
        Container(color: theme.colorScheme.surface),

        // 2. Blob Azul Animado (Top Left)
        AnimatedBuilder(
          animation: _blobController,
          builder: (context, child) {
            final moveX = math.sin(_blobController.value * math.pi * 2) * 30;
            final moveY = math.cos(_blobController.value * math.pi * 2) * 40;
            final scale = 0.9 + (_blobController.value * 0.2);
            final opacity = 0.15 + (_blobController.value * 0.1);

            return Positioned(
              left: -100 + moveX,
              top: -100 + moveY,
              child: Transform.scale(
                scale: scale,
                child: Container(
                  width: size.width * 0.8,
                  height: size.width * 0.8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: theme.primaryColor.withValues(alpha: opacity),
                    boxShadow: [
                      BoxShadow(
                        color: theme.primaryColor.withValues(
                          alpha: opacity + 0.1,
                        ),
                        blurRadius: 120,
                        spreadRadius: 40,
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),

        // 3. Blob Morado Animado (Bottom Right)
        AnimatedBuilder(
          animation: _blobController,
          builder: (context, child) {
            final moveX = math.cos(_blobController.value * math.pi * 2) * 40;
            final moveY = math.sin(_blobController.value * math.pi * 2) * 30;
            final scale = 0.85 + ((1 - _blobController.value) * 0.25);
            final opacity = 0.1 + ((1 - _blobController.value) * 0.15);

            return Positioned(
              right: -100 + moveX,
              bottom: -100 + moveY,
              child: Transform.scale(
                scale: scale,
                child: Container(
                  width: size.width * 0.8,
                  height: size.width * 0.8,
                  decoration: BoxDecoration(
                    shape: BoxShape.circle,
                    color: theme.colorScheme.secondary.withValues(
                      alpha: opacity,
                    ),
                    boxShadow: [
                      BoxShadow(
                        color: theme.colorScheme.secondary.withValues(
                          alpha: opacity + 0.1,
                        ),
                        blurRadius: 120,
                        spreadRadius: 40,
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        ),

        // 4. Partículas Flotantes
        AnimatedBuilder(
          animation: _particleController,
          builder: (context, child) {
            return CustomPaint(
              size: size,
              painter: _ParticlePainter(
                animation: _particleController,
                primaryColor: theme.primaryColor,
                secondaryColor: theme.colorScheme.secondary,
              ),
            );
          },
        ),

        // 5. Blur General
        Positioned.fill(
          child: BackdropFilter(
            filter: ImageFilter.blur(sigmaX: 30, sigmaY: 30),
            child: Container(color: Colors.transparent),
          ),
        ),

        // 6. El contenido real de la pantalla
        widget.child,
      ],
    );
  }
}

// Painter para las partículas flotantes
class _ParticlePainter extends CustomPainter {
  final Animation<double> animation;
  final Color primaryColor;
  final Color secondaryColor;
  final math.Random _random = math.Random(42); // Seed fijo para consistencia

  _ParticlePainter({
    required this.animation,
    required this.primaryColor,
    required this.secondaryColor,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()..style = PaintingStyle.fill;

    // Generar 30 partículas
    for (int i = 0; i < 30; i++) {
      final baseX = _random.nextDouble() * size.width;
      final baseY = _random.nextDouble() * size.height;

      // Movimiento ondulante
      final offsetX =
          math.sin((animation.value * math.pi * 2) + (i * 0.5)) * 20;
      final offsetY = ((animation.value + (i * 0.03)) % 1.0) * size.height;

      final x = baseX + offsetX;
      final y = (baseY + offsetY) % size.height;

      // Alternar colores
      final color = i % 2 == 0 ? primaryColor : secondaryColor;
      final opacity = 0.05 + (_random.nextDouble() * 0.1);
      final radius = 1 + (_random.nextDouble() * 2);

      paint.color = color.withValues(alpha: opacity);

      // Dibujar partícula con glow
      canvas.drawCircle(
        Offset(x, y),
        radius,
        paint..maskFilter = const MaskFilter.blur(BlurStyle.normal, 3),
      );
    }
  }

  @override
  bool shouldRepaint(_ParticlePainter oldDelegate) => true;
}
