import 'dart:ui';
import 'package:flutter/material.dart';
import 'package:go_router/go_router.dart';
import 'package:google_fonts/google_fonts.dart';

class MainWrapper extends StatefulWidget {
  final StatefulNavigationShell navigationShell;

  const MainWrapper({super.key, required this.navigationShell});

  @override
  State<MainWrapper> createState() => _MainWrapperState();
}

class _MainWrapperState extends State<MainWrapper>
    with SingleTickerProviderStateMixin {
  late AnimationController _transitionController;

  @override
  void initState() {
    super.initState();
    _transitionController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 300),
    );
    _transitionController.forward();
  }

  @override
  void didUpdateWidget(MainWrapper oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.navigationShell.currentIndex !=
        widget.navigationShell.currentIndex) {
      _transitionController.forward(from: 0.0);
    }
  }

  @override
  void dispose() {
    _transitionController.dispose();
    super.dispose();
  }

  void _goBranch(int index) {
    widget.navigationShell.goBranch(
      index,
      initialLocation: index == widget.navigationShell.currentIndex,
    );
  }

  @override
  Widget build(BuildContext context) {
    // Breakpoint simple para ejemplo: < 800px es móvil
    final isDesktop = MediaQuery.of(context).size.width > 800;

    if (isDesktop) {
      return Scaffold(
        body: Row(
          children: [
            // --- DESKTOP SIDEBAR ---
            _SophiaSidebar(
              currentIndex: widget.navigationShell.currentIndex,
              onTap: _goBranch,
            ),
            Expanded(
              child: FadeTransition(
                opacity: Tween<double>(
                  begin: 0.8,
                  end: 1.0,
                ).animate(_transitionController),
                child: SlideTransition(
                  position:
                      Tween<Offset>(
                        begin: const Offset(0.02, 0),
                        end: Offset.zero,
                      ).animate(
                        CurvedAnimation(
                          parent: _transitionController,
                          curve: Curves.easeOutCubic,
                        ),
                      ),
                  child: widget.navigationShell,
                ),
              ),
            ),
          ],
        ),
      );
    } else {
      // --- MOBILE BOTTOM BAR ---
      return Scaffold(
        body: FadeTransition(
          opacity: Tween<double>(
            begin: 0.8,
            end: 1.0,
          ).animate(_transitionController),
          child: widget.navigationShell,
        ),
        bottomNavigationBar: _GlassBottomBar(
          currentIndex: widget.navigationShell.currentIndex,
          onTap: _goBranch,
        ),
      );
    }
  }
}

class _SophiaSidebar extends StatelessWidget {
  final int currentIndex;
  final Function(int) onTap;

  const _SophiaSidebar({required this.currentIndex, required this.onTap});

  @override
  Widget build(BuildContext context) {
    return Container(
      width: 250, // Ancho fijo del sidebar
      color: const Color(0xFF0B1016),
      child: Column(
        children: [
          const SizedBox(height: 40),
          // Logo Area
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.blur_on, color: Color(0xFF00B4D8), size: 32),
              const SizedBox(width: 10),
              Text(
                "SOPHIA",
                style: GoogleFonts.audiowide(
                  // Fuente futurista
                  color: Colors.white,
                  fontSize: 20,
                  letterSpacing: 2,
                ),
              ),
            ],
          ),
          const SizedBox(height: 50),
          // Menu Items
          _SidebarItem(
            icon: Icons.auto_awesome, // Icono para IA
            label: "Chat",
            isSelected: currentIndex == 0,
            onTap: () => onTap(0),
          ),
          _SidebarItem(
            icon: Icons.dashboard_rounded,
            label: "System",
            isSelected: currentIndex == 1,
            onTap: () => onTap(1),
          ),
          _SidebarItem(
            icon: Icons.settings,
            label: "Settings",
            isSelected: currentIndex == 2,
            onTap: () => onTap(2),
          ),
        ],
      ),
    );
  }
}

class _SidebarItem extends StatefulWidget {
  final IconData icon;
  final String label;
  final bool isSelected;
  final VoidCallback onTap;

  const _SidebarItem({
    required this.icon,
    required this.label,
    required this.isSelected,
    required this.onTap,
  });

  @override
  State<_SidebarItem> createState() => _SidebarItemState();
}

