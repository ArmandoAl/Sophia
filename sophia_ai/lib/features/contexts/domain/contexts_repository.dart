import 'models.dart';

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
}
