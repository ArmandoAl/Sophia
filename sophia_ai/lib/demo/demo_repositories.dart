import 'package:sophia_ai/core/models/models.dart';
import 'package:sophia_ai/core/di/service_locator.dart';
import 'package:sophia_ai/features/actions/domain/action_proposals_repository.dart';
import 'package:sophia_ai/features/auth/domain/auth_repository.dart';
import 'package:sophia_ai/features/beliefs/domain/beliefs_repository.dart';
import 'package:sophia_ai/features/beliefs/domain/models.dart';
import 'package:sophia_ai/features/chat/domain/conversations_repository.dart';
import 'package:sophia_ai/features/contexts/domain/contexts_repository.dart';
import 'package:sophia_ai/features/contexts/domain/models.dart';
import 'package:sophia_ai/features/reminders/domain/reminders_repository.dart';
import 'package:sophia_ai/features/settings/domain/models.dart';
import 'package:sophia_ai/features/settings/domain/settings_services_repository.dart';
import 'package:sophia_ai/features/system/domain/system_repository.dart';
import 'package:sophia_ai/features/users/domain/user_repository.dart';

import 'demo_store.dart';

void registerDemoRepositories() {
  final store = DemoStore.instance;
  sl.registerSingleton<ConversationsRepository>(
    DemoConversationsRepository(store),
  );
  sl.registerSingleton<ActionProposalsRepository>(
    DemoActionProposalsRepository(store),
  );
  sl.registerSingleton<RemindersRepository>(DemoRemindersRepository(store));
  sl.registerSingleton<BeliefsRepository>(DemoBeliefsRepository(store));
  sl.registerSingleton<ContextsRepository>(DemoContextsRepository());
  sl.registerSingleton<UserRepository>(DemoUserRepository(store));
  sl.registerSingleton<AuthRepository>(DemoAuthRepository(store));
  sl.registerSingleton<SystemRepository>(DemoSystemRepository());
  sl.registerSingleton<SettingsServicesRepository>(DemoSettingsRepository());
}

class DemoConversationsRepository implements ConversationsRepository {
  DemoConversationsRepository(this.store);
  final DemoStore store;

  @override
  Future<Conversation> createConversation(
    CreateConversationRequest request,
  ) async => store.conversation;

  @override
  Future<ConversationListResponse> listConversations({
    String? status,
    int? limit,
    String? cursor,
  }) async => ConversationListResponse(
    conversations: [store.conversation],
    suggestedReplies: store.starterReplies,
  );

  @override
  Future<Conversation> getConversation(String id) async => store.conversation;

  @override
  Future<ConversationMessagesResponse> listMessages(
    String conversationId, {
    int? limit,
    String? cursor,
  }) async => ConversationMessagesResponse(messages: store.messages);

  @override
  Future<Conversation> archiveConversation(String id) async =>
      store.conversation;

  @override
  Future<SendConversationMessageResponse> sendMessage(
    String conversationId,
    SendConversationMessageRequest request,
  ) async {
    await Future<void>.delayed(const Duration(milliseconds: 500));
    final userMessage = store.userMessage(request.content);
    List<RuntimeProposedAction> actions;
    final String response;
    final List<String> replies;

    if (store.retryPending) {
      store.retryPending = false;
      final id = 'demo-proposal-${store.proposals.length + 1}';
      final title = _retryTitle(request.content);
      final input = _proposalInput(title, _retryTime(request.content));
      if (request.content.contains('mañana')) {
        input['scheduled_at'] = DateTime.parse(
          input['scheduled_at'] as String,
        ).add(const Duration(days: 1)).toIso8601String();
      }
      store.proposals[id] = input;
      actions = [
        RuntimeProposedAction(
          id: id,
          toolName: title,
          proposedInput: input,
          reason: 'Ajuste local de la propuesta anterior.',
          riskLevel: 'low',
          requiresConfirmation: true,
        ),
      ];
      response = 'Perfecto, probemos con esta alternativa.';
      replies = const [];
    } else {
      actions = const [];
      switch (store.conversationStep++) {
        case 0:
          store.selectedActivityTitle = _activityTitle(request.content);
          response = '¿En qué momento del día te queda mejor?';
          replies = const ['Por la mañana', 'Después de comer', 'Por la tarde'];
        case 1:
          store.selectedTime = _timeFor(request.content);
          response = '¿Cuánto tiempo te gustaría reservar?';
          replies = const ['10 minutos', '20 minutos', '30 minutos'];
        default:
          store.selectedDuration = request.content;
          final title = store.selectedActivityTitle;
          final id = 'demo-proposal-${store.proposals.length + 1}';
          final input = _proposalInput(title, store.selectedTime);
          input['description'] = 'Duración: ${store.selectedDuration}';
          store.proposals[id] = input;
          actions = [
            RuntimeProposedAction(
              id: id,
              toolName: title,
              proposedInput: input,
              reason: 'Propuesta de ejemplo, pendiente de tu decisión.',
              riskLevel: 'low',
              requiresConfirmation: true,
            ),
          ];
          response = 'Con lo que me contaste, te propongo esto:';
          replies = const [];
      }
    }

    final assistantMessage = store.assistantMessage(response);
    store.messages
      ..add(userMessage)
      ..add(assistantMessage);
    return SendConversationMessageResponse(
      conversation: store.conversation,
      userMessage: userMessage,
      assistantMessage: assistantMessage,
      proposedActions: actions,
      suggestedReplies: replies,
    );
  }

