import 'dart:convert';
import 'dart:ui';

import 'package:flutter/material.dart';

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
    this.color = Colors.cyan,
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

  const ActionProposalCard({
    super.key,
    required this.title,
    required this.actions,
    this.onConfirm,
    this.onReject,
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
    return ClipRRect(
      borderRadius: BorderRadius.circular(20),
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
        child: Container(
          margin: const EdgeInsets.symmetric(vertical: 8),
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: const Color(
              0xFF151B24,
            ).withValues(alpha: 0.8), // Glassmorphism
            borderRadius: BorderRadius.circular(20),
            border: Border.all(
              color: Colors.white.withValues(alpha: 0.1),
              width: 1,
            ),
            boxShadow: [
              BoxShadow(
                color: const Color(0xFF2E5CB8).withValues(alpha: 0.1),
                blurRadius: 15,
                offset: const Offset(0, 5),
              ),
            ],
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.title,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 20),
              ...widget.actions.map(_buildActionBlock),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildActionBlock(ProposedAction action) {
    final resolved = _resolved[action.id];
    return Padding(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _buildActionItem(action),
          const SizedBox(height: 12),
          if (resolved != null)
            Text(
              resolved,
              style: TextStyle(color: Colors.grey[400], fontSize: 12),
            )
          else
            _buildActionButtons(action),
        ],
      ),
    );
  }

  Widget _buildActionItem(ProposedAction action) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(10),
          decoration: BoxDecoration(
            color: action.color.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(12),
          ),
          child: Icon(action.icon, color: action.color, size: 24),
        ),
        const SizedBox(width: 16),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                action.label,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                ),
              ),
              const SizedBox(height: 2),
              Text(
                action.detail,
                style: TextStyle(color: Colors.grey[400], fontSize: 13),
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
          child: OutlinedButton(
            onPressed: busy ? null : () => _reject(action),
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.white70,
              side: const BorderSide(color: Colors.white24),
              padding: const EdgeInsets.symmetric(vertical: 10),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'No',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
            ),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: OutlinedButton(
            onPressed: busy ? null : () => _adjust(action),
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.white,
              side: const BorderSide(color: Colors.white24),
              padding: const EdgeInsets.symmetric(vertical: 10),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'Ajustar',
              style: TextStyle(fontWeight: FontWeight.w600, fontSize: 13),
            ),
          ),
        ),
        const SizedBox(width: 8),
        Expanded(
          child: ElevatedButton(
            onPressed: busy ? null : () => _confirmDirect(action),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF00D9A5),
              foregroundColor: Colors.black,
              padding: const EdgeInsets.symmetric(vertical: 10),
              elevation: 0,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
            ),
            child: const Text(
              'Hacerlo',
              style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13),
            ),
          ),
        ),
      ],
    );
  }

  Future<void> _confirmDirect(ProposedAction action) {
    return _run(action, () async {
      await widget.onConfirm?.call(
        action,
        decisionLatencyMs: _decisionLatencyMs,
      );
    }, resolvedLabel: 'Aprobado');
  }

  Future<void> _adjust(ProposedAction action) async {
    final corrected = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => _AdjustProposalDialog(action: action),
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
    final reason = await showModalBottomSheet<String>(
      context: context,
      backgroundColor: const Color(0xFF151B24),
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.vertical(top: Radius.circular(20)),
      ),
      builder: (context) {
        return SafeArea(
          child: Padding(
            padding: const EdgeInsets.fromLTRB(16, 12, 16, 24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                const Text(
                  '¿Por qué no?',
                  style: TextStyle(
                    color: Colors.white,
                    fontSize: 16,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                const SizedBox(height: 12),
                for (final option in _rejectionOptions)
                  ListTile(
                    title: Text(
                      option.$2,
                      style: const TextStyle(color: Colors.white),
                    ),
                    onTap: () => Navigator.of(context).pop(option.$1),
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
  const _AdjustProposalDialog({required this.action});

  final ProposedAction action;

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
    return AlertDialog(
      backgroundColor: const Color(0xFF151B24),
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
      title: const Text(
        'Ajustar propuesta',
        style: TextStyle(color: Colors.white),
      ),
      content: SizedBox(
        width: 360,
        child: SingleChildScrollView(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              for (final entry in _controllers.entries)
                Padding(
                  padding: const EdgeInsets.only(bottom: 12),
                  child: TextField(
                    controller: entry.value,
                    style: const TextStyle(color: Colors.white),
                    decoration: InputDecoration(
                      labelText: entry.key,
                      labelStyle: const TextStyle(color: Colors.white70),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: Colors.white24),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(12),
                        borderSide: const BorderSide(color: Color(0xFF00D9A5)),
                      ),
                    ),
                  ),
                ),
            ],
          ),
        ),
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text(
            'Cancelar',
            style: TextStyle(color: Colors.white70),
          ),
        ),
        ElevatedButton(
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
          style: ElevatedButton.styleFrom(
            backgroundColor: const Color(0xFF00D9A5),
            foregroundColor: Colors.black,
          ),
          child: const Text('Guardar'),
        ),
      ],
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
