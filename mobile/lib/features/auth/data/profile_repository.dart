import 'package:control_contable/api/generated/clients/auth_client.dart';
import 'package:control_contable/api/generated/models/delete_account_request.dart';
import 'package:control_contable/api/generated/models/user.dart';
import 'package:control_contable/api/generated/models/user_update.dart';
import 'package:control_contable/core/network/api_client_provider.dart';
import 'package:control_contable/core/network/api_failure.dart';
import 'package:dio/dio.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';

/// The envelope code a wrong current password may come with, whatever the
/// status (the contract answers `401 unauthorized` today; a distinct
/// `invalid_password` may follow, e.g. as a `403`).
const String _invalidPasswordCode = 'invalid_password';

/// The signed-in user's own account (`/v1/me`). Every method throws an
/// `ApiFailure` when the server refuses or cannot be reached.
abstract interface class ProfileRepository {
  /// The profile (`GET /v1/me`).
  Future<User> getMe();

  /// Renames the user (`PATCH /v1/me`); the email cannot change.
  Future<User> updateName(String name);

  /// Deletes the account and all its data (`DELETE /v1/me`), confirmed
  /// with the current [password]. A wrong one is an `ApiUnauthorized`,
  /// whether the server says so with a `401` or with the
  /// `invalid_password` code under any status.
  Future<void> deleteAccount(String password);
}

/// [ProfileRepository] over the generated auth client.
class ApiProfileRepository implements ProfileRepository {
  /// The repository on [_client].
  new(this._client);

  final AuthClient _client;

  @override
  Future<User> getMe() => guardApi(_client.getMe);

  @override
  Future<User> updateName(String name) =>
      guardApi(() => _client.updateMe(body: UserUpdate(name: name)));

  @override
  Future<void> deleteAccount(String password) => guardApi(
    () => _client.deleteMe(body: DeleteAccountRequest(password: password)),
    classify: _wrongPassword,
  );

  static ApiFailure? _wrongPassword(DioException exception) =>
      apiErrorCode(exception.response?.data) == _invalidPasswordCode
      ? const ApiUnauthorized()
      : null;
}

/// The account operations. Tests override it with a fake.
final profileRepositoryProvider = Provider<ProfileRepository>(
  (ref) => ApiProfileRepository(ref.watch(apiClientProvider).auth),
  name: 'profileRepositoryProvider',
);
