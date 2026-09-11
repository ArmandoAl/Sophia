import 'package:flutter/material.dart';

import '../theme/design_tokens.dart';
import 'motion/motion_widgets.dart';

class MapLocationCard extends StatelessWidget {
  const MapLocationCard({super.key});

  @override
  Widget build(BuildContext context) => Container(
    width: 280,
    decoration: BoxDecoration(
      color: context.colors.elevated,
      borderRadius: BorderRadius.circular(SophiaRadius.card),
      border: Border.all(color: context.colors.line),
    ),
    clipBehavior: Clip.antiAlias,
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          height: 120,
          color: context.colors.line,
          child: Stack(
            alignment: Alignment.center,
            children: [
              Icon(
                Icons.location_on,
                color: context.colors.accent,
                size: SophiaSpace.xxl,
              ),
              Positioned(
                bottom: SophiaSpace.xs,
                right: SophiaSpace.xs,
                child: Text(
                  'Parking Row C',
                  style: Theme.of(context).textTheme.labelSmall,
                ),
              ),
            ],
          ),
        ),
        Padding(
          padding: const EdgeInsets.all(SophiaSpace.sm),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Parking Row C, Level 2',
                style: Theme.of(context).textTheme.labelLarge,
              ),
              const SizedBox(height: SophiaSpace.xxs),
              Text(
                'Your car is parked here.',
                style: Theme.of(context).textTheme.bodySmall,
              ),
              const SizedBox(height: SophiaSpace.xs),
              TactileButton(
                onPressed: () {},
                child: Container(
                  width: double.infinity,
                  padding: const EdgeInsets.all(SophiaSpace.sm),
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: context.colors.accent,
                    borderRadius: BorderRadius.circular(SophiaRadius.control),
                  ),
                  child: Text(
                    'Open Navigation',
                    style: TextStyle(color: context.colors.surface),
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    ),
  );
}
