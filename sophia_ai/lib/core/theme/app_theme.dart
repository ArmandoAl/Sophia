import 'package:flutter/material.dart';
import 'design_tokens.dart';

class AppTheme {
  static final lightTheme = _theme(Brightness.light, SophiaColors.light);
  static final darkTheme = _theme(Brightness.dark, SophiaColors.dark);

  static ThemeData _theme(Brightness brightness, SophiaColors colors) =>
      ThemeData(
        useMaterial3: true,
        brightness: brightness,
        scaffoldBackgroundColor: colors.surface,
        colorScheme: ColorScheme(
          brightness: brightness,
          primary: colors.accent,
          onPrimary: colors.surface,
          secondary: colors.softInk,
          onSecondary: colors.surface,
          error: colors.critical,
          onError: colors.surface,
          surface: colors.elevated,
          onSurface: colors.ink,
        ),
        textTheme: SophiaType.theme(colors),
        extensions: [colors],
        disabledColor: colors.muted,
        shadowColor: colors.scrim.withValues(alpha: .14),
        dividerTheme: DividerThemeData(color: colors.line, thickness: 1),
        cardTheme: CardThemeData(
          color: colors.elevated,
          elevation: SophiaElevation.raised,
          shape: RoundedRectangleBorder(
            borderRadius: BorderRadius.circular(SophiaRadius.card),
          ),
        ),
        appBarTheme: AppBarTheme(
          backgroundColor: colors.surface,
          foregroundColor: colors.ink,
          elevation: SophiaElevation.flat,
          centerTitle: false,
        ),
        inputDecorationTheme: InputDecorationTheme(
          filled: true,
          fillColor: colors.elevated,
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(SophiaRadius.control),
            borderSide: BorderSide(color: colors.line),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(SophiaRadius.control),
            borderSide: BorderSide(color: colors.line),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(SophiaRadius.control),
            borderSide: BorderSide(color: colors.accent, width: 2),
          ),
        ),
        bottomSheetTheme: BottomSheetThemeData(
          backgroundColor: colors.elevated,
          modalBackgroundColor: colors.elevated,
          modalBarrierColor: colors.scrim.withValues(alpha: .28),
          elevation: SophiaElevation.overlay,
          shape: const RoundedRectangleBorder(
            borderRadius: BorderRadius.vertical(
              top: Radius.circular(SophiaRadius.sheet),
            ),
          ),
        ),
        focusColor: colors.accent.withValues(alpha: .22),
      );
}
