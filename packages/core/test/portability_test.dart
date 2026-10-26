import 'dart:io';

import 'package:test/test.dart';

void main() {
  test('el core no importa Flutter, dart:ui, dart:io ni plugins', () {
    final forbidden = RegExp(
      r'''^\s*(import|export)\s+['"](package:(flutter|flutter_bloc|flutter_secure_storage|get_it|go_router|firebase_\w+|share_plus|google_fonts)\b|dart:(ui|io|html))''',
      multiLine: true,
    );
    final offenders = <String>[];
    for (final file in Directory('lib').listSync(recursive: true).whereType<File>().where((f) => f.path.endsWith('.dart'))) {
      if (forbidden.hasMatch(file.readAsStringSync())) offenders.add(file.path);
    }
    expect(offenders, isEmpty);
  });

  test('sólo depende de bloc, equatable y http', () {
    final pubspec = File('pubspec.yaml').readAsStringSync();
    final dependencies = pubspec.split('dev_dependencies:').first.split('dependencies:').last;
    final names = RegExp(r'^  (\w+):', multiLine: true).allMatches(dependencies).map((m) => m.group(1)).toSet();
    expect(names, {'bloc', 'equatable', 'http'});
  });
}
