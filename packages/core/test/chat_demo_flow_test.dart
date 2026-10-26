import 'package:sophia_core/demo/demo_repositories.dart';
import 'package:sophia_core/demo/demo_store.dart';
import 'package:sophia_core/features/chat/domain/entities/chat_message.dart';
import 'package:sophia_core/features/chat/presentation/cubit/chat_message_cubit.dart';
import 'package:test/test.dart';

void main() {
  test('conversación demo: bienvenida, preguntas guiadas y propuesta', () async {
    final cubit = ChatMessageCubit(repository: DemoConversationsRepository(DemoStore.instance));
    addTearDown(cubit.close);

    await cubit.load();
    expect(cubit.state.messages.single.text, startsWith('Hola, Alex'));

    await cubit.sendMessage('Quiero una pausa');
    expect(cubit.state.suggestedReplies, ['Por la mañana', 'Después de comer', 'Por la tarde']);
    expect(cubit.state.messages.first.isUser, isFalse, reason: 'el más reciente va primero');

    await cubit.sendMessage('Por la tarde');
    await cubit.sendMessage('20 minutos');
    final last = cubit.state.messages.first;
    expect(last.type, MessageType.actionProposal);
    expect((last.metadata!['actions'] as List).single['label'], 'Pausa tranquila');
  });

  test('ignora mensajes vacíos', () async {
    final cubit = ChatMessageCubit(repository: DemoConversationsRepository(DemoStore.instance));
    addTearDown(cubit.close);
    await cubit.load();
    final before = cubit.state.messages.length;
    await cubit.sendMessage('   ');
    expect(cubit.state.messages.length, before);
  });
}
