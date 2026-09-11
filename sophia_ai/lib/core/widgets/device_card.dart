import 'package:flutter/material.dart';
import 'package:sophia_ai/features/smart_home/domain/entities/smart_device.dart';

import '../theme/design_tokens.dart';
import 'sophia_card.dart';

class DeviceCard extends StatelessWidget {
  const DeviceCard({super.key, required this.device, required this.onTap});

  final SmartDevice device;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) {
    final active = device.isOn;
    return SophiaCard(
      onTap: onTap,
      borderColor: active ? context.colors.accent : context.colors.line,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Container(
            padding: const EdgeInsets.all(SophiaSpace.xs),
            decoration: BoxDecoration(
              color: context.colors.accent.withValues(
                alpha: SophiaOpacity.subtle,
              ),
              shape: BoxShape.circle,
            ),
            child: Icon(
              device.icon,
              color: active ? context.colors.accent : context.colors.softInk,
              size: SophiaSpace.lg,
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(device.name, style: Theme.of(context).textTheme.labelLarge),
              const SizedBox(height: SophiaSpace.xxs),
              Text(
                active ? (device.statusText ?? 'On') : 'Off',
                style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: active
                      ? context.colors.accent
                      : context.colors.softInk,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
