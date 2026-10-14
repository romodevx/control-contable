import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Deletes the account and everything it owns (US-A4), confirmed with the
/// current password (D4), then forgets the session on this device.
///
/// `DELETE /v1/me` answers a wrong password with a `401`, the same status
/// as an expired bearer; the interceptor therefore never refreshes nor
/// retries it (a retry would count the wrong password twice). So the bearer
/// is freshened first with a `GET /v1/me`, which the interceptor does
/// refresh for: a `401` from the delete then means the password.
class DeleteAccount {
  /// Deletes through [_repository] and ends [_session].
  new(this._repository, this._session);

  final ProfileRepository _repository;
  final Session _session;

  /// Deletes the account. Throws an [ApiFailure] when nothing was deleted
  /// (`ApiUnauthorized`: wrong password); the session is then kept.
  Future<void> call(String password) async {
    await _repository.getMe();
    await _repository.deleteAccount(password);
    await _session.end();
  }
}

/// The account-deletion use case.
final deleteAccountProvider = Provider<DeleteAccount>(
  (ref) => DeleteAccount(
    ref.watch(profileRepositoryProvider),
    ref.watch(sessionProvider),
  ),
  name: 'deleteAccountProvider',
);
