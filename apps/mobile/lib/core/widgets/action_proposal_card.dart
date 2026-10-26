import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../theme/design_tokens.dart';
import 'motion/motion_widgets.dart';

class ProposedAction {
  final String id;
  final IconData icon;
  final String label;
  final String detail;
  final Color color;
  final Map<String, dynamic> proposedInput;

  const ProposedAction({
    required this.id,
    required this.icon,
    required this.label,
    required this.detail,
    required this.color,
    this.proposedInput = const <String, dynamic>{},
  });
}

typedef ConfirmProposedAction =
    Future<void> Function(
      ProposedAction action, {
      Map<String, dynamic>? correctedInput,
      required int decisionLatencyMs,
    });

typedef RejectProposedAction =
    Future<void> Function(
      ProposedAction action, {
      required String rejectionReason,
      required int decisionLatencyMs,
    });

class ActionProposalCard extends StatefulWidget {
  final String title;
  final List<ProposedAction> actions;
  final ConfirmProposedAction? onConfirm;
  final RejectProposedAction? onReject;
  final Object? heroTag;

  const ActionProposalCard({
    super.key,
    required this.title,
    required this.actions,
    this.onConfirm,
    this.onReject,
    this.heroTag,
  });

  @override
  State<ActionProposalCard> createState() => _ActionProposalCardState();
}

class _ActionProposalCardState extends State<ActionProposalCard> {
  static const _rejectionOptions = <(String, String)>[
    ('wrong_time', 'Hora incorrecta'),
    ('not_needed', 'No hace falta'),
    ('wrong_person', 'Persona incorrecta'),
    ('other', 'Otro'),
  ];

  late final DateTime _shownAt;
  final Set<String> _busy = <String>{};
  final Map<String, String> _resolved = <String, String>{};

  @override
  void initState() {
    super.initState();
    _shownAt = DateTime.now();
  }

