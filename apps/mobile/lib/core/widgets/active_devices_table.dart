import 'package:flutter/material.dart';
import '../theme/design_tokens.dart';
import 'sophia_card.dart';

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
          const SizedBox(height: SophiaSpace.md),
          // Header
          Padding(
            padding: const EdgeInsets.only(bottom: SophiaSpace.xs),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  "Device Name",
                  style: TextStyle(color: context.colors.softInk),
                ),
                Text("Status", style: TextStyle(color: context.colors.softInk)),
              ],
            ),
          ),
          Divider(color: context.colors.line),
          // List Items (Simulados según Screen 12)
          _buildRow(
            context,
            "Living Room Hub",
            "Online",
            context.colors.positive,
          ),
          _buildRow(
            context,
            "Smart Thermostat",
            "Online",
            context.colors.positive,
          ),
          _buildRow(
            context,
            "Kitchen Display",
            "Warning",
            context.colors.attention,
          ),
          _buildRow(
            context,
            "Garage Door Sensor",
            "Offline",
            context.colors.critical,
          ),
          _buildRow(context, "Bedroom Lamp", "Online", context.colors.positive),
        ],
      ),
    );
  }

  Widget _buildRow(
    BuildContext context,
    String name,
    String status,
    Color color,
  ) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(name, style: Theme.of(context).textTheme.bodyMedium),
          Row(
            children: [
              Container(
                width: SophiaSpace.xs,
                height: SophiaSpace.xs,
                decoration: BoxDecoration(color: color, shape: BoxShape.circle),
              ),
              const SizedBox(width: SophiaSpace.xs),
              Text(status, style: TextStyle(color: context.colors.softInk)),
            ],
          ),
        ],
      ),
    );
  }
}
