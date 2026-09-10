import 'package:equatable/equatable.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import '../../domain/contexts_repository.dart';
import '../../domain/models.dart';

class ContextsState extends Equatable {
  const ContextsState({
    this.contexts = const [],
    this.loading = false,
    this.error,
  });
  final List<UserContext> contexts;
  final bool loading;
  final String? error;
  @override
  List<Object?> get props => [contexts, loading, error];
}

class ContextsCubit extends Cubit<ContextsState> {
  ContextsCubit(this._repository) : super(const ContextsState());
  final ContextsRepository _repository;
  Future<void> load() async {
    emit(ContextsState(contexts: state.contexts, loading: true));
    try {
      emit(ContextsState(contexts: await _repository.list()));
    } catch (error) {
      emit(ContextsState(contexts: state.contexts, error: '$error'));
    }
  }

  Future<void> create({
    required String kind,
    required String slug,
    required String label,
    required List<String> aliases,
  }) async {
    await _repository.create(
      kind: kind,
      slug: slug,
      label: label,
      aliases: aliases,
    );
    await load();
  }

  Future<void> update(
    UserContext value,
    String label,
    List<String> aliases,
  ) async {
    await _repository.update(value.id, label: label, aliases: aliases);
    await load();
  }

  Future<void> archive(String id) async {
    await _repository.archive(id);
    await load();
  }
}
