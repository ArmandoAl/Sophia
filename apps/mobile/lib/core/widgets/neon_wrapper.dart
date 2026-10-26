import 'package:flutter/material.dart';

import '../theme/design_tokens.dart';

/// Legacy layout wrapper kept for API compatibility.
class NeonWrapper extends StatelessWidget {
  const NeonWrapper({super.key, required this.child});

  final Widget child;

  @override
  Widget build(BuildContext context) =>
      ColoredBox(color: context.colors.surface, child: child);
}
