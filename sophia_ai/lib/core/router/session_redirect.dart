import 'package:sophia_ai/features/session/presentation/cubit/session_state.dart';

/// Pure redirect rules for F1 session navigation (no GoRouter loops).
String? resolveSessionRedirect({
  required SessionState session,
  required String location,
}) {
  final isSplash = location == '/splash';
  final isAuth = location == '/login' || location == '/register';
  final isOnboarding = location == '/onboarding';
  final isPublic = isSplash || isAuth;

  if (session is SessionInitial || session is SessionLoading) {
    return isSplash ? null : '/splash';
  }

  if (session is SessionFailure && session.tokenRetained) {
    return isSplash ? null : '/splash';
  }

  if (session is SessionUnauthenticated ||
      (session is SessionFailure && !session.tokenRetained)) {
    if (isAuth) return null;
    return '/login';
  }

  if (session is SessionAuthenticated) {
    final needsOnboarding = !session.profile.onboardingCompleted;
    if (needsOnboarding) {
      return isOnboarding ? null : '/onboarding';
    }
    if (isPublic || isOnboarding) return '/chat';
    return null;
  }

  return null;
}
