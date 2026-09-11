import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../beliefs/domain/beliefs_repository.dart';
import '../../../beliefs/domain/models.dart';
import '../../../contexts/domain/contexts_repository.dart';
import '../../../contexts/domain/models.dart';

class EntitiesPage extends StatefulWidget {
  const EntitiesPage({super.key});

  @override
  State<EntitiesPage> createState() => _EntitiesPageState();
}

class _EntitiesPageState extends State<EntitiesPage> {
  final _repository = sl<ContextsRepository>();
  late Future<List<UserContext>> _entities = _repository.listEntities();

  void _reload() => setState(() => _entities = _repository.listEntities());

  Future<void> _run(Future<void> Function() operation) async {
    try {
      await operation();
      if (mounted) _reload();
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('No se pudo guardar: $error')));
      }
    }
  }

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Personas y grupos')),
    floatingActionButton: FloatingActionButton.extended(
      onPressed: () async {
        final draft = await _showEntityEditor(context);
        if (draft == null) return;
        await _run(
          () async => _repository.createEntity(
            kind: draft.kind,
            slug: _slugify(draft.label),
            label: draft.label,
            relationship: draft.relationship,
            aliases: draft.aliases,
          ),
        );
      },
      icon: const Icon(Icons.person_add_alt_1_outlined),
      label: const Text('Añadir'),
    ),
    body: FutureBuilder<List<UserContext>>(
      future: _entities,
      builder: (context, snapshot) {
        if (snapshot.connectionState == ConnectionState.waiting) {
          return const Center(child: CircularProgressIndicator());
        }
        if (snapshot.hasError) {
          return _LoadError(onRetry: _reload);
        }
        final entities =
            (snapshot.data ?? const <UserContext>[])
                .where(
                  (item) =>
                      (item.kind == 'person' || item.kind == 'group') &&
                      item.status != 'archived' &&
                      item.status != 'merged',
                )
                .toList()
              ..sort((a, b) {
                if (a.status != b.status) {
                  return a.status == 'pending_review' ? -1 : 1;
                }
                return a.label.toLowerCase().compareTo(b.label.toLowerCase());
              });
        if (entities.isEmpty) {
          return const _EmptyEntities();
        }
        final groups = <String, List<UserContext>>{};
        for (final entity in entities) {
          final relationship = entity.relationship.trim();
          (groups[relationship.isEmpty ? 'Otros' : relationship] ??= []).add(
            entity,
          );
        }
        final relationships = groups.keys.toList()
          ..sort((a, b) => a.toLowerCase().compareTo(b.toLowerCase()));
        return RefreshIndicator(
          onRefresh: () async {
            _reload();
            await _entities;
          },
          child: ListView(
            padding: const EdgeInsets.fromLTRB(
              SophiaSpace.lg,
              SophiaSpace.md,
              SophiaSpace.lg,
              96,
            ),
            children: [
              Text(
                'Tus notas sobre la gente que forma parte de tu vida.',
                style: TextStyle(color: context.colors.softInk),
              ),
              const SizedBox(height: SophiaSpace.lg),
              for (final relationship in relationships) ...[
                Text(
                  _capitalize(relationship),
                  style: Theme.of(context).textTheme.titleMedium,
                ),
                const SizedBox(height: SophiaSpace.sm),
                for (final entity in groups[relationship]!)
                  _EntityCard(
                    entity: entity,
                    onConfirm: () => _run(
                      () async => _repository.updateEntity(entity.id, {
                        'status': 'active',
                      }),
                    ),
                    onRename: () async {
                      final draft = await _showEntityEditor(context, entity);
                      if (draft == null) return;
                      await _run(
                        () async => _repository.updateEntity(entity.id, {
                          'label': draft.label,
                          'relationship': draft.relationship,
                          'aliases': draft.aliases,
                          'kind': draft.kind,
                        }),
                      );
                    },
                    onMerge: () async {
                      final target = await _pickMergeTarget(
                        context,
                        entities
                            .where(
                              (item) =>
                                  item.id != entity.id &&
                                  item.kind == entity.kind,
                            )
                            .toList(),
                      );
                      if (target != null) {
                        await _run(
                          () async =>
                              _repository.mergeEntity(entity.id, target.id),
                        );
                      }
                    },
                    onArchive: () =>
                        _run(() => _repository.archiveEntity(entity.id)),
                  ),
                const SizedBox(height: SophiaSpace.md),
              ],
            ],
          ),
        );
      },
    ),
  );
}

