import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

@immutable
class SophiaColors extends ThemeExtension<SophiaColors> {
  const SophiaColors({
    required this.surface,
    required this.elevated,
    required this.ink,
    required this.softInk,
    required this.muted,
    required this.line,
    required this.scrim,
    required this.accent,
    required this.positive,
    required this.attention,
    required this.critical,
  });

  final Color surface, elevated, ink, softInk, muted, line, scrim, accent;
  final Color positive, attention, critical;

  static const light = SophiaColors(
    surface: Color(0xFFF6F4F0),
    elevated: Color(0xFFFFFFFF),
    ink: Color(0xFF18201E),
    softInk: Color(0xFF52605C),
    muted: Color(0xFF66736E),
    line: Color(0xFFDDE2DF),
    scrim: Color(0xFF000000),
    accent: Color(0xFF39786B),
    positive: Color(0xFF337A58),
    attention: Color(0xFF8B5F20),
    critical: Color(0xFF9E4F57),
  );

  static const dark = SophiaColors(
    surface: Color(0xFF111715),
    elevated: Color(0xFF1A2220),
    ink: Color(0xFFF0F3F1),
    softInk: Color(0xFFB4C0BC),
    muted: Color(0xFF7F8D88),
    line: Color(0xFF303B37),
    scrim: Color(0xFF000000),
    accent: Color(0xFF73B6A6),
    positive: Color(0xFF72BD91),
    attention: Color(0xFFD5A65E),
    critical: Color(0xFFD98991),
  );

  @override
  SophiaColors copyWith({
    Color? surface,
    Color? elevated,
    Color? ink,
    Color? softInk,
    Color? muted,
    Color? line,
    Color? scrim,
    Color? accent,
    Color? positive,
    Color? attention,
    Color? critical,
  }) => SophiaColors(
    surface: surface ?? this.surface,
    elevated: elevated ?? this.elevated,
    ink: ink ?? this.ink,
    softInk: softInk ?? this.softInk,
    muted: muted ?? this.muted,
    line: line ?? this.line,
    scrim: scrim ?? this.scrim,
    accent: accent ?? this.accent,
    positive: positive ?? this.positive,
    attention: attention ?? this.attention,
    critical: critical ?? this.critical,
  );

  @override
  SophiaColors lerp(covariant SophiaColors? other, double t) {
    if (other == null) return this;
    return SophiaColors(
      surface: Color.lerp(surface, other.surface, t)!,
      elevated: Color.lerp(elevated, other.elevated, t)!,
      ink: Color.lerp(ink, other.ink, t)!,
      softInk: Color.lerp(softInk, other.softInk, t)!,
      muted: Color.lerp(muted, other.muted, t)!,
      line: Color.lerp(line, other.line, t)!,
      scrim: Color.lerp(scrim, other.scrim, t)!,
      accent: Color.lerp(accent, other.accent, t)!,
      positive: Color.lerp(positive, other.positive, t)!,
      attention: Color.lerp(attention, other.attention, t)!,
      critical: Color.lerp(critical, other.critical, t)!,
    );
  }
}

abstract final class SophiaSpace {
  // Retícula base de 4 pt; los saltos grandes mantienen una progresión contenida.
  static const double xxs = 4,
      xs = 8,
      sm = 12,
      md = 16,
      lg = 24,
      xl = 32,
      xxl = 48,
      xxxl = 64;
}

abstract final class SophiaRadius {
  // Control: campos y botones. Card: contenido agrupado. Sheet: capas modales.
  static const double control = 10, card = 16, sheet = 24;
}

abstract final class SophiaElevation {
  // Flat: estructura. Raised: contenido seleccionable. Overlay: capas temporales.
  static const double flat = 0, raised = 2, overlay = 8;
}

abstract final class SophiaSize {
  static const double navigationBreakpoint = 800;
  static const double sidebar = 240;
  static const double bottomBar = 68;
  static const double navigationIndicator = 2;
  static const double contentMaxWidth = 720;
  static const double messageMaxWidth = 340;
  static const double thinkingWidth = 48;
  static const double thinkingHeight = 24;
  static const double thinkingIndicator = 8;
  static const double chartHeight = 128;
  static const double minimumTapTarget = 48;
}

abstract final class SophiaDataViz {
  static const double minimumBarFraction = .08;
}

abstract final class SophiaOpacity {
  static const double faint = .08;
  static const double subtle = .14;
  static const double quiet = .3;
}

abstract final class SophiaType {
  // Escala modular aproximada 1.17: 12, 14, 16, 19, 23, 27, 32, 37, 44.
  static TextTheme theme(SophiaColors colors) =>
      GoogleFonts.manropeTextTheme().copyWith(
        displayLarge: _style(44, FontWeight.w600, 1.08, -1.1, colors.ink),
        displayMedium: _style(37, FontWeight.w600, 1.12, -.8, colors.ink),
        displaySmall: _style(32, FontWeight.w600, 1.16, -.5, colors.ink),
        headlineLarge: _style(32, FontWeight.w600, 1.18, -.4, colors.ink),
        headlineMedium: _style(27, FontWeight.w600, 1.2, -.3, colors.ink),
        headlineSmall: _style(23, FontWeight.w600, 1.24, -.2, colors.ink),
        titleLarge: _style(23, FontWeight.w600, 1.25, -.2, colors.ink),
        titleMedium: _style(19, FontWeight.w600, 1.32, 0, colors.ink),
        titleSmall: _style(16, FontWeight.w600, 1.38, .05, colors.ink),
        bodyLarge: _style(19, FontWeight.w400, 1.5, 0, colors.ink),
        bodyMedium: _style(16, FontWeight.w400, 1.5, .05, colors.ink),
        bodySmall: _style(14, FontWeight.w400, 1.45, .1, colors.softInk),
        labelLarge: _style(14, FontWeight.w600, 1.35, .2, colors.ink),
        labelMedium: _style(12, FontWeight.w600, 1.35, .3, colors.softInk),
        labelSmall: _style(12, FontWeight.w500, 1.35, .35, colors.muted),
      );

  /// Rol dato: lectura numérica estable, sin desplazamientos entre cifras.
  static TextStyle data(BuildContext context) => Theme.of(context)
      .textTheme
      .titleLarge!
      .copyWith(fontFeatures: const [FontFeature.tabularFigures()]);

  static TextStyle dataLabel(BuildContext context) => Theme.of(context)
      .textTheme
      .labelMedium!
      .copyWith(fontFeatures: const [FontFeature.tabularFigures()]);

  static TextStyle _style(
    double size,
    FontWeight weight,
    double height,
    double spacing,
    Color color,
  ) => TextStyle(
    fontSize: size,
    fontWeight: weight,
    height: height,
    letterSpacing: spacing,
    color: color,
  );
}

extension SophiaThemeContext on BuildContext {
  SophiaColors get colors =>
      Theme.of(this).extension<SophiaColors>() ??
      (Theme.of(this).brightness == Brightness.dark
          ? SophiaColors.dark
          : SophiaColors.light);
}
