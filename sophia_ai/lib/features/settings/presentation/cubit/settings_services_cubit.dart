import 'dart:convert';

import 'package:equatable/equatable.dart';
import 'package:firebase_core/firebase_core.dart';
import 'package:firebase_messaging/firebase_messaging.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:share_plus/share_plus.dart';

import '../../domain/models.dart';
import '../../domain/settings_services_repository.dart';

class PrivacyCubit extends Cubit<String?> {
  PrivacyCubit(this._repository) : super(null);
  final SettingsServicesRepository _repository;

  Future<void> export() async {
    emit(null);
    try {
      final bytes = utf8.encode(
        const JsonEncoder.withIndent(
          '  ',
        ).convert(await _repository.exportData()),
      );
      await SharePlus.instance.share(
        ShareParams(
          files: [
            XFile.fromData(
              bytes,
              mimeType: 'application/json',
              name: 'sofia-data.json',
            ),
          ],
        ),
      );
    } catch (error) {
      emit('$error');
    }
  }

  Future<bool> requestDeletion() async {
    emit(null);
    try {
      await _repository.requestDeletion();
      return true;
    } catch (error) {
      emit('$error');
      return false;
    }
  }
}

class NotificationsState extends Equatable {
  const NotificationsState({
    this.devices = const [],
    this.currentDeviceId,
    this.loading = false,
    this.error,
  });
  final List<DeviceToken> devices;
  final String? currentDeviceId;
  final bool loading;
  final String? error;
  @override
  List<Object?> get props => [devices, currentDeviceId, loading, error];
}

class NotificationsCubit extends Cubit<NotificationsState> {
  NotificationsCubit(this._repository) : super(const NotificationsState());
  final SettingsServicesRepository _repository;

  Future<void> load() async {
    emit(
      NotificationsState(
        devices: state.devices,
        currentDeviceId: state.currentDeviceId,
        loading: true,
      ),
    );
    try {
      final devices = await _repository.listDeviceTokens();
      emit(
        NotificationsState(
          devices: devices,
          currentDeviceId: await _currentDeviceId(devices),
        ),
      );
    } catch (error) {
      emit(
        NotificationsState(
          devices: state.devices,
          currentDeviceId: state.currentDeviceId,
          error: '$error',
        ),
      );
    }
  }

  Future<void> register() async {
    try {
      if (Firebase.apps.isEmpty) {
        await Firebase.initializeApp();
      }
      final messaging = FirebaseMessaging.instance;
      await messaging.requestPermission();
      final token = await messaging.getToken();
      if (token == null) {
        throw StateError(
          'Este dispositivo no devolvió un token de notificaciones.',
        );
      }
      final registered = await _repository.registerDevice(
        platform: _platform,
        token: token,
      );
      emit(
        NotificationsState(
          devices: [...state.devices, registered],
          currentDeviceId: registered.id,
        ),
      );
      await load();
    } catch (error) {
      emit(
        NotificationsState(
          devices: state.devices,
          currentDeviceId: state.currentDeviceId,
          error: '$error',
        ),
      );
    }
  }

  Future<void> remove(String id) async {
    await _repository.removeDevice(id);
    await load();
  }

  String get _platform => kIsWeb
      ? 'web'
      : switch (defaultTargetPlatform) {
          TargetPlatform.iOS => 'ios',
          TargetPlatform.android => 'android',
          _ => throw UnsupportedError(
            'Las notificaciones sólo están disponibles en iOS, Android y web.',
          ),
        };

  Future<String?> _currentDeviceId(List<DeviceToken> devices) async {
    try {
      if (Firebase.apps.isEmpty) await Firebase.initializeApp();
      final token = await FirebaseMessaging.instance.getToken();
      if (token == null) return null;
      final redacted = token.length <= 8
          ? '***'
          : '${token.substring(0, 4)}...${token.substring(token.length - 4)}';
      return devices
          .where((device) => device.tokenRedacted == redacted)
          .firstOrNull
          ?.id;
    } catch (_) {
      return null;
    }
  }
}

class IngestionState extends Equatable {
  const IngestionState({
    this.batches = const [],
    this.loading = false,
    this.error,
  });
  final List<IngestionBatch> batches;
  final bool loading;
  final String? error;
  @override
  List<Object?> get props => [batches, loading, error];
}

class IngestionCubit extends Cubit<IngestionState> {
  IngestionCubit(this._repository) : super(const IngestionState());
  final SettingsServicesRepository _repository;

  Future<void> load() async {
    emit(IngestionState(batches: state.batches, loading: true));
    try {
      emit(IngestionState(batches: await _repository.listIngestionBatches()));
    } catch (error) {
      emit(IngestionState(batches: state.batches, error: '$error'));
    }
  }

  Future<void> undo(String id) async {
    await _repository.undoIngestionBatch(id);
    await load();
  }
}
