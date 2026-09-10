import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sophia_ai/core/theme/motion.dart';
import 'package:sophia_ai/core/widgets/main_wrapper.dart';

void main() {
  testWidgets('reduced motion collapses every named duration to zero', (
    tester,
  ) async {
    late BuildContext context;
    await tester.pumpWidget(
      MaterialApp(
        home: MediaQuery(
          data: const MediaQueryData(disableAnimations: true),
          child: Builder(
            builder: (value) {
              context = value;
              return const SizedBox();
            },
          ),
        ),
      ),
    );

    for (final duration in [
      SophiaMotion.micro,
      SophiaMotion.short,
      SophiaMotion.medium,
      SophiaMotion.long,
      SophiaMotion.stagger,
    ]) {
      expect(SophiaMotion.resolve(context, duration), Duration.zero);
    }
  });

  testWidgets('lateral navigation gives outgoing and incoming distinct depth', (
    tester,
  ) async {
    var index = 0;
    late StateSetter rebuild;
    await tester.pumpWidget(
      MaterialApp(
        home: StatefulBuilder(
          builder: (context, setState) {
            rebuild = setState;
            return LateralBranchContainer(
              currentIndex: index,
              children: const [Text('A'), Text('B')],
            );
          },
        ),
      ),
    );

    rebuild(() => index = 1);
    await tester.pump();

    final outgoing = tester.widget<AnimatedScale>(
      find.ancestor(of: find.text('A'), matching: find.byType(AnimatedScale)),
    );
    final incoming = tester.widget<AnimatedScale>(
      find.ancestor(of: find.text('B'), matching: find.byType(AnimatedScale)),
    );
    expect(outgoing.scale, SophiaMotion.lateralOutgoingScale);
    expect(incoming.scale, 1);
  });
}
