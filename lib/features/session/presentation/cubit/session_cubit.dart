import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/models/users/ai_settings.dart';
import '../../../../core/models/users/me_response.dart';
import '../../../../core/models/users/user_profile.dart';
import '../../../../core/network/api_exception.dart';
import '../../../../core/storage/token_storage.dart';
import '../../../auth/domain/auth_repository.dart';
import '../../../users/domain/user_repository.dart';
import 'session_state.dart';

/// Source of truth for authenticated session data.
class SessionCubit extends Cubit<SessionState> {
  SessionCubit({
    required TokenStorage tokenStorage,
    required UserRepository userRepository,
    required AuthRepository authRepository,
  }) : _tokenStorage = tokenStorage,
       _userRepository = userRepository,
       _authRepository = authRepository,
       super(const SessionInitial());

  final TokenStorage _tokenStorage;
  final UserRepository _userRepository;
  final AuthRepository _authRepository;

  /// App start: read token → getMe or unauthenticated.
  Future<void> bootstrap() async {
    emit(const SessionLoading());
    final token = await _tokenStorage.readToken();
    if (token == null || token.isEmpty) {
      emit(const SessionUnauthenticated());
      return;
    }
    await _hydrateFromMe();
  }

  /// After successful login (token already stored by AuthRepository).
  Future<void> establishFromStoredToken() async {
    emit(const SessionLoading());
    await _hydrateFromMe();
  }

  void applyMe(MeResponse me) {
    emit(
      SessionAuthenticated(
        user: me.user,
        profile: me.profile,
        aiSettings: me.aiSettings,
      ),
    );
  }

  void applyProfile(UserProfile profile) {
    final current = state;
    if (current is SessionAuthenticated) {
      emit(current.copyWith(profile: profile));
    }
  }

  void applyAiSettings(AiSettings settings) {
    final current = state;
    if (current is SessionAuthenticated) {
      emit(current.copyWith(aiSettings: settings));
    }
  }

  Future<void> logout() async {
    emit(const SessionLoading());
    await _authRepository.logout();
    emit(const SessionUnauthenticated());
  }

  Future<void> _hydrateFromMe() async {
    try {
      final me = await _userRepository.getMe();
      applyMe(me);
    } on ApiException catch (e) {
      if (e.isUnauthorized) {
        await _tokenStorage.clearToken();
        emit(const SessionUnauthenticated());
        return;
      }
      emit(SessionFailure(message: e.message, error: e, tokenRetained: true));
    } catch (e) {
      emit(
        SessionFailure(message: e.toString(), error: e, tokenRetained: true),
      );
    }
  }
}