class _EntityCard extends StatelessWidget {
  const _EntityCard({
    required this.entity,
    required this.onConfirm,
    required this.onRename,
    required this.onMerge,
    required this.onArchive,
  });

  final UserContext entity;
  final VoidCallback onConfirm, onRename, onMerge, onArchive;

  @override
  Widget build(BuildContext context) {
    final pending = entity.status == 'pending_review';
    return Card(
      margin: const EdgeInsets.only(bottom: SophiaSpace.sm),
      shape: RoundedRectangleBorder(
        side: BorderSide(
          color: pending ? context.colors.accent : context.colors.line,
        ),
        borderRadius: BorderRadius.circular(SophiaRadius.card),
      ),
      child: InkWell(
        borderRadius: BorderRadius.circular(SophiaRadius.card),
        onTap: () => context.push('/entities/${entity.id}'),
        child: Padding(
          padding: const EdgeInsets.all(SophiaSpace.md),
          child: Column(
            children: [
              Row(
                children: [
                  CircleAvatar(
                    child: Icon(
                      entity.kind == 'group'
                          ? Icons.groups_outlined
                          : Icons.person_outline,
                    ),
                  ),
                  const SizedBox(width: SophiaSpace.md),
                  Expanded(
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          entity.label,
                          style: Theme.of(context).textTheme.titleMedium,
                        ),
                        if (entity.aliases.isNotEmpty)
                          Text(
                            entity.aliases.join(', '),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                            style: Theme.of(context).textTheme.bodySmall,
                          ),
                      ],
                    ),
                  ),
                  if (pending)
                    Chip(
                      avatar: const Icon(Icons.auto_awesome, size: 16),
                      label: const Text('Por revisar'),
                      visualDensity: VisualDensity.compact,
                    )
                  else
                    const Icon(Icons.chevron_right),
                ],
              ),
              if (pending) ...[
                const Divider(height: SophiaSpace.lg),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    _QuickAction(
                      icon: Icons.check,
                      label: 'Confirmar',
                      onTap: onConfirm,
                    ),
                    _QuickAction(
                      icon: Icons.edit_outlined,
                      label: 'Renombrar',
                      onTap: onRename,
                    ),
                    _QuickAction(
                      icon: Icons.merge_outlined,
                      label: 'Fusionar',
                      onTap: onMerge,
                    ),
                    _QuickAction(
                      icon: Icons.archive_outlined,
                      label: 'Archivar',
                      onTap: onArchive,
                    ),
                  ],
                ),
              ],
            ],
          ),
        ),
      ),
    );
  }
}

class _QuickAction extends StatelessWidget {
  const _QuickAction({
    required this.icon,
    required this.label,
    required this.onTap,
  });
  final IconData icon;
  final String label;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => Expanded(
    child: Semantics(
      button: true,
      label: label,
      child: InkWell(
        onTap: onTap,
        child: Padding(
          padding: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
          child: Column(
            children: [
              Icon(icon, size: 20),
              const SizedBox(height: SophiaSpace.xxs),
              Text(label, style: Theme.of(context).textTheme.labelSmall),
            ],
          ),
        ),
      ),
    ),
  );
}

class EntityDetailPage extends StatefulWidget {
  const EntityDetailPage({super.key, required this.id});
  final String id;

  @override
  State<EntityDetailPage> createState() => _EntityDetailPageState();
}

class _EntityDetailPageState extends State<EntityDetailPage> {
  final _entities = sl<ContextsRepository>();
  final _beliefs = sl<BeliefsRepository>();
  late Future<EntityDetails> _details = _entities.getEntity(widget.id);

  void _reload() => setState(() => _details = _entities.getEntity(widget.id));

  Future<void> _run(Future<void> Function() operation) async {
    try {
      await operation();
      if (mounted) _reload();
    } catch (error) {
      if (mounted) {
        ScaffoldMessenger.of(
          context,
        ).showSnackBar(SnackBar(content: Text('No se pudo guardar: $error')));
      }
    }
  }

