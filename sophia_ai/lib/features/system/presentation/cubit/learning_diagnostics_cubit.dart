import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../../beliefs/domain/beliefs_repository.dart';
import '../../../beliefs/domain/models.dart';

class LearningDiagnosticsState extends Equatable {
  const LearningDiagnosticsState({
    this.prompt,
    this.summaries = const [],
    this.loading = false,
    this.error,
  });
  final PromptVersion? prompt;
  final List<LearningSummary> summaries;
  final bool loading;
  final String? error;
  @override
  List<Object?> get props => [prompt, summaries, loading, error];
}

class LearningDiagnosticsCubit extends Cubit<LearningDiagnosticsState> {
  LearningDiagnosticsCubit(this._repository)
    : super(const LearningDiagnosticsState());
  final BeliefsRepository _repository;
  Future<void> load() async {
    emit(const LearningDiagnosticsState(loading: true));
    try {
      final results = await Future.wait([
        _repository.getPromptVersion(),
        _repository.listSummaries(),
      ]);
      emit(
        LearningDiagnosticsState(
          prompt: results[0] as PromptVersion?,
          summaries: results[1] as List<LearningSummary>,
        ),
      );
    } catch (error) {
      emit(LearningDiagnosticsState(error: error.toString()));
    }
  }
}
