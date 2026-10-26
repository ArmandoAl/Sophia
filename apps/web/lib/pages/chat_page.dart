import 'package:jaspr/dom.dart';
import 'package:jaspr/jaspr.dart';
import 'package:sophia_core/features/chat/domain/entities/chat_message.dart';
import 'package:sophia_core/features/chat/presentation/cubit/chat_message_state.dart';
import 'package:web/web.dart' as web;

Component _t(String value) => Component.text(value);

String _hhmm(DateTime time) {
  final local = time.toLocal();
  return '${local.hour.toString().padLeft(2, '0')}:${local.minute.toString().padLeft(2, '0')}';
}

/// Conversación: historial cronológico, respuestas sugeridas, indicador de escritura, error con reintento.
/// Las propuestas de acción se muestran en sólo lectura; confirmarlas queda en la app móvil (ver docs/web_parity.md).
class ChatPage extends StatefulComponent {
  const ChatPage({
    required this.assistantName,
    required this.userName,
    required this.state,
    required this.onSend,
    required this.onRetry,
    required this.onLogout,
    super.key,
  });

  final String assistantName;
  final String userName;
  final ChatMessageState state;
  final Future<void> Function(String text) onSend;
  final void Function() onRetry;
  final void Function() onLogout;

  @override
  State<ChatPage> createState() => _ChatPageState();
}

class _ChatPageState extends State<ChatPage> {
  int _shown = 0;

  Future<void> _send(String text) async {
    final field = web.document.getElementById('message') as web.HTMLInputElement?;
    if (field != null) field.value = '';
    await component.onSend(text);
    (web.document.getElementById('message') as web.HTMLElement?)?.focus();
  }

  @override
  void didUpdateComponent(ChatPage old) {
    super.didUpdateComponent(old);
    final count = component.state.messages.length + (component.state.isTyping ? 1 : 0);
    if (count != _shown) {
      _shown = count;
      Future<void>.delayed(Duration.zero, () {
        final log = web.document.getElementById('log');
        if (log != null) log.scrollTop = log.scrollHeight;
      });
    }
  }

  @override
  Component build(BuildContext context) {
    final state = component.state;
    final chronological = state.messages.reversed.toList();
    return Component.fragment([
      header(classes: 'bar', [
        div(classes: 'grow', [
          h1([_t(component.assistantName)]),
          p(classes: 'sub', [_t(state.conversation?.title ?? 'Conversación')]),
        ]),
        span(classes: 'sub', [_t(component.userName)]),
        button([_t('Salir')], classes: 'btn', onClick: component.onLogout),
      ]),
      main_(id: 'contenido', [
        if (state.errorMessage != null)
          div(classes: 'alert', attributes: {'role': 'alert'}, [
            span([_t(state.errorMessage!)]),
            button([_t('Reintentar')], classes: 'btn', onClick: component.onRetry),
          ]),
        if (state.isLoading && state.messages.isEmpty)
          p(classes: 'hint', attributes: {'role': 'status'}, [_t('Cargando conversación…')])
        else if (state.messages.isEmpty && state.errorMessage == null)
          p(classes: 'hint', [_t('Aún no hay mensajes. Escribe lo que quieras organizar.')])
        else
          ul(id: 'log', classes: 'log', attributes: {'aria-label': 'Mensajes', 'aria-live': 'polite'}, [
            for (final message in chronological) _message(message),
            if (state.isTyping) li(classes: 'typing', [_t('${component.assistantName} está escribiendo…')]),
          ]),
        if (state.suggestedReplies.isNotEmpty && !state.isTyping)
          div(classes: 'chips', attributes: {'role': 'group', 'aria-label': 'Respuestas sugeridas'}, [
            for (final reply in state.suggestedReplies) button([_t(reply)], classes: 'btn chip', onClick: () => _send(reply)),
          ]),
        form(
          classes: 'composer',
          events: {
            'submit': (event) {
              event.preventDefault();
              final field = web.document.getElementById('message') as web.HTMLInputElement;
              final text = field.value.trim();
              if (text.isNotEmpty) _send(text);
            },
          },
          [
            label(classes: 'sr', attributes: {'for': 'message'}, [_t('Mensaje')]),
            input(id: 'message', type: InputType.text, name: 'message', attributes: {'placeholder': 'Escribe un mensaje', 'autocomplete': 'off'}),
            button([_t('Enviar')], type: ButtonType.submit, classes: 'btn btn-primary', disabled: state.isTyping),
          ],
        ),
      ]),
    ]);
  }

  Component _message(ChatMessage message) {
    if (message.type == MessageType.actionProposal) {
      final actions = (message.metadata?['actions'] as List?)?.cast<Map<String, dynamic>>() ?? const [];
      return li(classes: 'proposal', [
        h2([_t(message.text)]),
        ul([
          for (final action in actions) li([_t('${action['label']}${action['detail'] == null || '${action['detail']}'.isEmpty ? '' : ' — ${action['detail']}'}')]),
        ]),
        p([_t('Propuesta pendiente: se confirma desde la app móvil.')]),
      ]);
    }
    return li(classes: message.isUser ? 'msg user' : 'msg', [
      _t(message.text),
      Component.element(tag: 'time', attributes: {'datetime': message.timestamp.toIso8601String()}, children: [_t(_hhmm(message.timestamp))]),
    ]);
  }
}
