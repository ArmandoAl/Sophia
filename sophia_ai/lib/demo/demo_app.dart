import 'package:flutter/material.dart';
import 'package:sophia_ai/core/theme/app_theme.dart';
import 'package:sophia_ai/core/theme/design_tokens.dart';
import 'package:sophia_ai/core/widgets/action_proposal_card.dart';

class DemoApp extends StatelessWidget {
  const DemoApp({super.key});

  @override
  Widget build(BuildContext context) => MaterialApp(
    title: 'Sofía · Demo',
    debugShowCheckedModeBanner: false,
    theme: AppTheme.darkTheme,
    home: const DemoHome(),
  );
}

class DemoHome extends StatefulWidget {
  const DemoHome({super.key});

  @override
  State<DemoHome> createState() => _DemoHomeState();
}

class _DemoHomeState extends State<DemoHome> {
  int _tab = 0;

  @override
  Widget build(BuildContext context) => Scaffold(
    body: SafeArea(
      child: Column(
        children: [
          MaterialBanner(
            content: const Text('DEMO · Datos ficticios. Sin IA ni acciones reales.'),
            leading: const Icon(Icons.science_outlined),
            actions: [const SizedBox.shrink()],
          ),
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: SophiaSpace.lg),
            child: Row(
              children: [
                for (final (index, label) in const [
                  (0, 'Conversación'),
                  (1, 'Actividades'),
                  (2, 'Creencias'),
                ])
                  TextButton(
                    onPressed: () => setState(() => _tab = index),
                    child: Text(label, style: TextStyle(color: _tab == index ? context.colors.accent : null)),
                  ),
              ],
            ),
          ),
          Expanded(child: _content()),
        ],
      ),
    ),
  );

  Widget _content() => switch (_tab) {
    0 => _conversation(),
    1 => _list('Actividades de ejemplo', const [
      ('08:30', 'Paseo matutino', 'Actividad sugerida · solo vista previa'),
      ('12:00', 'Pausa para comer', 'Recordatorio ficticio'),
      ('18:00', 'Llamar a una amistad', 'Propuesta de ejemplo'),
    ]),
    _ => _list('Creencias de ejemplo', const [
      ('Preferencias', 'Prefieres empezar el día con calma.', ''),
      ('Rutina', 'Una caminata breve suele ayudarte a despejarte.', ''),
      ('Contexto', 'Te gusta reservar la tarde para descansar.', ''),
    ]),
  };

  Widget _conversation() => ListView(
    padding: const EdgeInsets.all(SophiaSpace.lg),
    children: [
      _bubble('Hola. Esta es una conversación de muestra con contenido ficticio.', false),
      _bubble('¿Me ayudas a organizar una pausa para caminar hoy?', true),
      _bubble('Respuesta escrita para esta demo; no fue generada por una IA.', false),
      const SizedBox(height: SophiaSpace.md),
      ActionProposalCard(
        title: 'Propuesta ficticia · paseo',
        actions: const [
          ProposedAction(
            id: 'walk',
            icon: Icons.directions_walk,
            label: 'Paseo de 20 minutos',
            detail: 'Hoy a las 17:30 · ejemplo local; no crea un recordatorio',
            color: Colors.teal,
            proposedInput: {'hora': '17:30', 'duración': '20 minutos'},
          ),
        ],
        onConfirm: (action, {correctedInput, required decisionLatencyMs}) async {},
        onReject: (action, {required rejectionReason, required decisionLatencyMs}) async {},
      ),
    ],
  );

  Widget _bubble(String text, bool user) => Align(
    alignment: user ? Alignment.centerRight : Alignment.centerLeft,
    child: Container(
      constraints: const BoxConstraints(maxWidth: 560),
      margin: const EdgeInsets.only(bottom: SophiaSpace.sm),
      padding: const EdgeInsets.all(SophiaSpace.md),
      decoration: BoxDecoration(
        color: context.colors.elevated,
        borderRadius: BorderRadius.circular(SophiaRadius.card),
      ),
      child: Text(text),
    ),
  );

  Widget _list(String title, List<(String, String, String)> entries) => ListView(
    padding: const EdgeInsets.all(SophiaSpace.lg),
    children: [
      Text(title, style: Theme.of(context).textTheme.headlineSmall),
      const SizedBox(height: SophiaSpace.md),
      for (final (first, second, third) in entries)
        Card(
          child: ListTile(
            leading: Icon(_tab == 1 ? Icons.event_note : Icons.lightbulb_outline),
            title: Text(_tab == 1 ? '$first · $second' : first),
            subtitle: Text(_tab == 1 ? third : second),
          ),
        ),
    ],
  );
}