  @override
  Widget build(BuildContext context) => FutureBuilder<EntityDetails>(
    future: _details,
    builder: (context, snapshot) {
      final details = snapshot.data;
      return Scaffold(
        appBar: AppBar(
          title: Text(
            details == null
                ? 'Lo que sé'
                : 'Lo que sé de ${details.entity.label}',
          ),
          actions: [
            if (details != null)
              IconButton(
                tooltip: 'Editar estas notas',
                icon: const Icon(Icons.edit_outlined),
                onPressed: () async {
                  final draft = await _showEntityEditor(
                    context,
                    details.entity,
                  );
                  if (draft == null) return;
                  await _run(
                    () async => _entities.updateEntity(widget.id, {
                      'label': draft.label,
                      'relationship': draft.relationship,
                      'aliases': draft.aliases,
                      'kind': draft.kind,
                    }),
                  );
                },
              ),
          ],
        ),
        body: _detailBody(snapshot),
      );
    },
  );

  Widget _detailBody(AsyncSnapshot<EntityDetails> snapshot) {
    if (snapshot.connectionState == ConnectionState.waiting &&
        !snapshot.hasData) {
      return const Center(child: CircularProgressIndicator());
    }
    if (snapshot.hasError || !snapshot.hasData) {
      return _LoadError(onRetry: _reload);
    }
    final details = snapshot.data!;
    final visibleFacts = details.facts
        .where((item) => item.status != 'archived' && item.status != 'retired')
        .toList();
    final states = visibleFacts
        .where((item) => item.factKind == 'state')
        .toList();
    final traits = visibleFacts
        .where((item) => item.factKind == 'trait')
        .toList();
    final relationships = visibleFacts
        .where((item) => item.factKind == 'relationship')
        .toList();
    final userBeliefs = details.userBeliefs
        .where((item) => item.status != 'archived' && item.status != 'retired')
        .toList();
    final episodes = [...details.episodes]
      ..sort((a, b) => b.occurredAt.compareTo(a.occurredAt));
    return RefreshIndicator(
      onRefresh: () async {
        _reload();
        await _details;
      },
      child: ListView(
        padding: const EdgeInsets.all(SophiaSpace.lg),
        children: [
          _EntitySummary(entity: details.entity),
          const SizedBox(height: SophiaSpace.md),
          SwitchListTile.adaptive(
            contentPadding: EdgeInsets.zero,
            title: const Text('Silenciar hilos abiertos'),
            subtitle: Text(
              'Sofía no retomará temas pendientes sobre ${details.entity.label} por iniciativa propia.',
            ),
            value: details.entity.threadsMuted,
            onChanged: (value) => _run(
              () async =>
                  _entities.updateEntity(widget.id, {'threads_muted': value}),
            ),
          ),
          _FactSection(
            title: 'Lo que le pasa ahora',
            empty: 'No hay nada temporal anotado.',
            facts: states,
            validityVisible: true,
            onEdit: _editFact,
            onRetire: _retireFact,
          ),
          _FactSection(
            title: 'Cómo es',
            empty: 'Todavía no hay gustos o rasgos anotados.',
            facts: traits,
            onEdit: _editFact,
            onRetire: _retireFact,
          ),
          if (relationships.isNotEmpty)
            _FactSection(
              title: 'El vínculo entre ustedes',
              empty: '',
              facts: relationships,
              onEdit: _editFact,
              onRetire: _retireFact,
            ),
          _FactSection(
            title: 'Cómo eres tú con ${details.entity.label}',
            empty:
                'Sofía todavía no ha anotado nada sobre cómo eres con esta persona.',
            facts: userBeliefs,
            onEdit: _editFact,
            onRetire: _retireFact,
          ),
          const SizedBox(height: SophiaSpace.lg),
          Text(
            'Momentos que recuerdas',
            style: Theme.of(context).textTheme.titleLarge,
          ),
          const SizedBox(height: SophiaSpace.sm),
          if (episodes.isEmpty)
            Text(
              'Todavía no hay conversaciones concretas para recordar aquí.',
              style: TextStyle(color: context.colors.softInk),
            )
          else
            for (final episode in episodes)
              ListTile(
                contentPadding: EdgeInsets.zero,
                leading: const Icon(Icons.history),
                title: Text(episode.summary),
                subtitle: Text(_date(episode.occurredAt)),
              ),
          const SizedBox(height: SophiaSpace.xl),
        ],
      ),
    );
  }

