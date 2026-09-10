import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/beliefs_repository.dart';
import '../../domain/models.dart';

class BeliefsState extends Equatable {
  const BeliefsState({
    this.beliefs = const [],
    this.loading = false,
    this.saving = false,
    this.error,
    this.scope,
    this.scopeKey,
  });
  final List<Belief> beliefs;
  final bool loading, saving;
  final String? error, scope, scopeKey;
  @override
  List<Object?> get props => [beliefs, loading, saving, error, scope, scopeKey];
}

class BeliefsCubit extends Cubit<BeliefsState> {
  BeliefsCubit(this._repository) : super(const BeliefsState());
  final BeliefsRepository _repository;
  Future<void> load({String? scope, String? scopeKey}) async {
    emit(
      BeliefsState(
        beliefs: state.beliefs,
        loading: true,
        scope: scope,
        scopeKey: scopeKey,
      ),
    );
    try {
      emit(
        BeliefsState(
          beliefs: await _repository.list(scope: scope, scopeKey: scopeKey),
          scope: scope,
          scopeKey: scopeKey,
        ),
      );
    } catch (error) {
      emit(
        BeliefsState(
          beliefs: state.beliefs,
          error: '$error',
          scope: scope,
          scopeKey: scopeKey,
        ),
      );
    }
  }

  Future<void> retire(String id) async {
    await _repository.retire(id);
    await load(scope: state.scope, scopeKey: state.scopeKey);
  }

  Future<void> correct(String id, String statement) async {
    await _repository.updateStatement(id, statement);
    await load(scope: state.scope, scopeKey: state.scopeKey);
  }
}