  int get _decisionLatencyMs {
    final ms = DateTime.now().difference(_shownAt).inMilliseconds;
    return ms < 0 ? 0 : ms;
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: SophiaSpace.xs),
      padding: const EdgeInsets.all(SophiaSpace.lg),
      decoration: BoxDecoration(
        color: context.colors.elevated,
        borderRadius: BorderRadius.circular(SophiaRadius.card),
        border: Border.all(color: context.colors.line),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(widget.title, style: Theme.of(context).textTheme.titleMedium),
          const SizedBox(height: SophiaSpace.lg),
          ...widget.actions.map(_buildActionBlock),
        ],
      ),
    );
  }

  Widget _buildActionBlock(ProposedAction action) {
    final resolved = _resolved[action.id];
    return Padding(
      padding: const EdgeInsets.only(bottom: SophiaSpace.md),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildActionItem(action),
          const SizedBox(height: SophiaSpace.sm),
          MotionSwap(
            child: resolved != null
                ? Row(
                    key: ValueKey(resolved),
                    children: [
                      Icon(
                        Icons.check,
                        size: SophiaSpace.md,
                        color: context.colors.positive,
                      ),
                      const SizedBox(width: SophiaSpace.xs),
                      Text(
                        resolved,
                        style: TextStyle(color: context.colors.softInk),
                      ),
                    ],
                  )
                : KeyedSubtree(
                    key: const ValueKey('actions'),
                    child: _buildActionButtons(action),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildActionItem(ProposedAction action) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(SophiaSpace.sm),
          decoration: BoxDecoration(
            color: action.color.withValues(alpha: SophiaOpacity.subtle),
            borderRadius: BorderRadius.circular(SophiaRadius.control),
          ),
          child: Icon(action.icon, color: action.color, size: SophiaSpace.lg),
        ),
        const SizedBox(width: SophiaSpace.md),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                action.label,
                style: Theme.of(
                  context,
                ).textTheme.bodyMedium?.copyWith(fontWeight: FontWeight.w600),
              ),
              const SizedBox(height: SophiaSpace.xxs),
              Text(
                action.detail,
                style: Theme.of(
                  context,
                ).textTheme.bodySmall?.copyWith(color: context.colors.softInk),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildActionButtons(ProposedAction action) {
    final busy = _busy.contains(action.id);
    return Row(
      children: [
        Expanded(
          child: TactileButton(
            haptics: false,
            onPressed: busy ? null : () => _reject(action),
            child: IgnorePointer(
              child: OutlinedButton(
                onPressed: busy ? null : () {},
                style: OutlinedButton.styleFrom(
                  foregroundColor: context.colors.softInk,
                  side: BorderSide(color: context.colors.line),
                  padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(SophiaRadius.control),
                  ),
                ),
                child: Text(
                  'No',
                  style: Theme.of(context).textTheme.labelLarge,
                ),
              ),
            ),
          ),
        ),
        const SizedBox(width: SophiaSpace.xs),
        Expanded(
          child: TactileButton(
            onPressed: busy ? null : () => _adjust(action),
            child: IgnorePointer(
              child: OutlinedButton(
                onPressed: busy ? null : () {},
                style: OutlinedButton.styleFrom(
                  foregroundColor: context.colors.ink,
                  side: BorderSide(color: context.colors.line),
                  padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(SophiaRadius.control),
                  ),
                ),
                child: Text(
                  'Ajustar',
                  style: Theme.of(context).textTheme.labelLarge,
                ),
              ),
            ),
          ),
        ),
        const SizedBox(width: SophiaSpace.xs),
        Expanded(
          child: TactileButton(
            haptics: false,
            onPressed: busy ? null : () => _confirmDirect(action),
            child: IgnorePointer(
              child: ElevatedButton(
                onPressed: busy ? null : () {},
                style: ElevatedButton.styleFrom(
                  backgroundColor: context.colors.accent,
                  foregroundColor: context.colors.surface,
                  padding: const EdgeInsets.symmetric(vertical: SophiaSpace.sm),
                  elevation: 0,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(SophiaRadius.control),
                  ),
                ),
                child: Text(
                  'Hacerlo',
                  style: Theme.of(context).textTheme.labelLarge?.copyWith(
                    color: context.colors.surface,
                  ),
                ),
              ),
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _confirmDirect(ProposedAction action) {
    HapticFeedback.mediumImpact();
    return _run(action, () async {
      await widget.onConfirm?.call(
        action,
        decisionLatencyMs: _decisionLatencyMs,
      );
    }, resolvedLabel: 'Aprobado');
  }

  Future<void> _adjust(ProposedAction action) async {
    final corrected = await showSophiaSheet<Map<String, dynamic>>(
      context: context,
      builder: (context) =>
          _AdjustProposalDialog(action: action, heroTag: widget.heroTag),
    );
    if (corrected == null || !mounted) return;
    await _run(action, () async {
      await widget.onConfirm?.call(
        action,
        correctedInput: corrected,
        decisionLatencyMs: _decisionLatencyMs,
      );
    }, resolvedLabel: 'Aprobado con ajustes');
  }

  Future<void> _reject(ProposedAction action) async {
    HapticFeedback.lightImpact();
    final reason = await showSophiaSheet<String>(
      context: context,
      builder: (context) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(
              SophiaSpace.md,
              SophiaSpace.sm,
              SophiaSpace.md,
              SophiaSpace.lg,
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '¿Por qué no?',
                  style: Theme.of(context).textTheme.titleSmall,
                ),
                const SizedBox(height: SophiaSpace.sm),
                for (final option in _rejectionOptions)
                  TactileButton(
                    onPressed: () => Navigator.of(context).pop(option.$1),
                    child: IgnorePointer(
                      child: ListTile(title: Text(option.$2), onTap: () {}),
                    ),
                  ),
              ],
            ),
          ),
        );
      },
    );
    if (reason == null || !mounted) return;
    await _run(action, () async {
      await widget.onReject?.call(
        action,
        rejectionReason: reason,
        decisionLatencyMs: _decisionLatencyMs,
      );
    }, resolvedLabel: 'Rechazado');
  }

  Future<void> _run(
    ProposedAction action,
    Future<void> Function() work, {
    required String resolvedLabel,
  }) async {
    setState(() => _busy.add(action.id));
    try {
      await work();
      if (!mounted) return;
      setState(() {
        _busy.remove(action.id);
        _resolved[action.id] = resolvedLabel;
      });
    } catch (_) {
      if (!mounted) return;
      setState(() => _busy.remove(action.id));
    }
  }
}

