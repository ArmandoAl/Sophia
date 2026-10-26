import 'package:sophia_core/core/models/auth/auth_user.dart';
import 'package:sophia_core/core/models/conversations/conversation.dart';
import 'package:sophia_core/core/models/conversations/conversation_message.dart';
import 'package:sophia_core/core/models/enums/autonomy_level.dart';
import 'package:sophia_core/core/models/enums/proactivity_level.dart';
import 'package:sophia_core/core/models/reminders/reminder.dart';
import 'package:sophia_core/core/models/users/ai_settings.dart';
import 'package:sophia_core/core/models/users/me_response.dart';
import 'package:sophia_core/core/models/users/user_profile.dart';
import 'package:sophia_core/features/beliefs/domain/models.dart';

class DemoStore {
  DemoStore._() {
    final now = DateTime.now();
    user = const AuthUser(
      id: 'demo-user',
      name: 'Alex Demo',
      email: 'alex@example.test',
      role: 'user',
    );
    profile = UserProfile(
      id: 'demo-profile',
      userId: user.id,
      displayName: 'Alex Demo',
      preferredName: 'Alex',
      timezone: 'America/Tijuana',
      locale: 'es-MX',
      onboardingCompleted: true,
      createdAt: now,
      updatedAt: now,
    );
    aiSettings = AiSettings(
      userId: user.id,
      assistantName: 'Sofía',
      communicationStyle: 'cercano',
      proactivityLevel: ProactivityLevel.medium,
      autonomyLevel: AutonomyLevel.suggestive,
      memoryEnabled: true,
      remindersEnabled: true,
      planningEnabled: true,
      preferredLanguage: 'es',
      createdAt: now,
      updatedAt: now,
    );
    conversation = Conversation(
      id: 'demo-conversation',
      userId: user.id,
      title: 'Un día con más calma',
      status: 'active',
      createdAt: now,
      updatedAt: now,
      lastMessageAt: now,
    );
    messages = [
      _message(
        id: 'demo-welcome',
        role: 'assistant',
        content: 'Hola, Alex. ¿Qué te gustaría organizar hoy?',
      ),
    ];
    beliefs.addAll([
      const Belief(
        id: 'demo-belief-rhythm',
        statement: 'Te funciona mejor empezar el día sin prisas.',
        category: 'Rutina',
        scope: 'global',
        scopeKey: 'global',
        confidence: .91,
        decayedConfidence: .88,
        evidenceCount: 3,
        contradictionCount: 0,
        trustTier: 2,
        promptSlot: 'daily_rhythm',
      ),
      const Belief(
        id: 'demo-belief-walk',
        statement: 'Una caminata corta suele ayudarte a despejarte.',
        category: 'Bienestar',
        scope: 'global',
        scopeKey: 'global',
        confidence: .78,
        decayedConfidence: .74,
        evidenceCount: 2,
        contradictionCount: 0,
        trustTier: 1,
        promptSlot: 'wellbeing',
      ),
      const Belief(
        id: 'demo-belief-rest',
        statement: 'Prefieres dejar un espacio libre al final de la tarde.',
        category: 'Preferencias',
        scope: 'global',
        scopeKey: 'global',
        confidence: .69,
        decayedConfidence: .65,
        evidenceCount: 2,
        contradictionCount: 1,
        trustTier: 1,
        promptSlot: 'schedule',
      ),
    ]);
    reminders.add(
      _reminder(
        id: 'demo-reminder-water',
        title: 'Tomar una pausa',
        description: 'Un momento para estirarte y tomar agua.',
        at: DateTime(now.year, now.month, now.day, now.hour + 1),
      ),
    );
  }

  static final instance = DemoStore._();

  late final AuthUser user;
  late UserProfile profile;
  late AiSettings aiSettings;
  late final Conversation conversation;
  late final List<ConversationMessage> messages;
  final List<Belief> beliefs = [];
  final List<Reminder> reminders = [];
  final Map<String, Map<String, dynamic>> proposals = {};
  int conversationStep = 0;
  int rejectionCount = 0;
  bool retryPending = false;
  String selectedTime = '17:30';
  String selectedDuration = '20 minutos';
  String selectedActivityTitle = 'Paseo corto';

  MeResponse get me =>
      MeResponse(user: user, profile: profile, aiSettings: aiSettings);

  List<String> get starterReplies => const [
    'Quiero moverme un poco',
    'Necesito una pausa tranquila',
    'Ayúdame a ordenar mi tarde',
  ];

  List<String> get retryReplies => rejectionCount >= 3
      ? const ['Algo más breve', 'Probar mañana', 'Cambiar la idea']
      : const [
          'Cambiar la hora',
          'Elegir otra actividad',
          'Dejarlo para mañana',
        ];

  ConversationMessage _message({
    required String id,
    required String role,
    required String content,
  }) => ConversationMessage(
    id: id,
    conversationId: conversation.id,
    userId: user.id,
    role: role,
    content: content,
    createdAt: DateTime.now(),
  );

  ConversationMessage userMessage(String content) => _message(
    id: 'demo-user-${messages.length}',
    role: 'user',
    content: content,
  );

  ConversationMessage assistantMessage(String content, {String? id}) =>
      _message(
        id: id ?? 'demo-assistant-${messages.length}',
        role: 'assistant',
        content: content,
      );

  Reminder addActivity({
    required String id,
    required String title,
    required String description,
    required DateTime at,
  }) {
    final reminder = _reminder(
      id: id,
      title: title,
      description: description,
      at: at,
    );
    reminders.insert(0, reminder);
    return reminder;
  }

  Reminder _reminder({
    required String id,
    required String title,
    required String description,
    required DateTime at,
  }) {
    final now = DateTime.now();
    return Reminder(
      id: id,
      userId: user.id,
      title: title,
      description: description,
      status: ReminderStatus.pending,
      scheduledAt: at,
      timezone: 'America/Tijuana',
      recurrenceRule: RecurrenceRule.none,
      nextRunAt: at,
      failureReason: '',
      retryCount: 0,
      maxRetries: 0,
      source: ReminderSource.aiSuggested,
      createdBy: 'demo',
      createdAt: now,
      updatedAt: now,
    );
  }
}
