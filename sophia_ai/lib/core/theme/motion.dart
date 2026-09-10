import 'package:flutter/material.dart';
import 'package:flutter/physics.dart';

// INTERRUMPIBLE: el usuario conserva el control durante toda transición.
// CONTINUIDAD: lo que persiste se transforma; no desaparece para reaparecer.
// DEFERENCIA: el movimiento sirve al contenido y nunca compite con él.
// ORIGEN: cada elemento aparece desde el lugar que explica su procedencia.
abstract final class SophiaMotion {
  static const micro = Duration(milliseconds: 120);
  static const short = Duration(milliseconds: 220);
  static const medium = Duration(milliseconds: 320);
  static const long = Duration(milliseconds: 480);
  static const stagger = Duration(milliseconds: 40);

  static const double hierarchicalEnterOffset = 30;
  static const double hierarchicalExitOffset = 10;
  static const double hierarchicalScrimOpacity = .08;
  static const double lateralIncomingScale = 1.02;
  static const double lateralOutgoingScale = .98;
  static const double selectedTabScale = 1.08;
  static const double sheetBackdropScale = .94;
  static const double sheetBackdropScrimOpacity = .12;
  static const double userMessageOffset = 8;
  static const double assistantMessageOffset = 12;
  static const double proposalMessageOffset = 16;
  static const double proposalEntryScale = .98;
  static const double thinkingOpacityMin = .3;
  static const double thinkingOpacityMax = .72;

  static const structuralSpring = SpringDescription(
    mass: 1,
    stiffness: 420,
    damping: 38,
  );
  static const contentSpring = SpringDescription(
    mass: 1,
    stiffness: 300,
    damping: 28,
  );
  static const tactileSpring = SpringDescription(
    mass: 1,
    stiffness: 360,
    damping: 22,
  );

  static const structuralCurve = _SpringCurve(structuralSpring);
  static const contentCurve = _SpringCurve(contentSpring);
  static const tactileCurve = _SpringCurve(tactileSpring);
  static const assistantMessageCurve = Interval(1 / 9, 1, curve: contentCurve);

  static bool reduced(BuildContext context) =>
      MediaQuery.maybeOf(context)?.disableAnimations ?? false;

  static Duration resolve(BuildContext context, Duration value) =>
      reduced(context) ? Duration.zero : value;
}

class _SpringCurve extends Curve {
  const _SpringCurve(this.description);
  final SpringDescription description;

  @override
  double transformInternal(double t) =>
      SpringSimulation(description, 0, 1, 0).x(t * 1.25).clamp(0, 1.08);
}
