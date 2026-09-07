import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
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
          backgroundColor: Colors.transparent,
          appBar: AppBar(
            title: const Text("Living Room"),
            leading: IconButton(
              icon: const Icon(Icons.arrow_back_ios_new, size: 20),
              onPressed: () {}, // GoRouter pop
            ),
            actions: [
              IconButton(icon: const Icon(Icons.mic), onPressed: () {}),
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
        padding: const EdgeInsets.all(24.0),
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

                const SizedBox(height: 40),

                // 2. Scenes Selector
                Text("Scenes", style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 16),
                SingleChildScrollView(
                  scrollDirection: Axis.horizontal,
                  child: Row(
                    children: ["Cinema", "Reading", "Party", "Relax"].map((
                      scene,
                    ) {
                      final isActive = state.activeScene == scene;
                      return Padding(
                        padding: const EdgeInsets.only(right: 12),
                        child: ChoiceChip(
                          label: Text(scene),
                          selected: isActive,
                          onSelected: (_) =>
                              context.read<SmartHomeCubit>().changeScene(scene),
                          selectedColor: const Color(0xFF4361EE),
                          backgroundColor: const Color(0xFF151B24),
                          labelStyle: TextStyle(
                            color: isActive ? Colors.white : Colors.grey,
                          ),
                          padding: const EdgeInsets.symmetric(
                            horizontal: 16,
                            vertical: 12,
                          ),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(30),
                          ),
                          showCheckmark: false,
                          side: BorderSide.none,
                        ),
                      );
                    }).toList(),
                  ),
                ),

                const SizedBox(height: 30),

                // 3. Grid de Dispositivos
                GridView.builder(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
                    crossAxisCount: 2,
                    childAspectRatio: 1.4,
                    crossAxisSpacing: 16,
                    mainAxisSpacing: 16,
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
                const SizedBox(height: 80), // Espacio para el Music Player
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
      margin: const EdgeInsets.all(16),
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: const Color(0xFF1A1F36),
        borderRadius: BorderRadius.circular(16),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.3),
            blurRadius: 10,
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
              borderRadius: BorderRadius.circular(8),
              image: const DecorationImage(
                image: NetworkImage(
                  "https://picsum.photos/200",
                ), // Placeholder Art
                fit: BoxFit.cover,
              ),
            ),
          ),
          const SizedBox(width: 12),
          const Expanded(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  "Starlight Wonder",
                  style: TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Text(
                  "Casting to Nest Audio",
                  style: TextStyle(color: Color(0xFF3F51B5), fontSize: 12),
                ),
              ],
            ),
          ),
          IconButton(
            onPressed: () {},
            icon: const Icon(Icons.pause, color: Colors.white),
          ),
        ],
      ),
    );
  }
}
