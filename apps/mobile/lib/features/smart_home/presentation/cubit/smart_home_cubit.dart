import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/material.dart';
import '../../domain/entities/smart_device.dart';
import 'smart_home_state.dart';

// --- Cubit ---
class SmartHomeCubit extends Cubit<SmartHomeState> {
  SmartHomeCubit()
    : super(
        SmartHomeState(
          activeScene: 'Cinema',
          mainLight: const SmartDevice(
            id: 'hue_1',
            name: 'Philips Hue',
            type: DeviceType.light,
            isOn: true,
            value: 75,
            statusText: '4000K',
            icon: Icons.lightbulb,
          ),
          devices: [
            const SmartDevice(
              id: '1',
              name: 'Smart Plugs',
              type: DeviceType.plug,
              isOn: true,
              icon: Icons.power,
            ),
            const SmartDevice(
              id: '2',
              name: 'AC Unit',
              type: DeviceType.ac,
              isOn: false,
              statusText: '22°C',
              icon: Icons.ac_unit,
            ),
            const SmartDevice(
              id: '3',
              name: 'Blinds',
              type: DeviceType.blinds,
              isOn: false,
              statusText: 'Open',
              icon: Icons.blinds,
            ),
          ],
        ),
      );

  void toggleDevice(String id) {
    // Lógica para encender/apagar dispositivos de la lista
    final updatedDevices = state.devices.map((device) {
      if (device.id == id) return device.copyWith(isOn: !device.isOn);
      return device;
    }).toList();

    emit(
      SmartHomeState(
        devices: updatedDevices,
        activeScene: state.activeScene,
        mainLight: state.mainLight,
      ),
    );
  }

  void updateMainLightBrightness(double newVal) {
    emit(
      SmartHomeState(
        devices: state.devices,
        activeScene: state.activeScene,
        mainLight: state.mainLight.copyWith(value: newVal),
      ),
    );
  }

  void changeScene(String sceneName) {
    emit(
      SmartHomeState(
        devices: state.devices,
        activeScene: sceneName,
        mainLight:
            state.mainLight, // Aquí podrías cambiar el brillo según la escena
      ),
    );
  }
}
