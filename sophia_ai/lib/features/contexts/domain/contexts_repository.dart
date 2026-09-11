import 'models.dart';
import '../../beliefs/domain/models.dart';

class EntityDetails {
  const EntityDetails({
    required this.entity,
    required this.facts,
    required this.userBeliefs,
    required this.episodes,
  });

  final UserContext entity;
  final List<Belief> facts, userBeliefs;
  final List<Episode> episodes;
}

abstract interface class ContextsRepository {
  Future<List<UserContext>> list();
  Future<UserContext> create({
    required String kind,
    required String slug,
    required String label,
    required List<String> aliases,
  });
  Future<UserContext> update(
    String id, {
    required String label,
    required List<String> aliases,
  });
  Future<void> archive(String id);

  Future<List<UserContext>> listEntities({String? kind, String? status});
  Future<UserContext> createEntity({
    required String kind,
    required String slug,
    required String label,
    required String relationship,
    required List<String> aliases,
  });
  Future<EntityDetails> getEntity(String id);
  Future<UserContext> updateEntity(String id, Map<String, dynamic> changes);
  Future<UserContext> mergeEntity(String id, String into);
  Future<void> archiveEntity(String id);
}