class _AdjustProposalDialog extends StatefulWidget {
  const _AdjustProposalDialog({required this.action, required this.heroTag});

  final ProposedAction action;
  final Object? heroTag;

  @override
  State<_AdjustProposalDialog> createState() => _AdjustProposalDialogState();
}

class _AdjustProposalDialogState extends State<_AdjustProposalDialog> {
  late final Map<String, TextEditingController> _controllers;

  @override
  void initState() {
    super.initState();
    _controllers = {
      for (final entry in widget.action.proposedInput.entries)
        entry.key: TextEditingController(text: _stringify(entry.value)),
    };
  }

  @override
  void dispose() {
    for (final controller in _controllers.values) {
      controller.dispose();
    }
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final dialog = AlertDialog(
      backgroundColor: context.colors.elevated,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(SophiaRadius.sheet),
      ),
      title: Text(
        'Ajustar propuesta',
        style: TextStyle(color: context.colors.ink),
      ),
      content: SizedBox(
        width: SophiaSize.messageMaxWidth,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              for (final entry in _controllers.entries)
                Padding(
                  padding: const EdgeInsets.only(bottom: SophiaSpace.sm),
                  child: TextField(
                    controller: entry.value,
                    style: TextStyle(color: context.colors.ink),
                    decoration: InputDecoration(
                      labelText: entry.key,
                      labelStyle: TextStyle(color: context.colors.softInk),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(
                          SophiaRadius.control,
                        ),
                        borderSide: BorderSide(color: context.colors.line),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(
                          SophiaRadius.control,
                        ),
                        borderSide: BorderSide(color: context.colors.accent),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
      actions: [
        TactileButton(
          onPressed: () => Navigator.of(context).pop(),
          child: IgnorePointer(
            child: TextButton(
              onPressed: () {},
              child: Text(
                'Cancelar',
                style: TextStyle(color: context.colors.softInk),
              ),
            ),
          ),
        ),
        TactileButton(
          onPressed: () {
            final corrected = <String, dynamic>{};
            for (final entry in widget.action.proposedInput.entries) {
              corrected[entry.key] = _parseField(
                _controllers[entry.key]?.text ?? '',
                entry.value,
              );
            }
            Navigator.of(context).pop(corrected);
          },
          child: IgnorePointer(
            child: ElevatedButton(
              onPressed: () {},
              style: ElevatedButton.styleFrom(
                backgroundColor: context.colors.accent,
                foregroundColor: context.colors.surface,
              ),
              child: const Text('Guardar'),
            ),
          ),
        ),
      ],
    );
    if (widget.heroTag == null) return dialog;
    return Hero(
      tag: widget.heroTag!,
      transitionOnUserGestures: true,
      child: Material(
        color: context.colors.surface.withValues(alpha: 0),
        child: dialog,
      ),
    );
  }

  String _stringify(dynamic value) {
    if (value == null) return '';
    if (value is String) return value;
    if (value is Map || value is List) {
      return jsonEncode(value);
    }
    return value.toString();
  }

  dynamic _parseField(String text, dynamic original) {
    final trimmed = text.trim();
    if (original is bool) {
      return trimmed.toLowerCase() == 'true';
    }
    if (original is int) {
      return int.tryParse(trimmed) ?? original;
    }
    if (original is num) {
      return num.tryParse(trimmed) ?? original;
    }
    if (original is List || original is Map) {
      try {
        return jsonDecode(trimmed);
      } catch (_) {
        return original;
      }
    }
    return trimmed;
  }
}
