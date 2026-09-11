import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_animate/flutter_animate.dart';

import '../../theme/design_tokens.dart';
import '../../theme/motion.dart';

final ValueNotifier<int> _visibleSheetCount = ValueNotifier(0);

class SophiaSheetBackdrop extends StatelessWidget {
  const SophiaSheetBackdrop({super.key, required this.child});
  final Widget child;

  @override
  Widget build(BuildContext context) => ValueListenableBuilder<int>(
    valueListenable: _visibleSheetCount,
    child: child,
    builder: (context, count, child) => Stack(
      fit: StackFit.expand,
      children: [
        AnimatedScale(
          scale: count > 0 ? SophiaMotion.sheetBackdropScale : 1,
          duration: SophiaMotion.resolve(context, SophiaMotion.medium),
          curve: SophiaMotion.structuralCurve,
          child: child,
        ),
        IgnorePointer(
          child: AnimatedOpacity(
            opacity: count > 0 ? SophiaMotion.sheetBackdropScrimOpacity : 0,
            duration: SophiaMotion.resolve(context, SophiaMotion.medium),
            curve: SophiaMotion.structuralCurve,
            child: ColoredBox(color: context.colors.scrim),
          ),
        ),
      ],
    ),
  );
}

class StaggeredEntry extends StatefulWidget {
  const StaggeredEntry({super.key, required this.index, required this.child});
  final int index;
  final Widget child;

  @override
  State<StaggeredEntry> createState() => _StaggeredEntryState();
}

class _StaggeredEntryState extends State<StaggeredEntry>
    with SingleTickerProviderStateMixin {
  late final AnimationController controller;

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (!mounted) return;
    controller.duration = SophiaMotion.resolve(
      context,
      SophiaMotion.medium + SophiaMotion.stagger * widget.index,
    );
    if (!controller.isAnimating && controller.value == 0) controller.forward();
  }

  @override
  void initState() {
    super.initState();
    controller = AnimationController(vsync: this);
  }

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    if (SophiaMotion.reduced(context)) return widget.child;
    final effectDelay = SophiaMotion.stagger * widget.index;
    final owner = Animate(autoPlay: false);
    const fade = FadeEffect();
    const slide = SlideEffect(begin: Offset(0, .06));
    final entry = EffectEntry(
      effect: fade,
      delay: effectDelay,
      duration: SophiaMotion.medium,
      curve: SophiaMotion.contentCurve,
      owner: owner,
    );
    return fade.build(
      context,
      slide.build(
        context,
        widget.child,
        controller,
        EffectEntry(
          effect: slide,
          delay: effectDelay,
          duration: SophiaMotion.medium,
          curve: SophiaMotion.contentCurve,
          owner: owner,
        ),
      ),
      controller,
      entry,
    );
  }
}

class MotionSwap extends StatelessWidget {
  const MotionSwap({super.key, required this.child});
  final Widget child;

  @override
  Widget build(BuildContext context) => AnimatedSize(
    duration: SophiaMotion.resolve(context, SophiaMotion.short),
    curve: SophiaMotion.structuralCurve,
    child: AnimatedSwitcher(
      duration: SophiaMotion.resolve(context, SophiaMotion.short),
      switchInCurve: SophiaMotion.contentCurve,
      switchOutCurve: SophiaMotion.structuralCurve,
      layoutBuilder: (current, previous) => Stack(
        alignment: Alignment.center,
        children: [...previous, if (current != null) current],
      ),
      child: child,
    ),
  );
}

class TactileButton extends StatefulWidget {
  const TactileButton({
    super.key,
    required this.onPressed,
    required this.child,
    this.semanticLabel,
    this.haptics = true,
  });
  final VoidCallback? onPressed;
  final Widget child;
  final String? semanticLabel;
  final bool haptics;

  @override
  State<TactileButton> createState() => _TactileButtonState();
}

