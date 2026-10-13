import 'package:control_contable/api/generated/models/auth_session.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Where a new session is kept once the email is verified.
abstract interface class SessionStore {
  /// Keeps [session] as the signed-in session.
  Future<void> save(AuthSession session);
}

/// Keeps the session in memory only, until M1-12 replaces it with the
/// `flutter_secure_storage` store (mobile/README.md, "Session tokens").
/// Nothing reaches disk, so nothing can leak through a backup.
class InMemorySessionStore implements SessionStore {
  AuthSession? _session;

  /// The session kept, if any.
  AuthSession? get session => _session;

  @override
  Future<void> save(AuthSession session) async => _session = session;
}

/// The session store. Tests override it with a recorder.
final sessionStoreProvider = Provider<SessionStore>(
  (ref) => InMemorySessionStore(),
  name: 'sessionStoreProvider',
);
