import 'package:flutter/material.dart';
import 'package:flutter_test/flutter_test.dart';
import 'package:sophia_ai/core/theme/app_theme.dart';
import 'package:sophia_ai/features/beliefs/domain/models.dart';
import 'package:sophia_ai/features/system/presentation/pages/diagnostics_screen.dart';

void main() {
  testWidgets('learning chart distinguishes a worker run from silent days', (
    tester,
  ) async {
    final today = DateUtils.dateOnly(DateTime.now());
    final month = today.month.toString().padLeft(2, '0');
    final day = today.day.toString().padLeft(2, '0');
    final key = '${today.year}-$month-$day';

    await tester.pumpWidget(
      MaterialApp(
        theme: AppTheme.lightTheme,
        home: Scaffold(
          body: LearningWeekChart(
            summaries: [
              LearningSummary(
                date: key,
                approved: 1,
                corrected: 0,
                rejected: 0,
                inputTokens: 20,
                outputTokens: 10,
              ),
            ],
          ),
        ),
      ),
    );

    expect(
      find.byWidgetPredicate(
        (widget) =>
            widget is Tooltip &&
            widget.message == '$key: 30 tokens; 1 decisión revisada',
      ),
      findsOneWidget,
    );
    expect(
      find.byWidgetPredicate(
        (widget) =>
            widget is Tooltip &&
            (widget.message?.endsWith('el worker no registró una ejecución') ??
                false),
      ),
      findsNWidgets(6),
    );
  });
}