  Future<void> _editFact(Belief fact) async {
    final controller = TextEditingController(text: fact.statement);
    final value = await showDialog<String>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Corregir nota'),
        content: TextField(
          controller: controller,
          autofocus: true,
          maxLines: 3,
          decoration: const InputDecoration(
            labelText: 'Lo que quieres recordar',
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () =>
                Navigator.pop(dialogContext, controller.text.trim()),
            child: const Text('Guardar'),
          ),
        ],
      ),
    );
    controller.dispose();
    if (value != null && value.isNotEmpty && value != fact.statement) {
      await _run(() async => _beliefs.updateStatement(fact.id, value));
    }
  }

  Future<void> _retireFact(Belief fact) async {
    final retire =
        await showDialog<bool>(
          context: context,
          builder: (dialogContext) => AlertDialog(
            title: const Text('Retirar esta nota'),
            content: const Text(
              'Sofía dejará de usarla, pero conservará el historial de la corrección.',
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(dialogContext, false),
                child: const Text('Conservar'),
              ),
              FilledButton(
                onPressed: () => Navigator.pop(dialogContext, true),
                child: const Text('Retirar'),
              ),
            ],
          ),
        ) ??
        false;
    if (retire) await _run(() => _beliefs.retire(fact.id));
  }
}

class _EntitySummary extends StatelessWidget {
  const _EntitySummary({required this.entity});
  final UserContext entity;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(SophiaSpace.md),
    decoration: BoxDecoration(
      color: context.colors.elevated,
      border: Border.all(color: context.colors.line),
      borderRadius: BorderRadius.circular(SophiaRadius.card),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(entity.label, style: Theme.of(context).textTheme.headlineSmall),
        if (entity.relationship.isNotEmpty)
          Text(_capitalize(entity.relationship)),
        if (entity.aliases.isNotEmpty) ...[
          const SizedBox(height: SophiaSpace.xs),
          Text(
            'También le dices ${entity.aliases.join(', ')}',
            style: TextStyle(color: context.colors.softInk),
          ),
        ],
        if (entity.status == 'pending_review') ...[
          const SizedBox(height: SophiaSpace.sm),
          const Chip(label: Text('Sofía sugirió esta persona · por revisar')),
        ],
      ],
    ),
  );
}

class _FactSection extends StatelessWidget {
  const _FactSection({
    required this.title,
    required this.empty,
    required this.facts,
    required this.onEdit,
    required this.onRetire,
    this.validityVisible = false,
  });

  final String title, empty;
  final List<Belief> facts;
  final bool validityVisible;
  final ValueChanged<Belief> onEdit, onRetire;

  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.only(top: SophiaSpace.lg),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text(title, style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: SophiaSpace.sm),
        if (facts.isEmpty)
          Text(empty, style: TextStyle(color: context.colors.softInk))
        else
          for (final fact in facts)
            Card(
              margin: const EdgeInsets.only(bottom: SophiaSpace.xs),
              child: ListTile(
                title: Text(fact.statement),
                subtitle: validityVisible && fact.validUntil != null
                    ? Text('Vigente hasta ${_date(fact.validUntil!)}')
                    : null,
                trailing: PopupMenuButton<String>(
                  tooltip: 'Opciones de esta nota',
                  onSelected: (value) =>
                      value == 'edit' ? onEdit(fact) : onRetire(fact),
                  itemBuilder: (_) => const [
                    PopupMenuItem(value: 'edit', child: Text('Corregir')),
                    PopupMenuItem(value: 'retire', child: Text('Retirar')),
                  ],
                ),
              ),
            ),
      ],
    ),
  );
}

class _LoadError extends StatelessWidget {
  const _LoadError({required this.onRetry});
  final VoidCallback onRetry;

  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        const Text('No pudimos cargar estas notas.'),
        TextButton(onPressed: onRetry, child: const Text('Reintentar')),
      ],
    ),
  );
}

class _EmptyEntities extends StatelessWidget {
  const _EmptyEntities();

  @override
  Widget build(BuildContext context) => Center(
    child: Padding(
      padding: const EdgeInsets.all(SophiaSpace.xl),
      child: Text(
        'Cuando hables varias veces de alguien, aparecerá aquí para que puedas revisar lo que Sofía entendió.',
        textAlign: TextAlign.center,
        style: TextStyle(color: context.colors.softInk),
      ),
    ),
  );
}

