import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/models/auth/login_request.dart';
import '../../../../core/models/auth/register_request.dart';
import '../../../../core/network/api_exception.dart';
import '../../../session/presentation/cubit/session_cubit.dart';
import '../../../session/presentation/cubit/session_state.dart';
import '../../domain/auth_repository.dart';
import 'auth_state.dart';

/// Login / register form state. Session data lives in [SessionCubit].
///
/// **Register policy:** no auto-login. Register returns [AuthRegisterSuccess]
/// without storing a token; the user must call [login] afterwards.
class AuthCubit extends Cubit<AuthState> {
  AuthCubit({
    required AuthRepository authRepository,
    required SessionCubit sessionCubit,
  }) : _authRepository = authRepository,
       _sessionCubit = sessionCubit,
       super(const AuthInitial());

  final AuthRepository _authRepository;
  final SessionCubit _sessionCubit;

  Future<void> register({
    required String name,
    required String email,
    required String password,
  }) async {
    emit(const AuthLoading());
    try {
      final user = await _authRepository.register(
        RegisterRequest(name: name, email: email, password: password),
      );
      emit(AuthRegisterSuccess(user));
    } on ApiException catch (e) {
      emit(AuthFailure(message: e.message, exception: e));
    } catch (e) {
      emit(AuthFailure(message: e.toString()));
    }
  }

  Future<void> login({required String email, required String password}) async {
    emit(const AuthLoading());
    try {
      await _authRepository.login(
        LoginRequest(email: email, password: password),
      );
      await _sessionCubit.establishFromStoredToken();
      final session = _sessionCubit.state;
      if (session is SessionFailure) {
        emit(
          AuthFailure(
            message: session.message,
            exception: session.error is ApiException
                ? session.error as ApiException
                : null,
          ),
        );
        return;
      }
      if (session is SessionUnauthenticated) {
        emit(
          const AuthFailure(
            message: 'Session could not be established after login',
          ),
        );
        return;
      }
      emit(const AuthLoginSuccess());
    } on ApiException catch (e) {
      emit(AuthFailure(message: e.message, exception: e));
    } catch (e) {
      emit(AuthFailure(message: e.toString()));
    }
  }

  void reset() => emit(const AuthInitial());

  @override
  void emit(AuthState state) {
    if (isClosed) return;
    super.emit(state);
  }
}
