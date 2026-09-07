import 'package:flutter/material.dart';
import '../../../../core/widgets/sophia_card.dart';

class ActiveDevicesTable extends StatelessWidget {
  const ActiveDevicesTable({super.key});

  @override
  Widget build(BuildContext context) {
    return SophiaCard(
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            "Active Matter Devices",
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: 16),
          // Header
          const Padding(
            padding: EdgeInsets.only(bottom: 8.0),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text("Device Name", style: TextStyle(color: Colors.grey)),
                Text("Status", style: TextStyle(color: Colors.grey)),
              ],
            ),
          ),
          const Divider(color: Colors.white10),
          // List Items (Simulados según Screen 12)
          _buildRow("Living Room Hub", "Online", Colors.greenAccent),
          _buildRow("Smart Thermostat", "Online", Colors.greenAccent),
          _buildRow("Kitchen Display", "Warning", Colors.amberAccent),
          _buildRow("Garage Door Sensor", "Offline", Colors.redAccent),
          _buildRow("Bedroom Lamp", "Online", Colors.greenAccent),
        ],
      ),
    );
  }

  Widget _buildRow(String name, String status, Color color) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 12.0),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(name, style: const TextStyle(color: Colors.white, fontSize: 16)),
          Row(
            children: [
              Container(
                width: 8,
                height: 8,
                decoration: BoxDecoration(color: color, shape: BoxShape.circle),
              ),
              const SizedBox(width: 8),
              Text(status, style: const TextStyle(color: Colors.grey)),
            ],
          ),
        ],
      ),
    );
  }
}