  Map<String, dynamic> _proposalInput(String title, String time) => {
    'title': title,
    'description': 'Actividad de ejemplo, solo se agrega a esta demo.',
    'scheduled_at': _dateAt(time).toIso8601String(),
    'timezone': 'America/Tijuana',
  };

  String _timeFor(String input) => switch (input) {
    'Por la mañana' => '09:00',
    'Después de comer' => '14:30',
    _ => '17:30',
  };

  String _activityTitle(String firstMessage) => firstMessage.contains('pausa')
      ? 'Pausa tranquila'
      : firstMessage.contains('ordenar')
      ? 'Organizar la tarde'
      : 'Paseo corto';

  String _retryTitle(String input) => input.contains('mañana')
      ? 'Paseo de mañana'
      : input.contains('pausa')
      ? 'Pausa breve'
      : 'Paseo corto';

  String _retryTime(String input) => input.contains('mañana')
      ? '10:00'
      : input.contains('breve')
      ? '17:00'
      : input.contains('actividad')
      ? store.selectedTime
      : '18:30';

  DateTime _dateAt(String time) {
    final parts = time.split(':');
    final now = DateTime.now();
    return DateTime(
      now.year,
      now.month,
      now.day,
      int.parse(parts[0]),
      int.parse(parts[1]),
    );
  }
}

class DemoActionProposalsRepository implements ActionProposalsRepository {
  DemoActionProposalsRepository(this.store);
  final DemoStore store;

  @override
  Future<void> confirm(
    String proposalId, {
    Map<String, dynamic>? correctedInput,
    int? decisionLatencyMs,
  }) async {
    if (correctedInput != null) {
      store.proposals[proposalId] = {
        ...?store.proposals[proposalId],
        ...correctedInput,
      };
    }
  }

  @override
  Future<void> reject(
    String proposalId, {
    String? rejectionReason,
    int? decisionLatencyMs,
  }) async {
    store.rejectionCount++;
    store.retryPending = true;
  }

  @override
  Future<void> execute(String proposalId) async {
    final input = store.proposals[proposalId] ?? const <String, dynamic>{};
    final at =
        DateTime.tryParse(input['scheduled_at'] as String? ?? '') ??
        DateTime.now().add(const Duration(hours: 1));
    store.addActivity(
      id: 'demo-activity-${store.reminders.length + 1}',
      title: input['title'] as String? ?? 'Actividad de ejemplo',
      description: input['description'] as String? ?? 'Creada en modo demo.',
      at: at,
    );
  }
}

class DemoRemindersRepository implements RemindersRepository {
  DemoRemindersRepository(this.store);
  final DemoStore store;

  @override
  Future<ReminderListResponse> list({
    int? limit,
    String? cursor,
    String? status,
    String? activityId,
    DateTime? from,
    DateTime? to,
  }) async => ReminderListResponse(
    reminders: store.reminders
        .where((item) {
          if (status != null && status != item.status.name) return false;
          if (activityId != null && activityId != item.activityId) return false;
          if (from != null && item.nextRunAt.isBefore(from)) return false;
          if (to != null && item.nextRunAt.isAfter(to)) return false;
          return true;
        })
        .take(limit ?? store.reminders.length)
        .toList(),
  );

  @override
  Future<ReminderListResponse> listDue({int? limit}) => list(limit: limit);

  @override
  Future<Reminder> create(CreateReminderRequest request) async =>
      store.addActivity(
        id: 'demo-activity-${store.reminders.length + 1}',
        title: request.title,
        description: request.description ?? '',
        at: DateTime.tryParse(request.scheduledAt) ?? DateTime.now(),
      );

  @override
  Future<Reminder> getById(String id) async =>
      store.reminders.firstWhere((r) => r.id == id);

