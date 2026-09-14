import 'package:flutter_bloc/flutter_bloc.dart';

import '../../../../core/models/users/update_profile_request.dart';
import '../../../../core/models/users/user_profile.dart';
import '../../../../core/network/api_exception.dart';
import '../../../session/presentation/cubit/session_cubit.dart';
import '../../../session/presentation/cubit/session_state.dart';
import '../../domain/user_repository.dart';
import 'user_profile_state.dart';

class UserProfileCubit extends Cubit<UserProfileState> {
  UserProfileCubit({
    required UserRepository userRepository,
    required SessionCubit sessionCubit,
  }) : _userRepository = userRepository,
       _sessionCubit = sessionCubit,
       super(const UserProfileInitial());

  final UserRepository _userRepository;
  final SessionCubit _sessionCubit;

  /// Prefer session snapshot; optionally refresh via getMe.
  void loadFromSession() {
    final session = _sessionCubit.state;
    if (session is SessionAuthenticated) {
      emit(UserProfileLoaded(session.profile));
    }
  }

  Future<void> refresh() async {
    emit(const UserProfileLoading());
    try {
      final me = await _userRepository.getMe();
      _sessionCubit.applyMe(me);
      emit(UserProfileLoaded(me.profile));
    } on ApiException catch (e) {
      if (e.isUnauthorized) {
        await _sessionCubit.logout();
      }
      emit(UserProfileFailure(message: e.message, exception: e));
    } catch (e) {
      emit(UserProfileFailure(message: e.toString()));
    }
  }

  Future<void> update(UpdateProfileRequest request) async {
    final previous = _currentProfile;
    if (previous != null) {
      emit(UserProfileLoaded(previous, saving: true));
    } else {
      emit(const UserProfileLoading());
    }

    try {
      final updated = await _userRepository.updateProfile(request);
      _sessionCubit.applyProfile(updated);
      emit(UserProfileLoaded(updated));
    } on ApiException catch (e) {
      if (e.isUnauthorized) {
        await _sessionCubit.logout();
      }
      emit(
        UserProfileFailure(
          message: e.message,
          exception: e,
          previous: previous,
        ),
      );
    } catch (e) {
      emit(UserProfileFailure(message: e.toString(), previous: previous));
    }
  }

  UserProfile? get _currentProfile {
    final s = state;
    if (s is UserProfileLoaded) return s.profile;
    if (s is UserProfileFailure) return s.previous;
    final session = _sessionCubit.state;
    if (session is SessionAuthenticated) return session.profile;
    return null;
  }

  @override
  void emit(UserProfileState state) {
    if (isClosed) return;
    super.emit(state);
  }
}
