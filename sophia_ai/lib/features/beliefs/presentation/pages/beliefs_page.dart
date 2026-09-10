import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/services.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/theme/motion.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../../contexts/presentation/cubit/contexts_cubit.dart';
import '../../domain/models.dart';
import '../cubit/beliefs_cubit.dart';

class BeliefsPage extends StatelessWidget {
  const BeliefsPage({super.key});
  @override
  Widget build(BuildContext context) => MultiBlocProvider(
    providers: [
      BlocProvider(create: (_) => sl<BeliefsCubit>()..load()),
      BlocProvider(create: (_) => sl<ContextsCubit>()..load()),
    ],
    child: const _BeliefsView(),
  );
}

class _BeliefsView extends StatelessWidget {
  const _BeliefsView();
  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Lo que Sofía sabe de ti')),
    body: BlocBuilder<BeliefsCubit, BeliefsState>(
      builder: (context, state) {
        if (state.loading && state.beliefs.isEmpty) {
          return const MotionSwap(
            child: Padding(
              key: ValueKey('beliefs-loading'),
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
        final groups = <String, List<Belief>>{};
        for (final belief in state.beliefs) {
          (groups[belief.category] ??= []).add(belief);
        }
        var entryIndex = 0;
        return MotionSwap(
          child: RefreshIndicator(
            key: const ValueKey('beliefs-content'),
            onRefresh: () => context.read<BeliefsCubit>().load(
              scope: state.scope,
              scopeKey: state.scopeKey,
            ),
            child: ListView(
              padding: const EdgeInsets.all(SophiaSpace.lg),
              children: [
                Text(
                  'Esto es lo que creo haber entendido. Es tuyo: puedes matizarlo o decirme que no es cierto.',
                  style: TextStyle(color: context.colors.softInk),
                ),
                const SizedBox(height: SophiaSpace.lg),
                BlocBuilder<ContextsCubit, ContextsState>(
                  builder: (context, contexts) =>
                      DropdownButtonFormField<String>(
                        initialValue: state.scopeKey ?? state.scope ?? 'all',
                        decoration: const InputDecoration(
                          labelText: 'Contexto',
                        ),
                        items: [
                          const DropdownMenuItem(
                            value: 'all',
                            child: Text('Todas'),
                          ),
                          const DropdownMenuItem(
                            value: 'global',
                            child: Text('Global'),
                          ),
                          ...contexts.contexts
                              .where((item) => item.active)
                              .map(
                                (item) => DropdownMenuItem(
                                  value: item.scopeKey,
                                  child: Text(item.label),
                                ),
                              ),
                        ],
                        onChanged: (value) {
                          HapticFeedback.selectionClick();
                          if (value == 'global') {
                            context.read<BeliefsCubit>().load(scope: 'global');
                          } else if (value == null || value == 'all') {
                            context.read<BeliefsCubit>().load();
                          } else {
                            context.read<BeliefsCubit>().load(scopeKey: value);
                          }
                        },
                      ),
                ),
                if (state.error != null)
                  Padding(
                    padding: const EdgeInsets.only(top: SophiaSpace.md),
                    child: Text(
                      state.error!,
                      style: TextStyle(color: context.colors.critical),
                    ),
                  ),
                if (!state.loading && state.beliefs.isEmpty)
                  const Padding(
                    padding: EdgeInsets.only(top: SophiaSpace.xxl),
                    child: Center(
                      child: Text('Todavía no he formado creencias sobre ti.'),
                    ),
                  ),
                for (final entry in groups.entries)
                  StaggeredEntry(
                    index: entryIndex++,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Padding(
                          padding: const EdgeInsets.fromLTRB(
                            0,
                            SophiaSpace.xl,
                            0,
                            SophiaSpace.sm,
                          ),
                          child: Text(
                            _category(entry.key),
                            style: Theme.of(context).textTheme.titleMedium,
                          ),
                        ),
                        for (final belief in entry.value)
                          _BeliefCard(belief: belief),
                      ],
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    ),
  );
}

class _BeliefCard extends StatefulWidget {
  const _BeliefCard({required this.belief});
  final Belief belief;
  @override
  State<_BeliefCard> createState() => _BeliefCardState();
}

class _BeliefCardState extends State<_BeliefCard> {
  bool retiring = false;
  @override
  Widget build(BuildContext context) => MotionSwap(
    child: retiring
        ? const SizedBox.shrink(key: ValueKey('retired'))
        : Container(
            key: const ValueKey('belief'),
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
              Expanded(
                child: Hero(
                  tag: 'belief-${widget.belief.id}',
                  transitionOnUserGestures: true,
                  child: Material(
                    color: context.colors.surface.withValues(alpha: 0),
                    child: Text(
                      'Creo que ${_tentative(widget.belief.statement)}',
                      style: Theme.of(context).textTheme.bodyLarge,
                    ),
                  ),
                ),
              ),
              if (widget.belief.promptSlot == 'core')
                Tooltip(
                  message: 'Influye en cómo se comporta Sofía hoy',
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: SophiaSpace.xs,
                      vertical: SophiaSpace.xxs,
                    ),
                    decoration: BoxDecoration(
                      color: context.colors.accent.withValues(
                        alpha: SophiaOpacity.subtle,
                      ),
                      borderRadius: BorderRadius.circular(
                        SophiaRadius.control,
                      ),
                    ),
                    child: Text(
                      'Guía a Sofía hoy',
                      style: Theme.of(context).textTheme.labelMedium?.copyWith(
                        color: context.colors.accent,
                      ),
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: SophiaSpace.md),
          Tooltip(
            message:
                'Confianza ${(widget.belief.decayedConfidence * 100).round()}%',
            child: Semantics(
              label:
                  '${_confidenceLabel(widget.belief.decayedConfidence)}. Mantén pulsado para ver el porcentaje.',
              child: ClipRRect(
                borderRadius: BorderRadius.circular(SophiaRadius.control),
                child: LinearProgressIndicator(
                  value: widget.belief.decayedConfidence.clamp(0, 1),
                  minHeight: SophiaSpace.xxs,
                  backgroundColor: context.colors.line,
                  color: context.colors.accent,
                ),
              ),
            ),
          ),
          const SizedBox(height: SophiaSpace.xs),
          Row(
            children: [
              Expanded(
                child: Text(
                  _origin(widget.belief.trustTier),
                  style: Theme.of(context).textTheme.bodySmall?.copyWith(
                    color: context.colors.softInk,
                  ),
                ),
              ),
              Text(
                _confidenceLabel(widget.belief.decayedConfidence),
                style: Theme.of(context).textTheme.labelMedium,
              ),
            ],
          ),
          const SizedBox(height: SophiaSpace.sm),
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              TactileButton(
                semanticLabel: 'Indicar que esta creencia no es cierta',
                onPressed: _retire,
                child: const Padding(
                  padding: EdgeInsets.all(SophiaSpace.sm),
                  child: Text('No es cierto'),
                ),
              ),
              TactileButton(
                semanticLabel: 'Corregir esta creencia',
                onPressed: _edit,
                child: Padding(
                  padding: const EdgeInsets.all(SophiaSpace.sm),
                  child: Text(
                    'Corregir',
                    style: TextStyle(color: context.colors.accent),
                  ),
                ),
              ),
            ],
          ),
        ],
      ),
    ),
  );

  Future<void> _retire() async {
    final cubit = context.read<BeliefsCubit>();
    final confirmed =
        await showSophiaSheet<bool>(
          context: context,
          builder: (sheetContext) => Padding(
            padding: const EdgeInsets.all(SophiaSpace.lg),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '¿No encaja contigo?',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
                const SizedBox(height: SophiaSpace.sm),
                const Text(
                  'La retiraremos con calma y dejará de influir en Sofía.',
                ),
                const SizedBox(height: SophiaSpace.lg),
                Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TactileButton(
                      onPressed: () => Navigator.pop(sheetContext, false),
                      child: const Padding(
                        padding: EdgeInsets.all(SophiaSpace.sm),
                        child: Text('Conservar'),
                      ),
                    ),
                    TactileButton(
                      onPressed: () => Navigator.pop(sheetContext, true),
                      child: Padding(
                        padding: const EdgeInsets.all(SophiaSpace.sm),
                        child: Text(
                          'Retirar',
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
    if (!confirmed || !mounted) return;
    setState(() => retiring = true);
    await Future<void>.delayed(
      SophiaMotion.resolve(context, SophiaMotion.short),
    );
    await cubit.retire(widget.belief.id);
  }

  Future<void> _edit() async {
    final cubit = context.read<BeliefsCubit>();
    final controller = TextEditingController(text: widget.belief.statement);
    final value = await showSophiaSheet<String>(
      context: context,
      builder: (sheetContext) => Padding(
        padding: EdgeInsets.fromLTRB(
          SophiaSpace.lg,
          SophiaSpace.md,
          SophiaSpace.lg,
          MediaQuery.viewInsetsOf(sheetContext).bottom + SophiaSpace.lg,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Hero(
              tag: 'belief-${widget.belief.id}',
              transitionOnUserGestures: true,
              child: Material(
                color: context.colors.surface.withValues(alpha: 0),
                child: Text(
                  'Corrige lo que entendí',
                  style: Theme.of(context).textTheme.titleLarge,
                ),
              ),
            ),
            const SizedBox(height: SophiaSpace.md),
            TextField(controller: controller, autofocus: true, maxLines: 4),
            const SizedBox(height: SophiaSpace.md),
            Align(
              alignment: Alignment.centerRight,
              child: TactileButton(
                onPressed: () =>
                    Navigator.pop(sheetContext, controller.text.trim()),
                child: Container(
                  padding: const EdgeInsets.symmetric(
                    horizontal: SophiaSpace.md,
                    vertical: SophiaSpace.sm,
                  ),
                  decoration: BoxDecoration(
                    color: context.colors.accent,
                    borderRadius: BorderRadius.circular(SophiaRadius.control),
                  ),
                  child: Text(
                    'Guardar',
                    style: TextStyle(color: context.colors.surface),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );
    controller.dispose();
    if (value?.isNotEmpty == true) {
      await cubit.correct(widget.belief.id, value!);
    }
  }
}

String _tentative(String value) {
  final trimmed = value.trim();
  if (trimmed.isEmpty) return trimmed;
  return trimmed[0].toLowerCase() + trimmed.substring(1);
}

String _origin(int tier) => switch (tier) {
  1 => 'Porque lo decidiste',
  2 => 'Porque lo mencionaste',
  _ => 'Deducido de tus conversaciones',
};

String _confidenceLabel(double confidence) {
  if (confidence >= .75) return 'Bastante asentado';
  if (confidence >= .45) return 'Aún afinándolo';
  return 'Muy tentativo';
}

String _category(String value) => switch (value) {
  'schedule' => 'Horarios',
  'communication' => 'Comunicación',
  'priorities' => 'Prioridades',
  'work_style' => 'Forma de trabajar',
  'personal' => 'Personal',
  'constraint' => 'Límites',
  _ => value,
};
