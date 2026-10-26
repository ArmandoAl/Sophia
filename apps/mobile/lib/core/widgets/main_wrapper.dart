import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:go_router/go_router.dart';

import '../theme/design_tokens.dart';
import '../theme/motion.dart';
import 'motion/motion_widgets.dart';

class MainWrapper extends StatefulWidget {
  const MainWrapper({super.key, required this.navigationShell});
  final StatefulNavigationShell navigationShell;

  @override
  State<MainWrapper> createState() => _MainWrapperState();
}

class _MainWrapperState extends State<MainWrapper> {
  final scrollControllers = List.generate(3, (_) => ScrollController());

  @override
  void dispose() {
    for (final value in scrollControllers) {
      value.dispose();
    }
    super.dispose();
  }

  void go(int index) {
    if (index == widget.navigationShell.currentIndex) {
      final scroll = scrollControllers[index];
      if (scroll.hasClients) {
        scroll.animateTo(
          0,
          duration: SophiaMotion.resolve(context, SophiaMotion.medium),
          curve: SophiaMotion.structuralCurve,
        );
      }
      return;
    }
    HapticFeedback.selectionClick();
    widget.navigationShell.goBranch(index);
  }

  @override
  Widget build(BuildContext context) {
    final content = PrimaryScrollController(
      controller: scrollControllers[widget.navigationShell.currentIndex],
      child: widget.navigationShell,
    );
    if (MediaQuery.sizeOf(context).width > SophiaSize.navigationBreakpoint) {
      return Scaffold(
        body: Row(
          children: [
            _Sidebar(index: widget.navigationShell.currentIndex, onTap: go),
            Expanded(child: content),
          ],
        ),
      );
    }
    return Scaffold(
      body: content,
      bottomNavigationBar: _BottomBar(
        index: widget.navigationShell.currentIndex,
        onTap: go,
      ),
    );
  }
}

class _Sidebar extends StatelessWidget {
  const _Sidebar({required this.index, required this.onTap});
  final int index;
  final ValueChanged<int> onTap;

  @override
  Widget build(BuildContext context) => Container(
    width: SophiaSize.sidebar,
    color: context.colors.elevated,
    padding: const EdgeInsets.all(SophiaSpace.lg),
    child: Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Text('SOFÍA', style: Theme.of(context).textTheme.titleLarge),
        const SizedBox(height: SophiaSpace.xl),
        for (final (i, icon, label) in const [
          (0, Icons.chat_bubble_outline, 'Chat'),
          (1, Icons.today_outlined, 'Hoy'),
          (2, Icons.settings_outlined, 'Ajustes'),
        ])
          TactileButton(
            haptics: false,
            onPressed: () => onTap(i),
            child: Padding(
              padding: const EdgeInsets.all(SophiaSpace.sm),
              child: Row(
                children: [
                  Icon(
                    icon,
                    color: i == index
                        ? context.colors.accent
                        : context.colors.softInk,
                  ),
                  const SizedBox(width: SophiaSpace.sm),
                  Text(
                    label,
                    style: TextStyle(
                      color: i == index
                          ? context.colors.accent
                          : context.colors.ink,
                    ),
                  ),
                ],
              ),
            ),
          ),
      ],
    ),
  );
}

class _BottomBar extends StatelessWidget {
  const _BottomBar({required this.index, required this.onTap});
  final int index;
  final ValueChanged<int> onTap;

