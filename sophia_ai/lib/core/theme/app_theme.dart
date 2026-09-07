import 'package:flutter/material.dart';
import 'package:google_fonts/google_fonts.dart';

class AppTheme {
  // Colores extraídos del diseño HTML/Tailwind
  static const _primaryColor = Color(0xFF1313EC); // Vivid Blue
  static const _secondaryColor = Color(0xFF9333EA); // Purple-600 approx
  static const _backgroundColor = Color(0xFF101022); // Background-dark
  static const _surfaceColor = Color(0xFF1A1F2E);

  static final ThemeData darkTheme = ThemeData(
    useMaterial3: true,
    brightness: Brightness.dark,
    scaffoldBackgroundColor: _backgroundColor, // Fondo base
    primaryColor: _primaryColor,

    // Definimos el esquema de colores
    colorScheme: const ColorScheme.dark(
      primary: _primaryColor,
      secondary: _secondaryColor,
      surface: _surfaceColor,
      error: Color(0xFFEF476F),
      onSurface: Colors.white,
    ),

    // Tipografía: Space Grotesk para ese look futurista del HTML
    textTheme: GoogleFonts.spaceGroteskTextTheme(
      ThemeData.dark().textTheme,
    ).apply(bodyColor: Colors.white, displayColor: Colors.white),

    // AppBar transparente para ver el fondo neon
    appBarTheme: const AppBarTheme(
      backgroundColor: Colors.transparent,
      elevation: 0,
      centerTitle: false,
      titleTextStyle: TextStyle(
        fontFamily: 'Space Grotesk',
        fontSize: 20,
        fontWeight: FontWeight.bold,
        color: Colors.white,
      ),
    ),
  );
}