class _SidebarItemState extends State<_SidebarItem>
    with SingleTickerProviderStateMixin {
  bool _isHovered = false;
  late AnimationController _controller;
  late Animation<double> _glowAnimation;

  @override
  void initState() {
    super.initState();
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 200),
    );
    _glowAnimation = Tween<double>(begin: 0.0, end: 1.0).animate(_controller);
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);
    final color = widget.isSelected
        ? theme.primaryColor
        : _isHovered
        ? Colors.white70
        : Colors.grey;

    return MouseRegion(
      onEnter: (_) => setState(() {
        _isHovered = true;
        _controller.forward();
      }),
      onExit: (_) => setState(() {
        _isHovered = false;
        _controller.reverse();
      }),
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        margin: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: widget.isSelected
              ? theme.primaryColor.withValues(alpha: 0.1)
              : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
          border: Border.all(
            color: widget.isSelected
                ? theme.primaryColor.withValues(alpha: 0.3)
                : Colors.transparent,
            width: 1,
          ),
          boxShadow: widget.isSelected
              ? [
                  BoxShadow(
                    color: theme.primaryColor.withValues(alpha: 0.3),
                    blurRadius: 12,
                    spreadRadius: 0,
                  ),
                ]
              : [],
        ),
        child: ListTile(
          onTap: widget.onTap,
          leading: AnimatedBuilder(
            animation: _glowAnimation,
            builder: (context, child) {
              return Icon(
                widget.icon,
                color: color,
                shadows: widget.isSelected
                    ? [
                        Shadow(
                          color: theme.primaryColor.withValues(alpha: 0.6),
                          blurRadius: 8,
                        ),
                      ]
                    : [],
              );
            },
          ),
          title: Text(
            widget.label,
            style: TextStyle(
              color: color,
              fontWeight: widget.isSelected ? FontWeight.bold : FontWeight.w500,
              fontSize: 14,
            ),
          ),
          contentPadding: const EdgeInsets.symmetric(
            horizontal: 16,
            vertical: 4,
          ),
        ),
      ),
    );
  }
}

// Bottom Bar con efecto glassmorphism
class _GlassBottomBar extends StatelessWidget {
  final int currentIndex;
  final Function(int) onTap;

  const _GlassBottomBar({required this.currentIndex, required this.onTap});

  @override
  Widget build(BuildContext context) {
    final theme = Theme.of(context);

    return ClipRRect(
      child: BackdropFilter(
        filter: ImageFilter.blur(sigmaX: 10, sigmaY: 10),
        child: Container(
          decoration: BoxDecoration(
            color: const Color(0xFF101022).withValues(alpha: 0.8),
            border: Border(
              top: BorderSide(
                color: Colors.white.withValues(alpha: 0.1),
                width: 1,
              ),
            ),
          ),
          child: SafeArea(
            minimum: const EdgeInsets.only(bottom: 8),
            child: Padding(
              padding: const EdgeInsets.symmetric(vertical: 8),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.spaceAround,
                children: [
                  _BottomBarItem(
                    icon: Icons.chat_bubble_outline,
                    label: 'Chat',
                    isSelected: currentIndex == 0,
                    onTap: () => onTap(0),
                    color: theme.primaryColor,
                  ),

                  _BottomBarItem(
                    icon: Icons.dashboard_outlined,
                    label: 'System',
                    isSelected: currentIndex == 1,
                    onTap: () => onTap(1),
                    color: theme.primaryColor,
                  ),

                  _BottomBarItem(
                    icon: Icons.settings_outlined,
                    label: 'Settings',
                    isSelected: currentIndex == 2,
                    onTap: () => onTap(2),
                    color: theme.primaryColor,
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }
}

class _BottomBarItem extends StatelessWidget {
  final IconData icon;
  final String label;
  final bool isSelected;
  final VoidCallback onTap;
  final Color color;

  const _BottomBarItem({
    required this.icon,
    required this.label,
    required this.isSelected,
    required this.onTap,
    required this.color,
  });

  @override
  Widget build(BuildContext context) {
    return GestureDetector(
      onTap: onTap,
      child: AnimatedContainer(
        duration: const Duration(milliseconds: 200),
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: isSelected
              ? color.withValues(alpha: 0.15)
              : Colors.transparent,
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              size: 22,
              color: isSelected ? color : Colors.grey,
              shadows: isSelected
                  ? [Shadow(color: color.withValues(alpha: 0.6), blurRadius: 8)]
                  : [],
            ),
            const SizedBox(height: 2),
            Text(
              label,
              style: TextStyle(
                color: isSelected ? color : Colors.grey,
                fontSize: 10,
                fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                height: 1.2,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