  @override
  Widget build(BuildContext context) {
    const items = [
      (Icons.chat_bubble_outline, 'Chat'),
      (Icons.today_outlined, 'Hoy'),
      (Icons.settings_outlined, 'Ajustes'),
    ];
    return Material(
      color: context.colors.elevated,
      child: SafeArea(
        top: false,
        child: SizedBox(
          height: SophiaSize.bottomBar,
          child: LayoutBuilder(
            builder: (context, constraints) => Stack(
              children: [
                AnimatedPositioned(
                  duration: SophiaMotion.resolve(context, SophiaMotion.short),
                  curve: SophiaMotion.structuralCurve,
                  left:
                      constraints.maxWidth / items.length * index +
                      constraints.maxWidth / items.length * .25,
                  bottom: 5,
                  width: constraints.maxWidth / items.length * .5,
                  height: SophiaSize.navigationIndicator,
                  child: DecoratedBox(
                    decoration: BoxDecoration(
                      color: context.colors.accent,
                      borderRadius: BorderRadius.circular(SophiaRadius.control),
                    ),
                  ),
                ),
                Row(
                  children: [
                    for (var i = 0; i < items.length; i++)
                      Expanded(
                        child: _NavItem(
                          icon: items[i].$1,
                          label: items[i].$2,
                          selected: i == index,
                          onTap: () => onTap(i),
                        ),
                      ),
                  ],
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }
}

class _NavItem extends StatelessWidget {
  const _NavItem({
    required this.icon,
    required this.label,
    required this.selected,
    required this.onTap,
  });
  final IconData icon;
  final String label;
  final bool selected;
  final VoidCallback onTap;

  @override
  Widget build(BuildContext context) => TactileButton(
    haptics: false,
    semanticLabel: label,
    onPressed: onTap,
    child: Column(
      mainAxisAlignment: MainAxisAlignment.center,
      children: [
        AnimatedScale(
          scale: selected ? SophiaMotion.selectedTabScale : 1,
          duration: SophiaMotion.resolve(context, SophiaMotion.short),
          curve: SophiaMotion.contentCurve,
          child: TweenAnimationBuilder<Color?>(
            tween: ColorTween(
              begin: selected ? context.colors.accent : context.colors.muted,
              end: selected ? context.colors.accent : context.colors.muted,
            ),
            duration: SophiaMotion.resolve(context, SophiaMotion.short),
            curve: SophiaMotion.contentCurve,
            builder: (_, color, _) => Icon(icon, color: color),
          ),
        ),
        const SizedBox(height: SophiaSpace.xxs),
        AnimatedOpacity(
          opacity: selected ? 1 : .55,
          duration: SophiaMotion.resolve(context, SophiaMotion.short),
          child: Text(
            label,
            style: Theme.of(context).textTheme.labelMedium?.copyWith(
              color: selected ? context.colors.accent : context.colors.softInk,
            ),
          ),
        ),
      ],
    ),
  );
}

class LateralBranchContainer extends StatefulWidget {
  const LateralBranchContainer({
    super.key,
    required this.currentIndex,
    required this.children,
  });

  final int currentIndex;
  final List<Widget> children;

  @override
  State<LateralBranchContainer> createState() => _LateralBranchContainerState();
}

class _LateralBranchContainerState extends State<LateralBranchContainer> {
  late int previousIndex = widget.currentIndex;

  @override
  void didUpdateWidget(covariant LateralBranchContainer oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.currentIndex != widget.currentIndex) {
      previousIndex = oldWidget.currentIndex;
    }
  }

  @override
  Widget build(BuildContext context) => Stack(
    fit: StackFit.expand,
    children: [
      for (var index = 0; index < widget.children.length; index++)
        _branch(context, index),
    ],
  );

  Widget _branch(BuildContext context, int index) {
    final selected = index == widget.currentIndex;
    final outgoing = index == previousIndex && !selected;
    return IgnorePointer(
      ignoring: !selected,
      child: ExcludeSemantics(
        excluding: !selected,
        child: TickerMode(
          enabled: selected,
          child: AnimatedOpacity(
            opacity: selected ? 1 : 0,
            duration: SophiaMotion.resolve(context, SophiaMotion.short),
            curve: SophiaMotion.contentCurve,
            child: AnimatedScale(
              scale: selected
                  ? 1
                  : outgoing
                  ? SophiaMotion.lateralOutgoingScale
                  : SophiaMotion.lateralIncomingScale,
              duration: SophiaMotion.resolve(context, SophiaMotion.short),
              curve: SophiaMotion.contentCurve,
              child: widget.children[index],
            ),
          ),
        ),
      ),
    );
  }
}