  @override
  Future<Reminder> update(String id, UpdateReminderRequest request) async {
    final old = await getById(id);
    final now = DateTime.now();
    final updated = Reminder(
      id: old.id,
      userId: old.userId,
      activityId: old.activityId,
      title: request.title ?? old.title,
      description: request.description ?? old.description,
      status: old.status,
      scheduledAt:
          DateTime.tryParse(request.scheduledAt ?? '') ?? old.scheduledAt,
      timezone: request.timezone ?? old.timezone,
      recurrenceRule: request.recurrenceRule == null
          ? old.recurrenceRule
          : parseRecurrenceRule(request.recurrenceRule!),
      nextRunAt: DateTime.tryParse(request.scheduledAt ?? '') ?? old.nextRunAt,
      lastRunAt: old.lastRunAt,
      deliveredAt: old.deliveredAt,
      failedAt: old.failedAt,
      failureReason: old.failureReason,
      retryCount: old.retryCount,
      maxRetries: request.maxRetries ?? old.maxRetries,
      source: old.source,
      createdBy: old.createdBy,
      createdAt: old.createdAt,
      updatedAt: now,
    );
    store.reminders[store.reminders.indexOf(old)] = updated;
    return updated;
  }

  @override
  Future<Reminder> cancel(String id) => _status(id, ReminderStatus.canceled);

  @override
  Future<Reminder> archive(String id) => _status(id, ReminderStatus.archived);

  Future<Reminder> _status(String id, ReminderStatus status) async {
    final old = await getById(id);
    final updated = Reminder(
      id: old.id,
      userId: old.userId,
      title: old.title,
      description: old.description,
      status: status,
      scheduledAt: old.scheduledAt,
      timezone: old.timezone,
      recurrenceRule: old.recurrenceRule,
      nextRunAt: old.nextRunAt,
      failureReason: old.failureReason,
      retryCount: old.retryCount,
      maxRetries: old.maxRetries,
      source: old.source,
      createdBy: old.createdBy,
      createdAt: old.createdAt,
      updatedAt: DateTime.now(),
    );
    store.reminders[store.reminders.indexOf(old)] = updated;
    return updated;
  }
}

class DemoBeliefsRepository implements BeliefsRepository {
  DemoBeliefsRepository(this.store);
  final DemoStore store;

  @override
  Future<List<Belief>> list({String? scope, String? scopeKey}) async =>
      store.beliefs.where((item) => item.status == 'active').toList();

  @override
  Future<void> retire(String id) async {
    final index = store.beliefs.indexWhere((item) => item.id == id);
    if (index >= 0) {
      store.beliefs[index] = _copyBelief(
        store.beliefs[index],
        status: 'retired',
      );
    }
  }

  @override
  Future<Belief> updateStatement(String id, String statement) async {
    final index = store.beliefs.indexWhere((item) => item.id == id);
    if (index < 0) throw StateError('Creencia de ejemplo no encontrada.');
    return store.beliefs[index] = _copyBelief(
      store.beliefs[index],
      statement: statement,
    );
  }

  @override
  Future<PromptVersion?> getPromptVersion() async => const PromptVersion(
    version: 1,
    content: 'Prompt de ejemplo local.',
    tokenCount: 12,
  );

  @override
  Future<List<LearningSummary>> listSummaries({int limit = 7}) async =>
      const [];

  Belief _copyBelief(Belief item, {String? statement, String? status}) =>
      Belief(
        id: item.id,
        statement: statement ?? item.statement,
        category: item.category,
        scope: item.scope,
        scopeKey: item.scopeKey,
        confidence: item.confidence,
        decayedConfidence: item.decayedConfidence,
        evidenceCount: item.evidenceCount,
        contradictionCount: item.contradictionCount,
        trustTier: item.trustTier,
        promptSlot: item.promptSlot,
        status: status ?? item.status,
        subjectType: item.subjectType,
        subjectId: item.subjectId,
        factKind: item.factKind,
        sensitive: item.sensitive,
      );
}

class DemoContextsRepository implements ContextsRepository {
  final List<UserContext> _contexts = [
    const UserContext(
      id: 'demo-person-friend',
      kind: 'person',
      slug: 'friend',
      label: 'Una amistad',
      aliases: ['amistad'],
      active: true,
      relationship: 'Amistad',
    ),
  ];

  @override
  Future<List<UserContext>> list() async => List.of(_contexts);

  @override
  Future<UserContext> create({
    required String kind,
    required String slug,
    required String label,
    required List<String> aliases,
  }) async {
    final item = UserContext(
      id: 'demo-$slug',
      kind: kind,
      slug: slug,
      label: label,
      aliases: aliases,
      active: true,
    );
    _contexts.add(item);
    return item;
  }

  @override
  Future<UserContext> update(
    String id, {
    required String label,
    required List<String> aliases,
  }) async => _replace(id, label: label, aliases: aliases);

  @override
  Future<void> archive(String id) async => _remove(id);

  @override
  Future<List<UserContext>> listEntities({
    String? kind,
    String? status,
  }) async => _contexts.where((e) => kind == null || e.kind == kind).toList();

