import 'package:flutter/material.dart';

enum DeviceType { light, ac, plug, blinds }

class SmartDevice {
  final String id;
  final String name;
  final DeviceType type;
  final bool isOn;
  final double value; // Brillo (0-100) o Temperatura
  final String? statusText; // "4000K", "22°C", etc.
  final IconData icon;

  const SmartDevice({
    required this.id,
    required this.name,
    required this.type,
    this.isOn = false,
    this.value = 0,
    this.statusText,
    required this.icon,
  });

  SmartDevice copyWith({bool? isOn, double? value}) {
    return SmartDevice(
      id: id,
      name: name,
      type: type,
      icon: icon,
      statusText: statusText,
      isOn: isOn ?? this.isOn,
      value: value ?? this.value,
    );
  }
}
