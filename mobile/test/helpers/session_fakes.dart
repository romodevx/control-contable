import 'package:control_contable/api/generated/models/token_pair.dart';
import 'package:control_contable/core/session/refresh_token_store.dart';
import 'package:flutter/services.dart';

/// A token pair as `login`, `verifyEmail` or `refreshSession` answer it.
TokenPair tokenPair({required String access, required String refresh}) =>
    TokenPair(
      accessToken: access,
      accessExpiresIn: 900,
      refreshToken: refresh,
      refreshExpiresIn: 2592000,
      tokenType: 'Bearer',
    );

/// A [RefreshTokenStore] held in memory, so tests can read what the app
/// kept. A [readError] makes every read fail with it; [failingWrites] and
/// [failingDeletes] make that many of the next calls fail (a locked
/// Keychain, a Keystore hiccup); a [wipeError] makes [wipeAll] fail.
class InMemoryRefreshTokenStore implements RefreshTokenStore {
  /// A store that starts with [token] (none by default).
  new([this.token]);

  /// The token kept, if any.
  String? token;

  /// Every token written, in order (failed attempts included).
  final List<String> written = [];

  /// Thrown by [read] when set.
  Exception? readError;

  /// How many of the next writes fail.
  int failingWrites = 0;

  /// How many of the next deletes fail.
  int failingDeletes = 0;

  /// Thrown by [wipeAll] when set.
  Exception? wipeError;

  /// How many times the whole namespace was wiped.
  int wipes = 0;

  @override
  Future<String?> read() async {
    if (readError case final error?) throw error;
    return token;
  }

  @override
  Future<void> write(String refreshToken) async {
    written.add(refreshToken);
    if (failingWrites > 0) {
      failingWrites--;
      // A platform error that, like some real ones, echoes the value.
      throw PlatformException(code: 'write', message: refreshToken);
    }
    token = refreshToken;
  }

  @override
  Future<void> delete() async {
    if (failingDeletes > 0) {
      failingDeletes--;
      throw PlatformException(code: 'delete');
    }
    token = null;
  }

  @override
  Future<void> wipeAll() async {
    if (wipeError case final error?) throw error;
    wipes++;
    token = null;
  }
}