  @override
  Future<UserContext> createEntity({
    required String kind,
    required String slug,
    required String label,
    required String relationship,
    required List<String> aliases,
  }) async {
    final item = UserContext(
      id: 'demo-$slug',
      kind: kind,
      slug: slug,
      label: label,
      aliases: aliases,
      active: true,
      relationship: relationship,
    );
    _contexts.add(item);
    return item;
  }

  @override
  Future<EntityDetails> getEntity(String id) async => EntityDetails(
    entity: _contexts.firstWhere((e) => e.id == id),
    facts: const [],
    userBeliefs: const [],
    episodes: const [],
  );

  @override
  Future<UserContext> updateEntity(
    String id,
    Map<String, dynamic> changes,
  ) async => _replace(id, label: changes['label'] as String?);

  @override
  Future<UserContext> mergeEntity(String id, String into) async =>
      _replace(id, status: 'archived', active: false);

  @override
  Future<void> archiveEntity(String id) async => _remove(id);

  UserContext _replace(
    String id, {
    String? label,
    List<String>? aliases,
    String? status,
    bool? active,
  }) {
    final index = _contexts.indexWhere((e) => e.id == id);
    final old = _contexts[index];
    return _contexts[index] = UserContext(
      id: old.id,
      kind: old.kind,
      slug: old.slug,
      label: label ?? old.label,
      aliases: aliases ?? old.aliases,
      active: active ?? old.active,
      relationship: old.relationship,
      status: status ?? old.status,
    );
  }

  void _remove(String id) => _contexts.removeWhere((e) => e.id == id);
}

class DemoUserRepository implements UserRepository {
  DemoUserRepository(this.store);
  final DemoStore store;
  @override
  Future<MeResponse> getMe() async => store.me;
  @override
  Future<UserProfile> updateProfile(UpdateProfileRequest request) async {
    final old = store.profile;
    store.profile = UserProfile(
      id: old.id,
      userId: old.userId,
      displayName: request.displayName ?? old.displayName,
      preferredName: request.preferredName ?? old.preferredName,
      timezone: request.timezone ?? old.timezone,
      locale: request.locale ?? old.locale,
      avatarUrl: request.avatarUrl ?? old.avatarUrl,
      onboardingCompleted: old.onboardingCompleted,
      createdAt: old.createdAt,
      updatedAt: DateTime.now(),
    );
    return store.profile;
  }

  @override
  Future<AiSettings> getAiSettings() async => store.aiSettings;
  @override
  Future<AiSettings> updateAiSettings(UpdateAiSettingsRequest request) async {
    final old = store.aiSettings;
    store.aiSettings = AiSettings(
      userId: old.userId,
      assistantName: request.assistantName ?? old.assistantName,
      communicationStyle: request.communicationStyle ?? old.communicationStyle,
      proactivityLevel: request.proactivityLevel ?? old.proactivityLevel,
      autonomyLevel: request.autonomyLevel ?? old.autonomyLevel,
      memoryEnabled: request.memoryEnabled ?? old.memoryEnabled,
      remindersEnabled: request.remindersEnabled ?? old.remindersEnabled,
      planningEnabled: request.planningEnabled ?? old.planningEnabled,
      preferredLanguage: request.preferredLanguage ?? old.preferredLanguage,
      createdAt: old.createdAt,
      updatedAt: DateTime.now(),
    );
    return store.aiSettings;
  }

  @override
  Future<UserProfile> completeOnboarding() async =>
      updateProfile(const UpdateProfileRequest());
}

class DemoAuthRepository implements AuthRepository {
  DemoAuthRepository(this.store);
  final DemoStore store;
  @override
  Future<AuthUser> register(RegisterRequest request) async => store.user;
  @override
  Future<LoginResponse> login(LoginRequest request) async =>
      LoginResponse(token: 'demo-token', user: store.user);
  @override
  Future<void> logout() async {}
}

class DemoSystemRepository implements SystemRepository {
  @override
  Future<HealthResponse> getHealth() async => const HealthResponse(
    status: 'ok',
    environment: 'demo',
    firestore: 'simulado',
  );
}

class DemoSettingsRepository implements SettingsServicesRepository {
  @override
  Future<Map<String, dynamic>> exportData() async => {
    'demo': true,
    'message': 'Datos ficticios de demostración',
  };
  @override
  Future<void> requestDeletion() async {}
  @override
  Future<List<DeviceToken>> listDeviceTokens() async => const [];
  @override
  Future<DeviceToken> registerDevice({
    required String platform,
    required String token,
  }) async => DeviceToken(
    id: 'demo-device',
    platform: platform,
    createdAt: DateTime.now(),
    tokenRedacted: 'demo',
  );
  @override
  Future<void> removeDevice(String id) async {}
  @override
  Future<List<IngestionBatch>> listIngestionBatches() async => const [];
  @override
  Future<void> undoIngestionBatch(String id) async {}
}
