import 'package:equatable/equatable.dart';
import 'package:sophia_ai/features/insight/presentation/cubit/insight_cubit.dart';

// State
class InsightState extends Equatable {
  final double accuracy; // 0.92
  final List<DecisionItem> pendingDecisions;

  const InsightState({required this.accuracy, required this.pendingDecisions});
  @override
  List<Object> get props => [accuracy, pendingDecisions];
}
