import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/config/feature_flags.dart';
import '../../../../core/di/service_locator.dart';
import '../../../../core/theme/design_tokens.dart';
import '../../../../core/theme/motion.dart';
import '../../../../core/widgets/action_proposal_card.dart';
import '../../../../core/widgets/map_location_card.dart';
import '../../../../core/widgets/motion/motion_widgets.dart';
import '../../../../core/widgets/schedule_conflicts_card.dart';
import '../../../actions/domain/action_proposals_repository.dart';
import '../../domain/entities/chat_message.dart';
import '../cubit/chat_message_cubit.dart';
import '../cubit/chat_message_state.dart';

class ChatPage extends StatelessWidget {
  const ChatPage({super.key, this.cubit, this.autoLoad = true});
  final ChatMessageCubit? cubit;
  final bool autoLoad;

  @override
  Widget build(BuildContext context) => BlocProvider(
    create: (_) {
      final value = cubit ?? sl<ChatMessageCubit>();
      if (autoLoad) value.load();
      return value;
    },
    child: const _ChatView(),
  );
}

class _ChatView extends StatelessWidget {
  const _ChatView();

  @override
  Widget build(BuildContext context) => Scaffold(
    appBar: AppBar(
      title: Row(
        children: [
          CircleAvatar(
            radius: SophiaSpace.md,
            backgroundColor: context.colors.accent.withValues(
              alpha: SophiaOpacity.subtle,
            ),
            child: Icon(
              Icons.auto_awesome,
              size: SophiaSpace.md,
              color: context.colors.accent,
            ),
          ),
          const SizedBox(width: SophiaSpace.sm),
          Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              const Text('Sofía'),
              Text(
                'Presente',
                style: Theme.of(context).textTheme.labelMedium?.copyWith(
                  color: context.colors.positive,
                ),
              ),
            ],
          ),
        ],
      ),
    ),
    body: Column(
      children: [
        Expanded(
          child: BlocBuilder<ChatMessageCubit, ChatMessageState>(
            builder: (context, state) {
              if (state.isLoading ||
                  (state.conversation == null &&
                      state.messages.isEmpty &&
                      state.errorMessage == null)) {
                return const MotionSwap(
                  child: Padding(
                    key: Key('chat-loading'),
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
              if (state.errorMessage != null && state.messages.isEmpty) {
                return MotionSwap(
                  child: _ChatErrorState(
                    key: const ValueKey('chat-error'),
                    message: state.errorMessage!,
                  ),
                );
              }
              if (state.messages.isEmpty && !state.isTyping) {
                return MotionSwap(
                  child: Center(
                    key: const ValueKey('chat-empty'),
                    child: Text(
                      'Empieza una conversación con Sofía.',
                      style: TextStyle(color: context.colors.softInk),
                    ),
                  ),
                );
              }
              return MotionSwap(
                child: ListView.builder(
                  key: const ValueKey('chat-content'),
                  reverse: true,
                  padding: const EdgeInsets.all(SophiaSpace.md),
                  itemCount:
                      state.messages.length +
                      (state.errorMessage == null ? 0 : 1) +
                      (state.isTyping ? 1 : 0),
                  itemBuilder: (context, index) {
                    if (state.isTyping && index == 0) {
                      return const _ThinkingPulse();
                    }
                    final adjusted = index - (state.isTyping ? 1 : 0);
                    if (adjusted == 0 && state.errorMessage != null) {
                      return _InlineChatError(message: state.errorMessage!);
                    }
                    final messageIndex =
                        adjusted - (state.errorMessage == null ? 0 : 1);
                    return _MessageItem(
                      key: ValueKey(state.messages[messageIndex].id),
                      message: state.messages[messageIndex],
                    );
                  },
                ),
              );
            },
          ),
        ),
        const _ChatInput(),
      ],
    ),
  );
}

class _MessageItem extends StatelessWidget {
  const _MessageItem({super.key, required this.message});
  final ChatMessage message;

  @override
  Widget build(BuildContext context) {
    final mine = message.isUser;
    final proposal = message.type == MessageType.actionProposal;
    final content = switch (message.type) {
      MessageType.scheduleConflict => const ScheduleConflictCard(),
      MessageType.mapLocation => const MapLocationCard(),
      MessageType.actionProposal => _proposal(context),
      _ => Text(message.text),
    };
    final bubble = Align(
      alignment: mine ? Alignment.centerRight : Alignment.centerLeft,
      child: Container(
        constraints: const BoxConstraints(maxWidth: SophiaSize.messageMaxWidth),
        margin: const EdgeInsets.only(bottom: SophiaSpace.sm),
        padding: EdgeInsets.all(
          message.type == MessageType.actionProposal
              ? SophiaSpace.xxs
              : SophiaSpace.md,
        ),
        decoration: BoxDecoration(
          color: mine
              ? context.colors.accent.withValues(alpha: SophiaOpacity.subtle)
              : context.colors.elevated,
          border: Border.all(
            color: mine
                ? context.colors.accent.withValues(alpha: SophiaOpacity.quiet)
                : context.colors.line,
          ),
          borderRadius: BorderRadius.only(
            topLeft: const Radius.circular(SophiaRadius.card),
            topRight: const Radius.circular(SophiaRadius.card),
            bottomLeft: Radius.circular(
              mine ? SophiaRadius.card : SophiaRadius.control,
            ),
            bottomRight: Radius.circular(
              mine ? SophiaRadius.control : SophiaRadius.card,
            ),
          ),
        ),
        child: content,
      ),
    );
    final baseDuration = mine ? SophiaMotion.short : SophiaMotion.medium;
    final duration = mine ? baseDuration : baseDuration + SophiaMotion.stagger;
    final offset = proposal
        ? SophiaMotion.proposalMessageOffset
        : mine
        ? SophiaMotion.userMessageOffset
        : SophiaMotion.assistantMessageOffset;
    return TweenAnimationBuilder<double>(
      tween: Tween(begin: 0, end: 1),
      duration: SophiaMotion.resolve(context, duration),
      curve: mine
          ? SophiaMotion.contentCurve
          : SophiaMotion.assistantMessageCurve,
      child: bubble,
      builder: (_, value, child) {
        final opacity = value.clamp(0.0, 1.0).toDouble();
        return Opacity(
          opacity: opacity,
          child: Transform.translate(
            offset: Offset(0, (1 - value) * offset),
            child: Transform.scale(
              scale: proposal
                  ? SophiaMotion.proposalEntryScale +
                        (1 - SophiaMotion.proposalEntryScale) * value
                  : 1,
              alignment: Alignment.bottomLeft,
              child: child,
            ),
          ),
        );
      },
    );
  }

  Widget _proposal(BuildContext context) {
    final title = message.metadata?['title'] as String? ?? 'Propuesta';
    final source = message.metadata?['actions'] as List? ?? const [];
    final actions = source.map((raw) {
      final action = raw as Map;
      return ProposedAction(
        id: action['id'] as String,
        icon: _icon(action['icon'] as String?),
        label: action['label'] as String,
        detail: action['detail'] as String,
        color: context.colors.accent,
        proposedInput: Map<String, dynamic>.from(
          (action['proposed_input'] as Map?) ?? const {},
        ),
      );
    }).toList();
    final executionEnabled = sl<FeatureFlags>().aiActionExecutionEnabled;
    final heroTag = 'proposal-${message.id}';
    return Hero(
      tag: heroTag,
      transitionOnUserGestures: true,
      child: Material(
        color: context.colors.surface.withValues(alpha: 0),
        child: ActionProposalCard(
          heroTag: heroTag,
          title: title,
          actions: actions,
          onConfirm:
              (action, {correctedInput, required decisionLatencyMs}) async {
                if (!executionEnabled) return;
                final repository = sl<ActionProposalsRepository>();
                await repository.confirm(
                  action.id,
                  correctedInput: correctedInput,
                  decisionLatencyMs: decisionLatencyMs,
                );
                await repository.execute(action.id);
              },
          onReject:
              (
                action, {
                required rejectionReason,
                required decisionLatencyMs,
              }) async {
                if (!executionEnabled) return;
                await sl<ActionProposalsRepository>().reject(
                  action.id,
                  rejectionReason: rejectionReason,
                  decisionLatencyMs: decisionLatencyMs,
                );
              },
        ),
      ),
    );
  }

  IconData _icon(String? value) => switch (value) {
    'lightbulb' => Icons.lightbulb_outline,
    'music_note' => Icons.music_note,
    'ac_unit' => Icons.ac_unit,
    'tv' => Icons.tv,
    'coffee' => Icons.coffee,
    'sports_esports' => Icons.sports_esports,
    'weekend' => Icons.weekend_outlined,
    'door_front' => Icons.door_front_door_outlined,
    _ => Icons.check_circle_outline,
  };
}

class _ThinkingPulse extends StatefulWidget {
  const _ThinkingPulse();

  @override
  State<_ThinkingPulse> createState() => _ThinkingPulseState();
}

class _ThinkingPulseState extends State<_ThinkingPulse>
    with SingleTickerProviderStateMixin {
  late final AnimationController controller;

  @override
  void initState() {
    super.initState();
    controller = AnimationController(
      vsync: this,
      lowerBound: SophiaMotion.thinkingOpacityMin,
      upperBound: SophiaMotion.thinkingOpacityMax,
    );
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    final duration = SophiaMotion.resolve(context, SophiaMotion.long);
    if (duration == Duration.zero) {
      controller
        ..stop()
        ..value = SophiaMotion.thinkingOpacityMax;
      return;
    }
    controller.duration = duration;
    if (!controller.isAnimating) controller.repeat(reverse: true);
  }

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => Align(
    alignment: Alignment.centerLeft,
    child: FadeTransition(
      opacity: controller,
      child: Container(
        width: SophiaSize.thinkingWidth,
        height: SophiaSize.thinkingHeight,
        margin: const EdgeInsets.only(bottom: SophiaSpace.sm),
        decoration: BoxDecoration(
          color: context.colors.elevated,
          borderRadius: BorderRadius.circular(SophiaRadius.card),
        ),
        child: Center(
          child: Container(
            width: SophiaSize.thinkingIndicator,
            height: SophiaSize.thinkingIndicator,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: context.colors.accent,
            ),
          ),
        ),
      ),
    ),
  );
}

