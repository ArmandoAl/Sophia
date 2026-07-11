import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/features/insight/presentation/cubit/insight_state.dart';

class DecisionItem {
  final String id;
  final String title;
  final String timestamp;
  final String reasoning;
  final String iconEmoji; // Usaremos emojis o IconData

  const DecisionItem(
    this.id,
    this.title,
    this.timestamp,
    this.reasoning,
    this.iconEmoji,
  );
}

// Cubit
class InsightCubit extends Cubit<InsightState> {
  InsightCubit()
    : super(const InsightState(accuracy: 0.92, pendingDecisions: [])) {
    loadInsight();
  }

  void loadInsight() {
    // Simular carga de datos
    emit(
      InsightState(
        accuracy: 0.92,
        pendingDecisions: [
          const DecisionItem(
            '1',
            'Turned on Jazz Music',
            'Tuesday at 8:15 AM',
            'Based on your morning routine, I started some relaxing jazz to help you focus.',
            '🎵',
          ),
        ],
      ),
    );
  }

  void approveDecision(String id) {
    // Lógica para aprobar y quitar de la lista
    final newList = List<DecisionItem>.from(state.pendingDecisions)
      ..removeWhere((e) => e.id == id);
    emit(InsightState(accuracy: state.accuracy, pendingDecisions: newList));
  }
}
