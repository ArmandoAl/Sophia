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
      child: Align(
        alignment: Alignment.topCenter,
        child: ConstrainedBox(
          constraints: const BoxConstraints(
            maxWidth: SophiaSize.contentMaxWidth,
          ),
          child: ListView(
            padding: const EdgeInsets.all(SophiaSpace.lg),
            children: [
              Text(
                'Tus datos te pertenecen',
                style: Theme.of(context).textTheme.headlineSmall,
              ),
              const SizedBox(height: SophiaSpace.sm),
              Text(
                'Aquí decides qué conservar, qué llevarte y cuándo pedir que cerremos tu cuenta.',
                style: TextStyle(color: context.colors.softInk),
              ),
              const SizedBox(height: SophiaSpace.xl),
              _PrivacySection(
                icon: Icons.ios_share_outlined,
                title: 'Una copia para ti',
                description:
                    'Preparamos tus datos en un archivo JSON y abrimos el menú del sistema para que elijas dónde guardarlo.',
                actionLabel: 'Exportar mis datos',
                onPressed: () => context.read<PrivacyCubit>().export(),
              ),
              const SizedBox(height: SophiaSpace.xl),
              _PrivacySection(
                icon: Icons.person_remove_outlined,
                title: 'Cerrar tu cuenta',
                description:
                    'Enviarás una solicitud de borrado. Tu cuenta no se elimina inmediatamente.',
                actionLabel: 'Solicitar borrado',
                onPressed: () => _confirmDelete(context),
              ),
            ],
          ),
        ),
      ),
    ),
  );

  Future<void> _confirmDelete(BuildContext context) async {
    final controller = TextEditingController();
    var canSubmit = false;
    final confirmed = await showSophiaSheet<bool>(
      context: context,
      builder: (dialogContext) => StatefulBuilder(
        builder: (context, setState) => Padding(
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
                'Escribe BORRAR para enviar la solicitud. Esto no borra la cuenta de inmediato.',
              ),
              const SizedBox(height: SophiaSpace.md),
              TextField(
                controller: controller,
                autofocus: true,
                decoration: const InputDecoration(labelText: 'Confirmación'),
                onChanged: (value) => setState(
                  () => canSubmit = value.trim().toUpperCase() == 'BORRAR',
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
                    onPressed: canSubmit
                        ? () => Navigator.pop(dialogContext, true)
                        : null,
                    child: Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: SophiaSpace.md,
                        vertical: SophiaSpace.sm,
                      ),
                      decoration: BoxDecoration(
                        color: canSubmit
                            ? context.colors.accent
                            : context.colors.line,
                        borderRadius: BorderRadius.circular(
                          SophiaRadius.control,
                        ),
                      ),
                      child: Text(
                        'Enviar solicitud',
                        style: TextStyle(
                          color: canSubmit
                              ? context.colors.surface
                              : context.colors.muted,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ],
          ),
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

class _PrivacySection extends StatelessWidget {
  const _PrivacySection({
    required this.icon,
    required this.title,
    required this.description,
    required this.actionLabel,
    required this.onPressed,
  });

  final IconData icon;
  final String title, description, actionLabel;
  final VoidCallback onPressed;

  @override
  Widget build(BuildContext context) => Container(
    padding: const EdgeInsets.all(SophiaSpace.lg),
    decoration: BoxDecoration(
      color: context.colors.elevated,
      border: Border.all(color: context.colors.line),
      borderRadius: BorderRadius.circular(SophiaRadius.card),
    ),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Icon(icon, color: context.colors.softInk),
        const SizedBox(height: SophiaSpace.md),
        Text(title, style: Theme.of(context).textTheme.titleMedium),
        const SizedBox(height: SophiaSpace.xs),
        Text(
          description,
          style: Theme.of(
            context,
          ).textTheme.bodyMedium?.copyWith(color: context.colors.softInk),
        ),
        const SizedBox(height: SophiaSpace.lg),
        TactileButton(
          semanticLabel: actionLabel,
          onPressed: onPressed,
          child: Container(
            padding: const EdgeInsets.symmetric(
              horizontal: SophiaSpace.md,
              vertical: SophiaSpace.sm,
            ),
            decoration: BoxDecoration(
              border: Border.all(color: context.colors.line),
              borderRadius: BorderRadius.circular(SophiaRadius.control),
            ),
            child: Text(
              actionLabel,
              style: Theme.of(context).textTheme.labelLarge,
            ),
          ),
        ),
      ],
    ),
  );
}
