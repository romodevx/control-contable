import 'package:control_contable/core/network/api_failure.dart';
import 'package:control_contable/core/session/session.dart';
import 'package:control_contable/features/auth/data/profile_repository.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// Deletes the account and everything it owns (US-A4), confirmed with the
/// current password (D4), then forgets the session on this device.
///
/// A wrong password is a `403 invalid_password` (contract 0.3.0), never a
/// `401`: a `401` is the bearer, which the interceptor refreshes and
/// retries like any other call, so no warm-up request is needed.
class DeleteAccount {
  /// Deletes through [_repository] and ends [_session].
  new(this._repository, this._session);

  final ProfileRepository _repository;
  final Session _session;

  /// Deletes the account. Throws an [ApiFailure] when nothing was deleted
  /// (`ApiInvalidPassword`, `ApiRateLimited`, ...); the session is then
  /// kept.
  Future<void> call(String password) async {
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