class _EntityDraft {
  const _EntityDraft({
    required this.label,
    required this.relationship,
    required this.aliases,
    required this.kind,
  });
  final String label, relationship, kind;
  final List<String> aliases;
}

Future<_EntityDraft?> _showEntityEditor(
  BuildContext context, [
  UserContext? entity,
]) async {
  final label = TextEditingController(text: entity?.label);
  final relationship = TextEditingController(text: entity?.relationship);
  final aliases = TextEditingController(text: entity?.aliases.join(', '));
  var kind = entity?.kind ?? 'person';
  final result = await showDialog<_EntityDraft>(
    context: context,
    builder: (dialogContext) => StatefulBuilder(
      builder: (context, setState) => AlertDialog(
        title: Text(entity == null ? 'Añadir a tus notas' : 'Editar tus notas'),
        content: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              SegmentedButton<String>(
                segments: const [
                  ButtonSegment(
                    value: 'person',
                    icon: Icon(Icons.person_outline),
                    label: Text('Persona'),
                  ),
                  ButtonSegment(
                    value: 'group',
                    icon: Icon(Icons.groups_outlined),
                    label: Text('Grupo'),
                  ),
                ],
                selected: {kind},
                onSelectionChanged: (value) =>
                    setState(() => kind = value.first),
              ),
              const SizedBox(height: SophiaSpace.md),
              TextField(
                controller: label,
                autofocus: true,
                decoration: const InputDecoration(labelText: 'Nombre'),
              ),
              const SizedBox(height: SophiaSpace.md),
              TextField(
                controller: relationship,
                decoration: const InputDecoration(
                  labelText: 'Qué relación tiene contigo',
                  hintText: 'Hermana, amiga, primos…',
                ),
              ),
              const SizedBox(height: SophiaSpace.md),
              TextField(
                controller: aliases,
                decoration: const InputDecoration(
                  labelText: 'Otros nombres',
                  helperText: 'Sepáralos con comas.',
                ),
              ),
            ],
          ),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(dialogContext),
            child: const Text('Cancelar'),
          ),
          FilledButton(
            onPressed: () {
              final name = label.text.trim();
              if (name.isEmpty) return;
              Navigator.pop(
                dialogContext,
                _EntityDraft(
                  label: name,
                  relationship: relationship.text.trim(),
                  aliases: aliases.text
                      .split(',')
                      .map((value) => value.trim())
                      .where((value) => value.isNotEmpty)
                      .toList(),
                  kind: kind,
                ),
              );
            },
            child: const Text('Guardar'),
          ),
        ],
      ),
    ),
  );
  label.dispose();
  relationship.dispose();
  aliases.dispose();
  return result;
}

Future<UserContext?> _pickMergeTarget(
  BuildContext context,
  List<UserContext> candidates,
) => showModalBottomSheet<UserContext>(
  context: context,
  showDragHandle: true,
  builder: (context) => SafeArea(
    child: ListView(
      shrinkWrap: true,
      children: [
        const ListTile(
          title: Text('Fusionar con…'),
          subtitle: Text(
            'Todos los recuerdos pasarán a la persona que elijas.',
          ),
        ),
        for (final candidate in candidates)
          ListTile(
            leading: Icon(
              candidate.kind == 'group'
                  ? Icons.groups_outlined
                  : Icons.person_outline,
            ),
            title: Text(candidate.label),
            subtitle: candidate.relationship.isEmpty
                ? null
                : Text(candidate.relationship),
            onTap: () => Navigator.pop(context, candidate),
          ),
        if (candidates.isEmpty)
          const ListTile(
            title: Text('No hay otra persona con quien fusionar.'),
          ),
      ],
    ),
  ),
);

String _slugify(String value) {
  const accents = 'áéíóúüñ';
  const plain = 'aeiouun';
  var normalized = value.toLowerCase();
  for (var index = 0; index < accents.length; index++) {
    normalized = normalized.replaceAll(accents[index], plain[index]);
  }
  return normalized
      .replaceAll(RegExp(r'[^a-z0-9]+'), '-')
      .replaceAll(RegExp(r'^-+|-+$'), '');
}

String _date(DateTime value) =>
    '${value.day.toString().padLeft(2, '0')}/${value.month.toString().padLeft(2, '0')}/${value.year}';

String _capitalize(String value) => value.isEmpty
    ? value
    : '${value.substring(0, 1).toUpperCase()}${value.substring(1)}';
