import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/motion/motion_widgets.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import '../cubit/smart_home_cubit.dart';
import 'package:sophia_ai/core/widgets/circular_ligh_control.dart';
import 'package:sophia_ai/core/widgets/device_card.dart';
import 'package:sophia_ai/features/smart_home/presentation/cubit/smart_home_state.dart';

class SmartHomePage extends StatelessWidget {
  const SmartHomePage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) => SmartHomeCubit(),
        child: Scaffold(
          backgroundColor: context.colors.surface.withValues(alpha: 0),
          appBar: AppBar(
            title: const Text("Living Room"),
            leading: TactileButton(
              semanticLabel: 'Volver',
              onPressed: () {},
              child: const Padding(
                padding: EdgeInsets.all(SophiaSpace.sm),
                child: Icon(Icons.arrow_back_ios_new),
              ),
            ),
            actions: [
              TactileButton(
                semanticLabel: 'Micrófono',
                onPressed: () {},
                child: const Padding(
                  padding: EdgeInsets.all(SophiaSpace.sm),
                  child: Icon(Icons.mic),
                ),
              ),
              const SizedBox(width: SophiaSpace.xs),
            ],
          ),
          body: const _SmartHomeBody(),
          // Reproductor de música flotante (Bottom Sheet simulado visualmente)
          bottomSheet: const _MusicPlayerBar(),
        ),
      ),
    );
  }
}

class _SmartHomeBody extends StatelessWidget {
  const _SmartHomeBody();

  @override
  Widget build(BuildContext context) {
    return SingleChildScrollView(
      child: Padding(
        padding: const EdgeInsets.all(SophiaSpace.lg),
        child: BlocBuilder<SmartHomeCubit, SmartHomeState>(
          builder: (context, state) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // 1. Control Principal (Circular)
                Center(
                  child: CircularLightControl(
                    percentage: state.mainLight.value,
                    label: state.mainLight.name,
                    subLabel: state.mainLight.statusText ?? "",
                  ),
                ),

                const SizedBox(height: SophiaSpace.xl),

                // 2. Scenes Selector
                Text("Scenes", style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: SophiaSpace.md),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: ["Cinema", "Reading", "Party", "Relax"].map((
                      scene,
                    ) {
                      final isActive = state.activeScene == scene;
                      return Padding(
                        padding: const EdgeInsets.only(right: SophiaSpace.sm),
                        child: TactileButton(
                          onPressed: () =>
                              context.read<SmartHomeCubit>().changeScene(scene),
                          child: Container(
                            padding: const EdgeInsets.symmetric(
                              horizontal: SophiaSpace.md,
                              vertical: SophiaSpace.sm,
                            ),
                            decoration: BoxDecoration(
                              color: isActive
                                  ? context.colors.accent
                                  : context.colors.elevated,
                              border: Border.all(color: context.colors.line),
                              borderRadius: BorderRadius.circular(
                                SophiaRadius.sheet,
                              ),
                            ),
                            child: Text(
                              scene,
                              style: TextStyle(
                                color: isActive
                                    ? context.colors.surface
                                    : context.colors.softInk,
                              ),
                            ),
                          ),
                        ),
                      );
                    }).toList(),
                  ),
                ),

                const SizedBox(height: SophiaSpace.xl),

                // 3. Grid de Dispositivos
                GridView.builder(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    childAspectRatio: 1.4,
                    crossAxisSpacing: SophiaSpace.md,
                    mainAxisSpacing: SophiaSpace.md,
                  ),
                  itemCount: state.devices.length,
                  itemBuilder: (context, index) {
                    final device = state.devices[index];
                    return DeviceCard(
                      device: device,
                      onTap: () => context.read<SmartHomeCubit>().toggleDevice(
                        device.id,
                      ),
                    );
                  },
                ),
                const SizedBox(height: SophiaSpace.xxxl),
              ],
            );
          },
        ),
      ),
    );
  }
}

// Widget del reproductor de música (Screen.png parte inferior)
class _MusicPlayerBar extends StatelessWidget {
  const _MusicPlayerBar();

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.all(SophiaSpace.md),
      padding: const EdgeInsets.symmetric(
        horizontal: SophiaSpace.md,
        vertical: SophiaSpace.sm,
      ),
      decoration: BoxDecoration(
        color: context.colors.elevated,
        borderRadius: BorderRadius.circular(SophiaRadius.card),
        border: Border.all(color: context.colors.line),
        boxShadow: [
          BoxShadow(
            color: context.colors.scrim.withValues(alpha: SophiaOpacity.subtle),
            blurRadius: SophiaSpace.sm,
            offset: const Offset(0, 5),
          ),
        ],
      ),
      child: Row(
        children: [
          Container(
            width: 40,
            height: 40,
            decoration: BoxDecoration(
              borderRadius: BorderRadius.circular(SophiaRadius.control),
              image: const DecorationImage(
                image: NetworkImage(
                  "https://picsum.photos/200",
                ), // Placeholder Art
                fit: BoxFit.cover,
              ),
            ),
          ),
          const SizedBox(width: SophiaSpace.sm),
          Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  "Starlight Wonder",
                  style: Theme.of(context).textTheme.labelLarge,
                ),
                Text(
                  "Casting to Nest Audio",
                  style: Theme.of(context).textTheme.labelSmall?.copyWith(
                    color: context.colors.accent,
                  ),
                ),
              ],
            ),
          ),
          TactileButton(
            semanticLabel: 'Pausar',
            onPressed: () {},
            child: const Padding(
              padding: EdgeInsets.all(SophiaSpace.sm),
              child: Icon(Icons.pause),
            ),
          ),
        ],
      ),
    );
  }
}
