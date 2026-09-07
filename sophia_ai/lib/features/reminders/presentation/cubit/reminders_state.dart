import 'package:equatable/equatable.dart';
import '../../../../core/models/models.dart';

class RemindersState extends Equatable {
  const RemindersState({
    this.reminders = const [],
    this.isLoading = false,
    this.isSubmitting = false,
    this.errorMessage,
  });
  final List<Reminder> reminders;
  final bool isLoading, isSubmitting;
  final String? errorMessage;
  RemindersState copyWith({
    List<Reminder>? reminders,
    bool? isLoading,
    bool? isSubmitting,
    String? errorMessage,
    bool clearError = false,
  }) => RemindersState(
    reminders: reminders ?? this.reminders,
    isLoading: isLoading ?? this.isLoading,
    isSubmitting: isSubmitting ?? this.isSubmitting,
    errorMessage: clearError ? null : errorMessage ?? this.errorMessage,
  );
  @override
  List<Object?> get props => [reminders, isLoading, isSubmitting, errorMessage];
}