class _ChatInput extends StatefulWidget {
  const _ChatInput();
  @override
  State<_ChatInput> createState() => _ChatInputState();
}

class _ChatInputState extends State<_ChatInput> {
  final controller = TextEditingController();

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  void send() {
    final text = controller.text.trim();
    if (text.isEmpty) return;
    context.read<ChatMessageCubit>().sendMessage(text);
    controller.clear();
    setState(() {});
  }

  @override
  Widget build(BuildContext context) =>
      BlocBuilder<ChatMessageCubit, ChatMessageState>(
        builder: (context, state) => Material(
          color: context.colors.elevated,
          child: SafeArea(
            top: false,
            child: Padding(
              padding: const EdgeInsets.all(SophiaSpace.md),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.end,
                children: [
                  Expanded(
                    child: AnimatedSize(
                      duration: SophiaMotion.resolve(
                        context,
                        SophiaMotion.short,
                      ),
                      curve: SophiaMotion.structuralCurve,
                      alignment: Alignment.bottomCenter,
                      child: TextField(
                        controller: controller,
                        enabled: !state.isTyping && !state.isLoading,
                        minLines: 1,
                        maxLines: 5,
                        textInputAction: TextInputAction.newline,
                        onChanged: (_) => setState(() {}),
                        decoration: const InputDecoration(
                          hintText: 'Escribe a Sofía…',
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: SophiaSpace.sm),
                  TactileButton(
                    semanticLabel: 'Enviar mensaje',
                    onPressed: state.isTyping || controller.text.trim().isEmpty
                        ? null
                        : send,
                    child: CircleAvatar(
                      backgroundColor: context.colors.accent,
                      foregroundColor: context.colors.surface,
                      child: const Icon(Icons.arrow_upward),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      );
}

class _ChatErrorState extends StatelessWidget {
  const _ChatErrorState({super.key, required this.message});
  final String message;
  @override
  Widget build(BuildContext context) => Center(
    child: Column(
      mainAxisSize: MainAxisSize.min,
      children: [
        Text(message, style: TextStyle(color: context.colors.softInk)),
        const SizedBox(height: SophiaSpace.sm),
        TactileButton(
          onPressed: context.read<ChatMessageCubit>().retry,
          child: const Padding(
            padding: EdgeInsets.all(SophiaSpace.sm),
            child: Text('Reintentar'),
          ),
        ),
      ],
    ),
  );
}

class _InlineChatError extends StatelessWidget {
  const _InlineChatError({required this.message});
  final String message;
  @override
  Widget build(BuildContext context) => TactileButton(
    onPressed: context.read<ChatMessageCubit>().retry,
    child: IgnorePointer(
      child: TextButton.icon(
        onPressed: () {},
        icon: const Icon(Icons.refresh),
        label: Text(message),
      ),
    ),
  );
}
