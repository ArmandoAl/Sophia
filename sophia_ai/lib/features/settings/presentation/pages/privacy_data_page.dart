import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../cubit/settings_services_cubit.dart';

class PrivacyDataPage extends StatelessWidget {
  const PrivacyDataPage({super.key});

  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) => sl<PrivacyCubit>(),
    child: const _PrivacyDataView(),
  );
}

class _PrivacyDataView extends StatelessWidget {
  const _PrivacyDataView();

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(title: const Text('Privacidad y datos')),
    body: BlocListener<PrivacyCubit, String?>(
      listener: (context, error) {
        if (error != null) {
          ScaffoldMessenger.of(
            context,
          ).showSnackBar(SnackBar(content: Text(error)));
        }
      },
      child: ListView(
        padding: const EdgeInsets.all(SophiaSpace.lg),
        children: [
          const Text(
            'Tus datos te pertenecen. Puedes obtener una copia o solicitar que eliminemos tu cuenta.',
          ),
          const SizedBox(height: SophiaSpace.xl),
          TactileButton(
            onPressed: () => context.read<PrivacyCubit>().export(),
            child: Container(
              decoration: BoxDecoration(
                color: context.colors.elevated,
                border: Border.all(color: context.colors.line),
                borderRadius: BorderRadius.circular(SophiaRadius.card),
              ),
              child: ListTile(
                leading: const Icon(Icons.ios_share_outlined),
                title: const Text('Exportar mis datos'),
                subtitle: const Text(
                  'Descarga un JSON y abre el menú para compartirlo.',
                ),
              ),
            ),
          ),
          const SizedBox(height: SophiaSpace.lg),
          TactileButton(
            onPressed: () => _confirmDelete(context),
            child: Container(
              decoration: BoxDecoration(
                color: context.colors.elevated,
                border: Border.all(color: context.colors.line),
                borderRadius: BorderRadius.circular(SophiaRadius.card),
              ),
              child: ListTile(
                leading: Icon(
                  Icons.person_remove_outlined,
                  color: context.colors.critical,
                ),
                title: const Text('Solicitar borrado'),
                subtitle: const Text(
                  'Es una solicitud; tu cuenta no se borra inmediatamente.',
                ),
              ),
            ),
          ),
        ],
      ),
    ),
  );

  Future<void> _confirmDelete(BuildContext context) async {
    final controller = TextEditingController();
    final confirmed = await showSophiaSheet<bool>(
      context: context,
      builder: (dialogContext) => Padding(
        padding: EdgeInsets.fromLTRB(
          SophiaSpace.lg,
          SophiaSpace.md,
          SophiaSpace.lg,
          MediaQuery.viewInsetsOf(dialogContext).bottom + SophiaSpace.lg,
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              'Solicitar borrado',
              style: Theme.of(context).textTheme.titleLarge,
            ),
            const SizedBox(height: SophiaSpace.sm),
            const Text(
              'Escribe BORRAR para enviar la solicitud. No es un borrado inmediato.',
            ),
            const SizedBox(height: SophiaSpace.md),
            TextField(
              controller: controller,
              autofocus: true,
              decoration: const InputDecoration(labelText: 'BORRAR'),
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                TextButton(
                  onPressed: () => Navigator.pop(dialogContext, false),
                  child: const Text('Cancelar'),
                ),
                FilledButton(
                  onPressed: () =>
                      Navigator.pop(dialogContext, controller.text == 'BORRAR'),
                  child: const Text('Solicitar'),
                ),
              ],
            ),
          ],
        ),
      ),
    );
    controller.dispose();
    if (confirmed != true || !context.mounted) return;
    final ok = await context.read<PrivacyCubit>().requestDeletion();
    if (ok && context.mounted) {
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(const SnackBar(content: Text('Solicitud enviada.')));
    }
  }
}
