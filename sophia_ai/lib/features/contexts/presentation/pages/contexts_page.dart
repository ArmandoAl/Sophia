import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/theme/motion.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../domain/models.dart';
import '../cubit/contexts_cubit.dart';

class ContextsPage extends StatelessWidget {
  const ContextsPage({super.key});

  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) => sl<ContextsCubit>()..load(),
    child: Builder(
      builder: (context) => Scaffold(
        appBar: AppBar(
          title: const Text('Contextos'),
          actions: [
            Hero(
              tag: 'new-context',
              transitionOnUserGestures: true,
              child: Material(
                color: context.colors.surface.withValues(alpha: 0),
                child: TactileButton(
                  semanticLabel: 'Crear contexto',
                  onPressed: () => _showEditor(context),
                  child: const Padding(
                    padding: EdgeInsets.all(SophiaSpace.sm),
                    child: Icon(Icons.add),
                  ),
                ),
              ),
            ),
            const SizedBox(width: SophiaSpace.xs),
          ],
        ),
        body: BlocBuilder<ContextsCubit, ContextsState>(
          builder: (context, state) {
            if (state.loading && state.contexts.isEmpty) {
              return const MotionSwap(
                child: Padding(
                  key: ValueKey('contexts-loading'),
                  padding: EdgeInsets.all(SophiaSpace.lg),
                  child: Column(
                    children: [
                      ContentSkeleton(),
                      SizedBox(height: SophiaSpace.md),
                      ContentSkeleton(),
                    ],
                  ),
                ),
              );
            }
            final active = state.contexts.where((item) => item.active).toList();
            return MotionSwap(
              child: ListView(
                key: const ValueKey('contexts-content'),
                padding: const EdgeInsets.all(SophiaSpace.lg),
                children: [
                  Text(
                    'Ayudan a Sofía a entender cuándo eres tú con una persona y cuándo estás en un modo concreto.',
                    style: TextStyle(color: context.colors.softInk),
                  ),
                  if (state.error != null)
                    Padding(
                      padding: const EdgeInsets.only(top: SophiaSpace.sm),
                      child: Text(
                        state.error!,
                        style: TextStyle(color: context.colors.critical),
                      ),
                    ),
                  const SizedBox(height: SophiaSpace.lg),
                  if (active.isEmpty && !state.loading)
                    Text(
                      'Todavía no has creado contextos.',
                      style: TextStyle(color: context.colors.softInk),
                    ),
                  for (var index = 0; index < active.length; index++)
                    StaggeredEntry(
                      index: index,
                      child: _ContextCard(item: active[index]),
                    ),
                ],
              ),
            );
          },
        ),
      ),
    ),
  );
}

class _ContextCard extends StatelessWidget {
  const _ContextCard({required this.item});

  final UserContext item;