class _TactileButtonState extends State<TactileButton>
    with SingleTickerProviderStateMixin {
  bool pressed = false;
  bool focused = false;
  late final AnimationController controller;

  @override
  void initState() {
    super.initState();
    controller = AnimationController(vsync: this);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    controller.duration = SophiaMotion.resolve(context, SophiaMotion.micro);
  }

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  void _setPressed(bool value) {
    if (pressed == value) return;
    setState(() => pressed = value);
    value ? controller.forward() : controller.reverse();
  }

  void _activate() {
    if (widget.onPressed == null) return;
    if (widget.haptics) HapticFeedback.lightImpact();
    widget.onPressed!();
  }

  @override
  Widget build(BuildContext context) => Semantics(
    button: true,
    enabled: widget.onPressed != null,
    label: widget.semanticLabel,
    child: FocusableActionDetector(
      enabled: widget.onPressed != null,
      onShowFocusHighlight: (value) => setState(() => focused = value),
      shortcuts: const {
        SingleActivator(LogicalKeyboardKey.enter): ActivateIntent(),
        SingleActivator(LogicalKeyboardKey.space): ActivateIntent(),
      },
      actions: {
        ActivateIntent: CallbackAction<ActivateIntent>(
          onInvoke: (_) {
            _activate();
            return null;
          },
        ),
      },
      mouseCursor: widget.onPressed == null
          ? SystemMouseCursors.basic
          : SystemMouseCursors.click,
      child: GestureDetector(
        behavior: HitTestBehavior.opaque,
        onTapDown: widget.onPressed == null ? null : (_) => _setPressed(true),
        onTapCancel: widget.onPressed == null ? null : () => _setPressed(false),
        onTapUp: widget.onPressed == null
            ? null
            : (_) {
                _setPressed(false);
                _activate();
              },
        child: AnimatedContainer(
          duration: SophiaMotion.resolve(context, SophiaMotion.short),
          curve: SophiaMotion.structuralCurve,
          decoration: BoxDecoration(
            border: Border.all(
              color: focused
                  ? context.colors.accent
                  : context.colors.surface.withValues(alpha: 0),
            ),
            borderRadius: BorderRadius.circular(SophiaRadius.control),
          ),
          child: SophiaMotion.reduced(context)
              ? ConstrainedBox(
                  constraints: const BoxConstraints(
                    minWidth: SophiaSize.minimumTapTarget,
                    minHeight: SophiaSize.minimumTapTarget,
                  ),
                  child: widget.child,
                )
              : const ScaleEffect(
                  begin: Offset(1, 1),
                  end: Offset(.97, .97),
                ).build(
                  context,
                  ConstrainedBox(
                    constraints: const BoxConstraints(
                      minWidth: SophiaSize.minimumTapTarget,
                      minHeight: SophiaSize.minimumTapTarget,
                    ),
                    child: widget.child,
                  ),
                  controller,
                  EffectEntry(
                    effect: const ScaleEffect(
                      begin: Offset(1, 1),
                      end: Offset(.97, .97),
                    ),
                    delay: Duration.zero,
                    duration: SophiaMotion.micro,
                    curve: SophiaMotion.tactileCurve,
                    owner: Animate(autoPlay: false),
                  ),
                ),
        ),
      ),
    ),
  );
}

class SophiaSwitch extends StatelessWidget {
  const SophiaSwitch({
    super.key,
    required this.value,
    required this.onChanged,
    required this.semanticLabel,
  });

  final bool value;
  final ValueChanged<bool>? onChanged;
  final String semanticLabel;

  @override
  Widget build(BuildContext context) => Semantics(
    toggled: value,
    enabled: onChanged != null,
    label: semanticLabel,
    child: TactileButton(
      semanticLabel: semanticLabel,
      onPressed: onChanged == null ? null : () => onChanged!(!value),
      child: AnimatedContainer(
        duration: SophiaMotion.resolve(context, SophiaMotion.short),
        curve: SophiaMotion.contentCurve,
        width: SophiaSpace.xxl,
        height: SophiaSpace.lg,
        padding: const EdgeInsets.all(SophiaSpace.xxs),
        decoration: BoxDecoration(
          color: value ? context.colors.accent : context.colors.elevated,
          border: Border.all(
            color: value ? context.colors.accent : context.colors.line,
          ),
          borderRadius: BorderRadius.circular(SophiaRadius.sheet),
        ),
        child: AnimatedAlign(
          duration: SophiaMotion.resolve(context, SophiaMotion.short),
          curve: SophiaMotion.contentCurve,
          alignment: value ? Alignment.centerRight : Alignment.centerLeft,
          child: Container(
            width: SophiaSpace.md,
            height: SophiaSpace.md,
            decoration: BoxDecoration(
              shape: BoxShape.circle,
              color: value ? context.colors.surface : context.colors.muted,
            ),
          ),
        ),
      ),
    ),
  );
}

class ContentSkeleton extends StatefulWidget {
  const ContentSkeleton({super.key, this.height = 72});
  final double height;

  @override
  State<ContentSkeleton> createState() => _ContentSkeletonState();
}

class _ContentSkeletonState extends State<ContentSkeleton>
    with SingleTickerProviderStateMixin {
  late final AnimationController controller;

  @override
  void initState() {
    super.initState();
    controller = AnimationController(vsync: this, duration: SophiaMotion.long);
  }

  @override
  void didChangeDependencies() {
    super.didChangeDependencies();
    if (SophiaMotion.reduced(context)) {
      controller.value = .5;
    } else if (!controller.isAnimating) {
      controller.repeat();
    }
  }

  @override
  void dispose() {
    controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) => AnimatedBuilder(
    animation: controller,
    builder: (_, _) => Container(
      height: widget.height,
      decoration: BoxDecoration(
        borderRadius: BorderRadius.circular(SophiaRadius.card),
        gradient: LinearGradient(
          begin: Alignment(-1.5 + controller.value * 3, 0),
          end: Alignment(-.5 + controller.value * 3, 0),
          colors: [
            context.colors.line,
            context.colors.ink.withValues(alpha: .06),
            context.colors.line,
          ],
        ),
      ),
    ),
  );
}

Future<T?> showSophiaSheet<T>({
  required BuildContext context,
  required WidgetBuilder builder,
}) async {
  _visibleSheetCount.value++;
  try {
    return await showModalBottomSheet<T>(
      context: context,
      isScrollControlled: true,
      showDragHandle: true,
      enableDrag: true,
      useSafeArea: true,
      sheetAnimationStyle: AnimationStyle(
        duration: SophiaMotion.resolve(context, SophiaMotion.medium),
        reverseDuration: SophiaMotion.resolve(context, SophiaMotion.short),
      ),
      builder: builder,
    );
  } finally {
    _visibleSheetCount.value--;
  }
}
