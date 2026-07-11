import 'dart:ui';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:sophia_ai/core/widgets/neon_wrapper.dart';
import 'package:sophia_ai/features/chat/presentation/cubit/chat_message_state.dart';
import '../cubit/chat_message_cubit.dart';
import 'package:sophia_ai/features/chat/domain/entities/chat_message.dart';
import 'package:sophia_ai/core/widgets/schedule_conflicts_card.dart';
import 'package:sophia_ai/core/widgets/map_location_card.dart';
import 'package:sophia_ai/core/widgets/voice_visualizer.dart';
import 'package:sophia_ai/core/widgets/action_proposal_card.dart';

class ChatPage extends StatelessWidget {
  const ChatPage({super.key});

  @override
  Widget build(BuildContext context) {
    return NeonWrapper(
      child: BlocProvider(
        create: (_) => ChatMessageCubit(),
        child: Scaffold(
          backgroundColor: Colors.transparent,
          appBar: AppBar(
            forceMaterialTransparency: true,
            title: Row(
              children: [
                const CircleAvatar(
                  backgroundImage: NetworkImage(
                    "https://i.pravatar.cc/150?u=sophia",
                  ), // Sophia Avatar
                  radius: 16,
                ),
                const SizedBox(width: 10),
                Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text("Sophia", style: TextStyle(fontSize: 16)),
                    Text(
                      "Calm",
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.greenAccent[400],
                      ),
                    ),
                  ],
                ),
              ],
            ),
            actions: [
              Container(
                margin: const EdgeInsets.only(right: 16),
                decoration: BoxDecoration(
                  color: Colors.white.withValues(alpha: 0.05),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(
                    color: Colors.white.withValues(alpha: 0.1),
                  ),
                ),
                child: IconButton(
                  icon: const Icon(Icons.more_horiz, color: Colors.white),
                  onPressed: () {},
                ),
              ),
            ],
          ),
          body: Column(
            children: [
              // LISTA DE MENSAJES
              Expanded(
                child: BlocBuilder<ChatMessageCubit, ChatMessageState>(
                  builder: (context, state) {
                    return ListView.builder(
                      reverse: true, // Los chats empiezan desde abajo
                      padding: const EdgeInsets.all(10),
                      itemCount: state.messages.length,
                      itemBuilder: (context, index) {
                        final msg = state.messages[index];
                        return _buildMessageItem(context, msg);
                      },
                    );
                  },
                ),
              ),

              // ÁREA DE INPUT
              const _GlassChatInputArea(),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildMessageItem(BuildContext context, ChatMessage msg) {
    final isMe = msg.isUser;

    // Si el mensaje tiene un tipo especial, renderizamos el widget correspondiente
    Widget content;
    switch (msg.type) {
      case MessageType.scheduleConflict:
        content = const ScheduleConflictCard();
        break;
      case MessageType.mapLocation:
        content = const MapLocationCard();
        break;
      case MessageType.actionProposal:
        final proposalTitle = msg.metadata?['title'] as String? ?? 'Proposal';
        final actionsData = msg.metadata?['actions'] as List? ?? [];

        final actions = actionsData.map((action) {
          return ProposedAction(
            id: action['id'] as String,
            icon: _getIconData(action['icon'] as String),
            label: action['label'] as String,
            detail: action['detail'] as String,
            color: _getColorFromString(action['color'] as String?),
          );
        }).toList();

        content = ActionProposalCard(
          title: proposalTitle,
          actions: actions,
          onConfirm: () {
            // TODO: Implementar confirmación de propuesta
            debugPrint("Propuesta confirmada: $proposalTitle");
          },
          onModify: () {
            // TODO: Implementar modificación de propuesta
            debugPrint("Modificar propuesta: $proposalTitle");
          },
        );
        break;
      default:
        content = Text(
          msg.text,
          style: TextStyle(
            color: Colors.white.withValues(alpha: 0.9),
            height: 1.4, // Mejor lectura
          ),
        );
    }

    return Column(
      children: [
        Align(
          alignment: isMe ? Alignment.centerRight : Alignment.centerLeft,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            crossAxisAlignment: CrossAxisAlignment.end,
            children: [
              // Avatar para IA
              if (!isMe) ...[
                const CircleAvatar(
                  backgroundImage: NetworkImage(
                    "https://i.pravatar.cc/150?u=sophia",
                  ),
                  radius: 14,
                ),
                const SizedBox(width: 8),
              ],

              // Burbuja Glassmorphism
              Flexible(
                child: ClipRRect(
                  borderRadius: BorderRadius.only(
                    topLeft: const Radius.circular(20),
                    topRight: const Radius.circular(20),
                    bottomLeft: Radius.circular(isMe ? 20 : 3),
                    bottomRight: Radius.circular(isMe ? 3 : 20),
                  ),
                  child: BackdropFilter(
                    filter: ImageFilter.blur(sigmaX: 5, sigmaY: 5), // Blur real
                    child: Container(
                      constraints: const BoxConstraints(maxWidth: 300),
                      padding: const EdgeInsets.symmetric(
                        horizontal: 10,
                        vertical: 12,
                      ),
                      decoration: BoxDecoration(
                        // Lógica de colores basada en el HTML
                        color: isMe
                            ? Theme.of(context).primaryColor.withValues(
                                alpha: 0.3,
                              ) // bg-primary/30
                            : Colors.white.withValues(
                                alpha: 0.05,
                              ), // bg-white/5
                        border: Border.all(
                          color: isMe
                              ? Theme.of(
                                  context,
                                ).primaryColor.withValues(alpha: 0.5)
                              : Colors.white.withValues(alpha: 0.1),
                          width: 1,
                        ),
                      ),
                      child: content,
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
        SizedBox(height: 10),
      ],
    );
  }

  // Helper para convertir string a IconData
  IconData _getIconData(String iconName) {
    switch (iconName) {
      case 'lightbulb':
        return Icons.lightbulb;
      case 'music_note':
        return Icons.music_note;
      case 'ac_unit':
        return Icons.ac_unit;
      case 'tv':
        return Icons.tv;
      case 'coffee':
        return Icons.coffee;
      case 'sports_esports':
        return Icons.sports_esports;
      case 'weekend':
        return Icons.weekend;
      case 'door_front':
        return Icons.door_front_door;
      default:
        return Icons.check_circle;
    }
  }

  // Helper para convertir string a Color
  Color _getColorFromString(String? colorName) {
    switch (colorName) {
      case 'cyan':
        return Colors.cyan;
      case 'purple':
        return Colors.purple;
      case 'blue':
        return Colors.blue;
      case 'green':
        return Colors.green;
      case 'orange':
        return Colors.orange;
      case 'pink':
        return Colors.pink;
      default:
        return Colors.cyan;
    }
  }
}

class _ChatInputArea extends StatefulWidget {
  const _ChatInputArea();

  @override
  State<_ChatInputArea> createState() => _ChatInputAreaState();
}

class _ChatInputAreaState extends State<_ChatInputArea> {
  final TextEditingController _ctrl = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return BlocBuilder<ChatMessageCubit, ChatMessageState>(
      builder: (context, state) {
        return Container(
          padding: const EdgeInsets.all(16),
          color: const Color(0xFF0B1016), // Fondo igual al Scaffold
          child: SafeArea(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Si está escuchando, mostramos el visualizador (Screen 3)
                if (state.isListening)
                  Padding(
                    padding: const EdgeInsets.only(bottom: 20),
                    child: Column(
                      children: [
                        const Text(
                          "Listening...",
                          style: TextStyle(color: Colors.cyanAccent),
                        ),
                        const SizedBox(height: 10),
                        const VoiceVisualizer(),
                      ],
                    ),
                  ),

                Row(
                  children: [
                    IconButton(
                      icon: const Icon(Icons.add, color: Colors.grey),
                      onPressed: () {},
                    ),
                    Expanded(
                      child: TextField(
                        controller: _ctrl,
                        decoration: InputDecoration(
                          hintText: state.isListening
                              ? "Say something..."
                              : "Type your message...",
                          hintStyle: const TextStyle(color: Colors.grey),
                          filled: true,
                          fillColor: const Color(0xFF151B24),
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(30),
                            borderSide: BorderSide.none,
                          ),
                          contentPadding: const EdgeInsets.symmetric(
                            horizontal: 20,
                            vertical: 10,
                          ),
                        ),
                        onSubmitted: (val) {
                          context.read<ChatMessageCubit>().sendMessage(val);
                          _ctrl.clear();
                        },
                      ),
                    ),
                    const SizedBox(width: 8),
                    // Botón Micrófono o Enviar
                    GestureDetector(
                      onTap: () {
                        if (_ctrl.text.isNotEmpty) {
                          context.read<ChatMessageCubit>().sendMessage(
                            _ctrl.text,
                          );
                          _ctrl.clear();
                        } else {
                          context.read<ChatMessageCubit>().toggleListening();
                        }
                      },
                      child: Container(
                        padding: const EdgeInsets.all(12),
                        decoration: BoxDecoration(
                          shape: BoxShape.circle,
                          color: state.isListening
                              ? Colors.redAccent
                              : const Color(0xFF2E5CB8),
                          boxShadow: [
                            BoxShadow(
                              color:
                                  (state.isListening ? Colors.red : Colors.blue)
                                      .withValues(alpha: 0.4),
                              blurRadius: 10,
                              spreadRadius: 2,
                            ),
                          ],
                        ),
                        child: Icon(
                          _ctrl.text.isEmpty ? Icons.mic : Icons.send,
                          color: Colors.white,
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        );
      },
    );
  }
}

class _GlassChatInputArea extends StatelessWidget {
  const _GlassChatInputArea();

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    // Barra inferior estilo Glass
    return ClipRRect(
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10), // Blur intenso fondo
        child: Container(
          padding: const EdgeInsets.fromLTRB(16, 16, 16, 32),
          decoration: BoxDecoration(
            color: const Color(
              0xFF101022,
            ).withValues(alpha: 0.6), // background-dark/50
            border: Border.all(
              color: Colors.white.withValues(alpha: 0.05),
              width: 0,
            ),
          ),
          child: Row(
            children: [
              // Botón +
              CircleAvatar(
                backgroundColor: Colors.transparent,
                child: IconButton(
                  icon: const Icon(Icons.add, color: Color(0xFF9292C9)),
                  onPressed: () {},
                ),
              ),
              const SizedBox(width: 8),

              // Input Field
              Expanded(
                child: Container(
                  height: 50,
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.05), // bg-white/5
                    borderRadius: BorderRadius.circular(30),
                    border: Border.all(color: Colors.transparent),
                  ),
                  padding: const EdgeInsets.symmetric(horizontal: 20),
                  alignment: Alignment.centerLeft,
                  child: const TextField(
                    decoration: InputDecoration(
                      border: InputBorder.none,
                      hintText: "Message Sophia...",
                      hintStyle: TextStyle(color: Color(0xFF9292C9)),
                      isDense: true,
                    ),
                    style: TextStyle(color: Colors.white),
                  ),
                ),
              ),

              const SizedBox(width: 12),

              // Botón Micrófono con Glow (Shadow)
              Container(
                width: 50,
                height: 50,
                decoration: BoxDecoration(
                  shape: BoxShape.circle,
                  color: theme.primaryColor,
                  boxShadow: [
                    BoxShadow(
                      color: theme.primaryColor.withValues(alpha: 0.6),
                      blurRadius: 20, // shadow-lg shadow-primary/40
                      spreadRadius: 2,
                    ),
                  ],
                ),
                child: const Icon(Icons.mic, color: Colors.white),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