  @override
  Widget build(BuildContext context) {
    final person = item.kind == 'person';
    final tone = person ? context.colors.softInk : context.colors.accent;
    return Container(
      margin: const EdgeInsets.only(bottom: SophiaSpace.sm),
      padding: const EdgeInsets.all(SophiaSpace.md),
      decoration: BoxDecoration(
        color: context.colors.elevated,
        border: Border.all(color: context.colors.line),
        borderRadius: BorderRadius.circular(SophiaRadius.card),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Container(
                padding: const EdgeInsets.symmetric(
                  horizontal: SophiaSpace.xs,
                  vertical: SophiaSpace.xxs,
                ),
                decoration: BoxDecoration(
                  color: tone.withValues(alpha: SophiaOpacity.subtle),
                  borderRadius: BorderRadius.circular(SophiaRadius.control),
                ),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      person ? Icons.person_outline : Icons.layers_outlined,
                      size: SophiaSpace.md,
                      color: tone,
                    ),
                    const SizedBox(width: SophiaSpace.xxs),
                    Text(
                      person ? 'PERSONA' : 'MODO',
                      style: Theme.of(
                        context,
                      ).textTheme.labelMedium?.copyWith(color: tone),
                    ),
                  ],
                ),
              ),
              const Spacer(),
              Text(
                item.slug,
                overflow: TextOverflow.ellipsis,
                style: Theme.of(context).textTheme.labelMedium,
              ),
            ],
          ),
          const SizedBox(height: SophiaSpace.md),
          Hero(
            tag: 'context-${item.id}',
            transitionOnUserGestures: true,
            child: Material(
              color: context.colors.surface.withValues(alpha: 0),
              child: Text(
                item.label,
                style: Theme.of(context).textTheme.titleMedium,
              ),
            ),
          ),
          if (item.aliases.isNotEmpty) ...[
            const SizedBox(height: SophiaSpace.xs),
            Text(
              'También: ${item.aliases.join(', ')}',
              style: Theme.of(context).textTheme.bodySmall,
            ),
          ],
          const SizedBox(height: SophiaSpace.sm),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              TactileButton(
                semanticLabel: 'Archivar ${item.label}',
                onPressed: () => _archive(context, item),
                child: const Padding(
                  padding: EdgeInsets.all(SophiaSpace.sm),
                  child: Text('Archivar'),
                ),
              ),
              TactileButton(
                semanticLabel: 'Editar ${item.label}',
                onPressed: () => _showEditor(context, item),
                child: Padding(
                  padding: const EdgeInsets.all(SophiaSpace.sm),
                  child: Text(
                    'Editar',
                    style: TextStyle(color: context.colors.accent),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

Future<void> _archive(BuildContext context, UserContext item) async {
  final cubit = context.read<ContextsCubit>();
  final confirmed =
      await showSophiaSheet<bool>(
        context: context,
        builder: (dialogContext) => Padding(
          padding: const EdgeInsets.all(SophiaSpace.lg),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                'Archivar ${item.label}',
                style: Theme.of(context).textTheme.titleLarge,
              ),
              const SizedBox(height: SophiaSpace.sm),
              const Text(
                'Sofía dejará de usarlo, pero conservará su historial.',
              ),
              const SizedBox(height: SophiaSpace.lg),
              Row(
                mainAxisAlignment: MainAxisAlignment.end,
                children: [
                  TactileButton(
                    onPressed: () => Navigator.pop(dialogContext, false),
                    child: const Padding(
                      padding: EdgeInsets.all(SophiaSpace.sm),
                      child: Text('Conservar'),
                    ),
                  ),
                  TactileButton(
                    onPressed: () => Navigator.pop(dialogContext, true),
                    child: Padding(
                      padding: const EdgeInsets.all(SophiaSpace.sm),
                      child: Text(
                        'Archivar',
                        style: TextStyle(color: context.colors.accent),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ) ??
      false;
  if (confirmed) await cubit.archive(item.id);
}

Future<void> _showEditor(BuildContext context, [UserContext? existing]) async {
  final cubit = context.read<ContextsCubit>();
  final label = TextEditingController(text: existing?.label);
  final slug = TextEditingController(text: existing?.slug);
  final aliases = TextEditingController(text: existing?.aliases.join(', '));
  var kind = existing?.kind ?? 'person';
  var slugEdited = existing != null;
  final save =
      await showSophiaSheet<bool>(
        context: context,
        builder: (dialogContext) => StatefulBuilder(
          builder: (context, setState) => Padding(
            padding: EdgeInsets.fromLTRB(
              SophiaSpace.lg,
              SophiaSpace.md,
              SophiaSpace.lg,
              MediaQuery.viewInsetsOf(context).bottom + SophiaSpace.lg,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Hero(
                  tag: existing == null
                      ? 'new-context'
                      : 'context-${existing.id}',
                  transitionOnUserGestures: true,
                  child: Material(
                    color: context.colors.surface.withValues(alpha: 0),
                    child: Text(
                      existing == null ? 'Nuevo contexto' : 'Editar contexto',
                      style: Theme.of(context).textTheme.titleLarge,
                    ),
                  ),
                ),
                const SizedBox(height: SophiaSpace.sm),
                Text(
                  'Dale un nombre que te resulte natural. Sofía usará el contexto para interpretar mejor lo que le cuentas.',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
                const SizedBox(height: SophiaSpace.lg),
                Flexible(
                  child: SingleChildScrollView(
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        if (existing == null) ...[
                          _KindSelector(
                            value: kind,
                            onChanged: (value) => setState(() => kind = value),
                          ),
                          const SizedBox(height: SophiaSpace.md),
                        ],
                        TextField(
                          controller: label,
                          autofocus: true,
                          textInputAction: TextInputAction.next,
                          decoration: const InputDecoration(
                            labelText: 'Nombre visible',
                          ),
                          onChanged: existing == null
                              ? (value) {
                                  if (!slugEdited) slug.text = _slugify(value);
                                }
                              : null,
                        ),
                        if (existing == null) ...[
                          const SizedBox(height: SophiaSpace.md),
                          TextField(
                            controller: slug,
                            textInputAction: TextInputAction.next,
                            decoration: const InputDecoration(
                              labelText: 'Identificador',
                              helperText:
                                  'Se genera automáticamente, pero puedes cambiarlo.',
                            ),
                            onChanged: (_) => slugEdited = true,
                          ),
                        ],
                        const SizedBox(height: SophiaSpace.md),
                        TextField(
                          controller: aliases,
                          decoration: const InputDecoration(
                            labelText: 'Otros nombres (opcional)',
                            helperText: 'Sepáralos con comas.',
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
                const SizedBox(height: SophiaSpace.md),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TactileButton(
                      onPressed: () => Navigator.pop(dialogContext, false),
                      child: const Padding(
                        padding: EdgeInsets.all(SophiaSpace.sm),
                        child: Text('Cancelar'),
                      ),
                    ),
                    TactileButton(
                      onPressed: () => Navigator.pop(dialogContext, true),
                      child: Container(
                        padding: const EdgeInsets.symmetric(
                          horizontal: SophiaSpace.md,
                          vertical: SophiaSpace.sm,
                        ),
                        decoration: BoxDecoration(
                          color: context.colors.accent,
                          borderRadius: BorderRadius.circular(
                            SophiaRadius.control,
                          ),
                        ),
                        child: Text(
                          'Guardar',
                          style: TextStyle(color: context.colors.surface),
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ) ??
      false;
  if (save &&
      label.text.trim().isNotEmpty &&
      (existing != null || slug.text.trim().isNotEmpty)) {
    final values = aliases.text
        .split(',')
        .map((value) => value.trim())
        .where((value) => value.isNotEmpty)
        .toList();
    if (existing == null) {
      await cubit.create(
        kind: kind,
        slug: slug.text.trim(),
        label: label.text.trim(),
        aliases: values,
      );
    } else {
      await cubit.update(existing, label.text.trim(), values);
    }
  }
  label.dispose();
  slug.dispose();
  aliases.dispose();
}

class _KindSelector extends StatelessWidget {
  const _KindSelector({required this.value, required this.onChanged});

  final String value;
  final ValueChanged<String> onChanged;

  @override
  Widget build(BuildContext context) => Row(
    children: [
      for (final option in const [
        ('person', Icons.person_outline, 'Persona'),
        ('mode', Icons.layers_outlined, 'Modo'),
      ])
        Expanded(
          child: Padding(
            padding: EdgeInsets.only(
              right: option.$1 == 'person' ? SophiaSpace.xs : 0,
            ),
            child: TactileButton(
              semanticLabel: 'Tipo ${option.$3}',
              onPressed: () => onChanged(option.$1),
              child: AnimatedContainer(
                duration: SophiaMotion.resolve(context, SophiaMotion.short),
                curve: SophiaMotion.contentCurve,
                padding: const EdgeInsets.all(SophiaSpace.sm),
                decoration: BoxDecoration(
                  color: value == option.$1
                      ? context.colors.accent.withValues(
                          alpha: SophiaOpacity.subtle,
                        )
                      : context.colors.elevated,
                  border: Border.all(
                    color: value == option.$1
                        ? context.colors.accent
                        : context.colors.line,
                  ),
                  borderRadius: BorderRadius.circular(SophiaRadius.control),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(option.$2, size: SophiaSpace.md),
                    const SizedBox(width: SophiaSpace.xs),
                    Text(option.$3),
                  ],
                ),
              ),
            ),
          ),
        ),
    ],
  );
}

String _slugify(String value) {
  const accents = 'áéíóúüñ';
  const plain = 'aeiouun';
  var normalized = value.toLowerCase();
  for (var i = 0; i < accents.length; i++) {
    normalized = normalized.replaceAll(accents[i], plain[i]);
  }
  return normalized
      .replaceAll(RegExp(r'[^a-z0-9]+'), '-')
      .replaceAll(RegExp(r'^-+|-+$'), '');
}
