import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../cubit/settings_services_cubit.dart';

class ImportConversationsPage extends StatefulWidget {
  const ImportConversationsPage({super.key});
  @override
  State<ImportConversationsPage> createState() =>
      _ImportConversationsPageState();
}

class _ImportConversationsPageState extends State<ImportConversationsPage> {
  bool understood = false;

  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) => sl<IngestionCubit>(),
    child: Scaffold(
      appBar: AppBar(title: const Text('Importar conversaciones')),
      body: MotionSwap(
        child: understood
            ? const _BatchList()
            : _Warning(onContinue: () => setState(() => understood = true)),
      ),
    ),
  );
}

class _Warning extends StatelessWidget {
  const _Warning({required this.onContinue});
  final VoidCallback onContinue;
  @override
  Widget build(BuildContext context) => Padding(
    padding: const EdgeInsets.all(SophiaSpace.lg),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(
          Icons.privacy_tip_outlined,
          size: 40,
          color: context.colors.attention,
        ),
        const SizedBox(height: SophiaSpace.lg),
        Text(
          'Antes de continuar',
          style: Theme.of(context).textTheme.headlineSmall,
        ),
        const SizedBox(height: SophiaSpace.md),
        const Text(
          'Importar procesa mensajes de otras personas que no dieron su consentimiento.',
        ),
        const SizedBox(height: SophiaSpace.sm),
        const Text(
          'Lo aprendido entra como creencia de nivel 3, con confianza máxima de 0.5, hasta que una decisión tuya lo corrobore.',
        ),
        const Spacer(),
        SizedBox(
          width: double.infinity,
          child: TactileButton(
            onPressed: onContinue,
            child: Container(
              padding: const EdgeInsets.all(SophiaSpace.md),
              alignment: Alignment.center,
              decoration: BoxDecoration(
                color: context.colors.accent,
                borderRadius: BorderRadius.circular(SophiaRadius.control),
              ),
              child: Text(
                'He leído y entiendo',
                style: TextStyle(color: context.colors.surface),
              ),
            ),
          ),
        ),
      ],
    ),
  );
}

class _BatchList extends StatefulWidget {
  const _BatchList();
  @override
  State<_BatchList> createState() => _BatchListState();
}

class _BatchListState extends State<_BatchList> {
  @override
  void initState() {
    super.initState();
    context.read<IngestionCubit>().load();
  }

  @override
  Widget build(
    BuildContext context,
  ) => BlocBuilder<IngestionCubit, IngestionState>(
    builder: (context, state) {
      if (state.loading && state.batches.isEmpty) {
        return const MotionSwap(
          child: Padding(
            key: ValueKey('imports-loading'),
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
      return MotionSwap(
        child: RefreshIndicator(
          key: const ValueKey('imports-content'),
          onRefresh: context.read<IngestionCubit>().load,
          child: ListView(
            padding: const EdgeInsets.all(SophiaSpace.lg),
            children: [
              const Text(
                'El selector de archivos llegará con el conversor. Por ahora puedes revisar y deshacer lotes enviados por API.',
              ),
              const SizedBox(height: SophiaSpace.lg),
              if (state.error != null)
                Text(
                  state.error!,
                  style: TextStyle(color: context.colors.critical),
                ),
              if (state.batches.isEmpty)
                const Padding(
                  padding: EdgeInsets.only(top: 32),
                  child: Text('No hay lotes importados.'),
                ),
              for (final batch in state.batches)
                Card(
                  child: ListTile(
                    title: Text(
                      'Lote ${batch.id.substring(0, batch.id.length.clamp(0, 8))}',
                    ),
                    subtitle: Text(
                      '${batch.status} · ${batch.beliefsCreated} creencias',
                    ),
                    trailing: TactileButton(
                      onPressed: () => _undo(context, batch.id),
                      child: const Padding(
                        padding: EdgeInsets.all(SophiaSpace.xs),
                        child: Text('Deshacer'),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      );
    },
  );

  Future<void> _undo(BuildContext context, String id) async {
    final ok = await showSophiaSheet<bool>(
      context: context,
      builder: (dialogContext) => Padding(
        padding: const EdgeInsets.all(SophiaSpace.lg),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Deshacer lote',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: SophiaSpace.sm),
            const Text(
              'Se retirarán todas las creencias generadas por este lote.',
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: () => Navigator.pop(dialogContext, false),
                  child: const Text('Cancelar'),
                ),
                FilledButton(
                  onPressed: () => Navigator.pop(dialogContext, true),
                  child: const Text('Deshacer'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
    if (ok == true && context.mounted) {
      await context.read<IngestionCubit>().undo(id);
    }
  }
}
