import 'dart:ui';
import 'package:flutter/material.dart';

class MapLocationCard extends StatelessWidget {
  const MapLocationCard({super.key});

  @override
  Widget build(BuildContext context) {
    return ClipRRect(
      borderRadius: BorderRadius.circular(16),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
        child: Container(
          width: 280,
          decoration: BoxDecoration(
            color: const Color(0xFF151B24).withValues(alpha: 0.8),
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.white.withValues(alpha: 0.1)),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // Fake Map Visual
              Container(
                height: 120,
                color: Colors.grey[800],
                child: Stack(
                  alignment: Alignment.center,
                  children: [
                    // Aquí iría el widget de Google Maps real
                    Container(color: const Color(0xFF2C3E50)),
                    const Icon(
                      Icons.location_on,
                      color: Colors.blueAccent,
                      size: 40,
                    ),
                    const Positioned(
                      bottom: 10,
                      right: 10,
                      child: Text(
                        "Parking Row C",
                        style: TextStyle(color: Colors.white70, fontSize: 10),
                      ),
                    ),
                  ],
                ),
              ),
              Padding(
                padding: const EdgeInsets.all(12.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text(
                      "Parking Row C, Level 2",
                      style: TextStyle(
                        color: Colors.white,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 4),
                    const Text(
                      "Your car is parked here.",
                      style: TextStyle(color: Colors.grey, fontSize: 12),
                    ),
                    const SizedBox(height: 8),
                    SizedBox(
                      width: double.infinity,
                      child: ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.blueAccent,
                        ),
                        onPressed: () {},
                        child: const Text(
                          "Open Navigation",
                          style: TextStyle(color: Colors.white),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
