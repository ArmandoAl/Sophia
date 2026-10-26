import 'package:equatable/equatable.dart';
import '../../domain/entities/smart_device.dart';

// --- State ---
class SmartHomeState extends Equatable {
  final List<SmartDevice> devices;
  final String activeScene; // "Cinema", "Reading", etc.
  final SmartDevice mainLight; // El foco principal (Philips Hue)

  const SmartHomeState({
    required this.devices,
    required this.activeScene,
    required this.mainLight,
  });

  @override
  List<Object> get props => [devices, activeScene, mainLight];
}
